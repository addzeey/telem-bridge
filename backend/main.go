package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
)

//go:embed dist
var content embed.FS

// Version of the application, set at build time via -ldflags
var Version = "0.0.21"

func spaHandler(distFS fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws") {
			http.NotFound(w, r)
			return
		}
		// Try to serve the file if it exists
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		f, err := distFS.Open(path)
		if err == nil {
			defer f.Close()
			info, err := f.Stat()
			if err == nil && !info.IsDir() {
				data, _ := io.ReadAll(f)
				http.ServeContent(w, r, path, info.ModTime(), bytes.NewReader(data))
				return
			}
		}
		// Fallback: serve index.html for client-side routing
		index, err := distFS.Open("index.html")
		if err != nil {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
			return
		}
		defer index.Close()
		w.Header().Set("Content-Type", "text/html")
		content, _ := io.ReadAll(index)
		http.ServeContent(w, r, "index.html", time.Now(), bytes.NewReader(content))
	}
}

func setupLogging() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Could not get user config dir: %v", err)
		return
	}
	appDir := filepath.Join(configDir, "f1-telem-bridge")
	os.MkdirAll(appDir, 0755)
	logPath := filepath.Join(appDir, "app.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("Could not open log file: %v", err)
		return
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}

// Path for packet forwarding config
var packetForwardingConfigPath string

func InitPacketForwardingConfig() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	appDir := filepath.Join(configDir, "f1-telem-bridge")
	os.MkdirAll(appDir, 0755)
	packetForwardingConfigPath = filepath.Join(appDir, "packet_forwarding.json")

	if _, err := os.Stat(packetForwardingConfigPath); os.IsNotExist(err) {
		SavePacketForwardingConfig()
	} else {
		LoadPacketForwardingConfig()
	}
}

func SavePacketForwardingConfig() {
	f, err := os.Create(packetForwardingConfigPath)
	if err != nil {
		log.Printf("[error] Could not create packet forwarding config file: %v", err)
		return
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(PacketForwardingConfig); err != nil {
		log.Printf("[error] Could not encode packet forwarding config: %v", err)
	}
}

func LoadPacketForwardingConfig() {
	f, err := os.Open(packetForwardingConfigPath)
	if err != nil {
		log.Printf("[error] Could not open packet forwarding config file: %v", err)
		return
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&PacketForwardingConfig); err != nil {
		log.Printf("[error] Could not decode packet forwarding config: %v", err)
	}
}

// REST API for getting/setting enabled packet types
func handlePacketForwardingAPI(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[panic] PacketForwarding API handler crashed: %v", r)
		}
	}()
	if r.Method == http.MethodGet {
		json.NewEncoder(w).Encode(PacketForwardingConfig)
	} else if r.Method == http.MethodPost {
		var update map[uint8]bool
		json.NewDecoder(r.Body).Decode(&update)
		for k, v := range update {
			PacketForwardingConfig[k] = v
		}
		SavePacketForwardingConfig()
		w.WriteHeader(http.StatusOK)
	}
}

// REST API for version
func handleVersionAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": Version})
}

var udpListenerStop chan struct{}
var udpListenerConn *net.UDPConn

func restartUDPListener() {
	if udpListenerStop != nil {
		close(udpListenerStop)
	}
	if udpListenerConn != nil {
		udpListenerConn.Close()
		udpListenerConn = nil
	}
	udpListenerStop = make(chan struct{})
	go func(stopCh chan struct{}) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[panic] UDP listener crashed: %v", r)
			}
		}()
		addr := net.UDPAddr{
			IP:   net.ParseIP(Config.UDPAddr),
			Port: Config.UDPPort,
		}
		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			log.Printf("Failed to listen on UDP: %v", err)
			return
		}
		udpListenerConn = conn
		defer func() {
			conn.Close()
			udpListenerConn = nil
		}()
		log.Printf("Listening for F1 UDP on %s:%d\n", Config.UDPAddr, Config.UDPPort)
		buf := make([]byte, 2048)
		for {
			select {
			case <-stopCh:
				return
			default:
				n, _, err := conn.ReadFromUDP(buf)
				if err != nil {
					log.Println("UDP read error:", err)
					continue
				}
				handleUDPPacket(buf[:n])
			}
		}
	}(udpListenerStop)
}

// setCORS sets CORS headers in dev mode
func setCORS(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("DEV") == "1" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func main() {
	setupLogging()
	log.Printf("F1 Telemetry Bridge version: %s", Version)
	InitConfig()
	InitTelemetryFieldsConfig()
	InitPacketForwardingConfig()
	InitOSCAddressesConfig()

	distFS, _ := fs.Sub(content, "dist")

	// API and WebSocket endpoints
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/config", handleConfigAPI)
	http.HandleFunc("/api/fields", handleTelemetryFieldsAPI)
	http.HandleFunc("/api/packet-forwarding", handlePacketForwardingAPI)
	http.HandleFunc("/api/osc-addresses", handleOSCAddressesAPI)
	// Service restart endpoints
	http.HandleFunc("/api/restart/osc", handleRestartOSC)
	http.HandleFunc("/api/restart/udp", handleRestartUDP)
	http.HandleFunc("/api/restart/all", handleRestartAll)
	// Version endpoint
	http.HandleFunc("/api/version", handleVersionAPI)

	// Serve static files and SPA fallback
	http.HandleFunc("/", spaHandler(distFS))

	// Start UDP listener with restart support
	restartUDPListener()
	restartOSCService()

	// Open browser to dashboard
	go func() {
		time.Sleep(500 * time.Millisecond)
		url := "http://localhost:1337"
		var err error
		if os.Getenv("OS") == "Windows_NT" {
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		} else if _, errLook := exec.LookPath("xdg-open"); errLook == nil {
			err = exec.Command("xdg-open", url).Start()
		} else if _, errLook := exec.LookPath("open"); errLook == nil {
			err = exec.Command("open", url).Start()
		}
		if err != nil {
			log.Printf("[warn] Could not open browser: %v", err)
		}
	}()

	// --- Graceful shutdown logic ---
	server := &http.Server{Addr: ":1337"}
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt)
	go func() {
		log.Println("Server running on http://localhost:1337")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-shutdownCh
	log.Println("[shutdown] shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[shutdown] HTTP server shutdown error: %v", err)
	}
	if udpListenerConn != nil {
		udpListenerConn.Close()
	}
	if udpListenerStop != nil {
		close(udpListenerStop)
	}
	log.Println("[shutdown] Cleanup complete. Exiting.")
}

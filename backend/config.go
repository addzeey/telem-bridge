package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/hypebeast/go-osc/osc"
)

type AppConfig struct {
	UDPAddr         string `json:"udp_addr"`
	UDPPort         int    `json:"udp_port"`
	OSCAddr         string `json:"osc_addr"`
	OSCPort         int    `json:"osc_port"`
	EnableOSC       bool   `json:"enable_osc"`
	BroadcastRateHz int    `json:"broadcast_rate_hz"`
	DebugOutput     bool   `json:"debug_output"`
}

var Config AppConfig
var configPath string

func InitConfig() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	appDir := filepath.Join(configDir, "f1-telem-bridge")
	os.MkdirAll(appDir, 0755)
	configPath = filepath.Join(appDir, "config.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Initial setup - ask or use defaults
		Config = AppConfig{
			UDPAddr:         "127.0.0.1",
			UDPPort:         20777,
			OSCAddr:         "127.0.0.1",
			OSCPort:         9000,
			EnableOSC:       false,
			BroadcastRateHz: 2,     // Default to 2Hz
			DebugOutput:     false, // Default to no debug output
		}
		SaveConfig()
	} else {
		LoadConfig()
	}
}

func SaveConfig() {
	f, err := os.Create(configPath)
	if err != nil {
		log.Printf("[error] Could not create config file: %v", err)
		return
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(Config); err != nil {
		log.Printf("[error] Could not encode config: %v", err)
	}
}

func LoadConfig() {
	f, err := os.Open(configPath)
	if err != nil {
		log.Printf("[error] Could not open config file: %v", err)
		return
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&Config); err != nil {
		log.Printf("[error] Could not decode config: %v", err)
	}
}

var oscClientMu sync.Mutex
var oscClient *osc.Client

func restartOSCService() {
	oscClientMu.Lock()
	defer oscClientMu.Unlock()
	oscClient = osc.NewClient(Config.OSCAddr, Config.OSCPort)
	log.Printf("[service] OSC restarted at %s:%d", Config.OSCAddr, Config.OSCPort)
}

func handleConfigAPI(w http.ResponseWriter, r *http.Request) {
	setCORS(w, r)
	if os.Getenv("DEV") == "1" && r.Method == http.MethodOptions {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[panic] Config API handler crashed: %v", r)
		}
	}()

	if r.Method == http.MethodGet {
		json.NewEncoder(w).Encode(Config)
	} else if r.Method == http.MethodPost {
		var oldConfig = Config
		json.NewDecoder(r.Body).Decode(&Config)
		SaveConfig()

		if oldConfig.UDPAddr != Config.UDPAddr || oldConfig.UDPPort != Config.UDPPort {
			log.Printf("[config] UDP address/port changed: %s:%d -> %s:%d", oldConfig.UDPAddr, oldConfig.UDPPort, Config.UDPAddr, Config.UDPPort)
		}
		if oldConfig.OSCAddr != Config.OSCAddr || oldConfig.OSCPort != Config.OSCPort {
			log.Printf("[config] OSC address/port changed: %s:%d -> %s:%d", oldConfig.OSCAddr, oldConfig.OSCPort, Config.OSCAddr, Config.OSCPort)
			restartOSCService()
		}
		if oldConfig.EnableOSC != Config.EnableOSC {
			log.Printf("[config] EnableOSC changed: %v -> %v", oldConfig.EnableOSC, Config.EnableOSC)
		}

		log.Println("[service] UDP restart")
		restartUDPListener()
		w.WriteHeader(http.StatusOK)
	}
}

// Service restart endpoints
func handleRestartOSC(w http.ResponseWriter, r *http.Request) {
	setCORS(w, r)
	if os.Getenv("DEV") == "1" && r.Method == http.MethodOptions {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[panic] RestartOSC handler crashed: %v", r)
		}
	}()
	restartOSCService()
	w.WriteHeader(http.StatusOK)
}

func handleRestartUDP(w http.ResponseWriter, r *http.Request) {
	setCORS(w, r)
	if os.Getenv("DEV") == "1" && r.Method == http.MethodOptions {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[panic] RestartUDP handler crashed: %v", r)
		}
	}()
	log.Println("[service] UDP restart (manual)")
	restartUDPListener()
	w.WriteHeader(http.StatusOK)
}

func handleRestartAll(w http.ResponseWriter, r *http.Request) {
	setCORS(w, r)
	if os.Getenv("DEV") == "1" && r.Method == http.MethodOptions {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[panic] RestartAll handler crashed: %v", r)
		}
	}()
	log.Println("[service] UDP restart (manual)")
	restartUDPListener()
	restartOSCService()
	w.WriteHeader(http.StatusOK)
}

// Telemetry field config for enabling/disabling forwarding
var telemetryFieldsConfigPath string

type TelemetryFieldConfig struct {
	Enabled map[string]bool `json:"enabled"`
}

var TelemetryFields TelemetryFieldConfig

func InitTelemetryFieldsConfig() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	appDir := filepath.Join(configDir, "f1-telem-bridge")
	os.MkdirAll(appDir, 0755)
	telemetryFieldsConfigPath = filepath.Join(appDir, "telemetry_fields.json")

	if _, err := os.Stat(telemetryFieldsConfigPath); os.IsNotExist(err) {
		TelemetryFields = TelemetryFieldConfig{
			Enabled: map[string]bool{
				"Speed":    true,
				"Throttle": true,
				"Steer":    true,
				"Brake":    true,
				"Clutch":   true,
				"Gear":     true,
				"RPM":      true,
				// Add more fields as you expand the struct
			},
		}
		SaveTelemetryFieldsConfig()
	} else {
		LoadTelemetryFieldsConfig()
	}
}

func SaveTelemetryFieldsConfig() {
	f, err := os.Create(telemetryFieldsConfigPath)
	if err != nil {
		log.Printf("[error] Could not create telemetry fields config file: %v", err)
		return
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(TelemetryFields); err != nil {
		log.Printf("[error] Could not encode telemetry fields config: %v", err)
	}
}

func LoadTelemetryFieldsConfig() {
	f, err := os.Open(telemetryFieldsConfigPath)
	if err != nil {
		log.Printf("[error] Could not open telemetry fields config file: %v", err)
		return
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&TelemetryFields); err != nil {
		log.Printf("[error] Could not decode telemetry fields config: %v", err)
	}
}

// API for getting/setting enabled telemetry fields
func handleTelemetryFieldsAPI(w http.ResponseWriter, r *http.Request) {
	setCORS(w, r)
	if os.Getenv("DEV") == "1" && r.Method == http.MethodOptions {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[panic] TelemetryFields API handler crashed: %v", r)
		}
	}()

	if r.Method == http.MethodGet {
		json.NewEncoder(w).Encode(TelemetryFields)
	} else if r.Method == http.MethodPost {
		json.NewDecoder(r.Body).Decode(&TelemetryFields)
		SaveTelemetryFieldsConfig()
		w.WriteHeader(http.StatusOK)
	}
}

// ForwardingConfig controls which F1 25 UDP packet types are forwarded/broadcast
// Set to true to forward, false to ignore
var ForwardingConfig = struct {
	ForwardMotion              bool
	ForwardSession             bool
	ForwardLapData             bool
	ForwardEvent               bool
	ForwardParticipants        bool
	ForwardCarSetups           bool
	ForwardCarTelemetry        bool
	ForwardCarStatus           bool
	ForwardFinalClassification bool
	ForwardLobbyInfo           bool
	ForwardCarDamage           bool
	ForwardSessionHistory      bool
	ForwardTyreSets            bool
	ForwardMotionEx            bool
	ForwardTimeTrial           bool
	ForwardLapPositions        bool
}{
	ForwardMotion:              true,
	ForwardSession:             true,
	ForwardLapData:             true,
	ForwardEvent:               true,
	ForwardParticipants:        true,
	ForwardCarSetups:           true,
	ForwardCarTelemetry:        true,
	ForwardCarStatus:           true,
	ForwardFinalClassification: true,
	ForwardLobbyInfo:           true,
	ForwardCarDamage:           true,
	ForwardSessionHistory:      true,
	ForwardTyreSets:            true,
	ForwardMotionEx:            true,
	ForwardTimeTrial:           true,
	ForwardLapPositions:        true,
}

// OSC Address Mapping
var oscAddressesConfigPath string

type OSCAddressEntry struct {
	Address   string `json:"address"`
	ValueType string `json:"type"`
	Enabled   bool   `json:"enabled"`
	AllowZero bool   `json:"allowZero"`
}

var OSCAddresses = map[string]OSCAddressEntry{
	// Car Telemetry
	"Speed":                     {Address: "/car/speed", ValueType: "float", Enabled: true},
	"Throttle":                  {Address: "/car/throttle", ValueType: "float", Enabled: true},
	"Steer":                     {Address: "/car/steer", ValueType: "float", Enabled: true},
	"Brake":                     {Address: "/car/brake", ValueType: "float", Enabled: true},
	"Clutch":                    {Address: "/car/clutch", ValueType: "int", Enabled: true},
	"Gear":                      {Address: "/car/gear", ValueType: "int", Enabled: true},
	"EngineRPM":                 {Address: "/car/engine_rpm", ValueType: "int", Enabled: true},
	"DRS":                       {Address: "/car/drs", ValueType: "int", Enabled: true},
	"RevLightsPercent":          {Address: "/car/rev_lights_percent", ValueType: "int", Enabled: true},
	"RevLightsBitValue":         {Address: "/car/rev_lights_bits", ValueType: "int", Enabled: true},
	"BrakesTemperatureRL":       {Address: "/car/brakes_temp/rl", ValueType: "int", Enabled: true},
	"BrakesTemperatureRR":       {Address: "/car/brakes_temp/rr", ValueType: "int", Enabled: true},
	"BrakesTemperatureFL":       {Address: "/car/brakes_temp/fl", ValueType: "int", Enabled: true},
	"BrakesTemperatureFR":       {Address: "/car/brakes_temp/fr", ValueType: "int", Enabled: true},
	"TyresSurfaceTemperatureRL": {Address: "/car/tyres_surface_temp/rl", ValueType: "int", Enabled: true},
	"TyresSurfaceTemperatureRR": {Address: "/car/tyres_surface_temp/rr", ValueType: "int", Enabled: true},
	"TyresSurfaceTemperatureFL": {Address: "/car/tyres_surface_temp/fl", ValueType: "int", Enabled: true},
	"TyresSurfaceTemperatureFR": {Address: "/car/tyres_surface_temp/fr", ValueType: "int", Enabled: true},
	"TyresInnerTemperatureRL":   {Address: "/car/tyres_inner_temp/rl", ValueType: "int", Enabled: true},
	"TyresInnerTemperatureRR":   {Address: "/car/tyres_inner_temp/rr", ValueType: "int", Enabled: true},
	"TyresInnerTemperatureFL":   {Address: "/car/tyres_inner_temp/fl", ValueType: "int", Enabled: true},
	"TyresInnerTemperatureFR":   {Address: "/car/tyres_inner_temp/fr", ValueType: "int", Enabled: true},
	"EngineTemperature":         {Address: "/car/engine_temp", ValueType: "int", Enabled: true},
	"TyresPressureRL":           {Address: "/car/tyres_pressure/rl", ValueType: "float", Enabled: true},
	"TyresPressureRR":           {Address: "/car/tyres_pressure/rr", ValueType: "float", Enabled: true},
	"TyresPressureFL":           {Address: "/car/tyres_pressure/fl", ValueType: "float", Enabled: true},
	"TyresPressureFR":           {Address: "/car/tyres_pressure/fr", ValueType: "float", Enabled: true},
	"SurfaceTypeRL":             {Address: "/car/surface_type/rl", ValueType: "int", Enabled: true},
	"SurfaceTypeRR":             {Address: "/car/surface_type/rr", ValueType: "int", Enabled: true},
	"SurfaceTypeFL":             {Address: "/car/surface_type/fl", ValueType: "int", Enabled: true},
	"SurfaceTypeFR":             {Address: "/car/surface_type/fr", ValueType: "int", Enabled: true},

	// Motion (player car only, but can be extended for all cars)
	"WorldPositionX":     {Address: "/motion/world_pos/x", ValueType: "float", Enabled: true},
	"WorldPositionY":     {Address: "/motion/world_pos/y", ValueType: "float", Enabled: true},
	"WorldPositionZ":     {Address: "/motion/world_pos/z", ValueType: "float", Enabled: true},
	"WorldVelocityX":     {Address: "/motion/world_vel/x", ValueType: "float", Enabled: true},
	"WorldVelocityY":     {Address: "/motion/world_vel/y", ValueType: "float", Enabled: true},
	"WorldVelocityZ":     {Address: "/motion/world_vel/z", ValueType: "float", Enabled: true},
	"Yaw":                {Address: "/motion/yaw", ValueType: "float", Enabled: true},
	"Pitch":              {Address: "/motion/pitch", ValueType: "float", Enabled: true},
	"Roll":               {Address: "/motion/roll", ValueType: "float", Enabled: true},
	"GForceLateral":      {Address: "/motion/gforce/lateral", ValueType: "float", Enabled: true},
	"GForceLongitudinal": {Address: "/motion/gforce/longitudinal", ValueType: "float", Enabled: true},
	"GForceVertical":     {Address: "/motion/gforce/vertical", ValueType: "float", Enabled: true},

	// MotionEx (per wheel)
	"WheelSpeedRL":      {Address: "/motion_ex/wheel_speed/rl", ValueType: "float", Enabled: true},
	"WheelSpeedRR":      {Address: "/motion_ex/wheel_speed/rr", ValueType: "float", Enabled: true},
	"WheelSpeedFL":      {Address: "/motion_ex/wheel_speed/fl", ValueType: "float", Enabled: true},
	"WheelSpeedFR":      {Address: "/motion_ex/wheel_speed/fr", ValueType: "float", Enabled: true},
	"WheelSlipRatioRL":  {Address: "/motion_ex/wheel_slip_ratio/rl", ValueType: "float", Enabled: true},
	"WheelSlipRatioRR":  {Address: "/motion_ex/wheel_slip_ratio/rr", ValueType: "float", Enabled: true},
	"WheelSlipRatioFL":  {Address: "/motion_ex/wheel_slip_ratio/fl", ValueType: "float", Enabled: true},
	"WheelSlipRatioFR":  {Address: "/motion_ex/wheel_slip_ratio/fr", ValueType: "float", Enabled: true},
	"WheelSlipAngleRL":  {Address: "/motion_ex/wheel_slip_angle/rl", ValueType: "float", Enabled: true},
	"WheelSlipAngleRR":  {Address: "/motion_ex/wheel_slip_angle/rr", ValueType: "float", Enabled: true},
	"WheelSlipAngleFL":  {Address: "/motion_ex/wheel_slip_angle/fl", ValueType: "float", Enabled: true},
	"WheelSlipAngleFR":  {Address: "/motion_ex/wheel_slip_angle/fr", ValueType: "float", Enabled: true},
	"WheelLatForceRL":   {Address: "/motion_ex/wheel_lat_force/rl", ValueType: "float", Enabled: true},
	"WheelLatForceRR":   {Address: "/motion_ex/wheel_lat_force/rr", ValueType: "float", Enabled: true},
	"WheelLatForceFL":   {Address: "/motion_ex/wheel_lat_force/fl", ValueType: "float", Enabled: true},
	"WheelLatForceFR":   {Address: "/motion_ex/wheel_lat_force/fr", ValueType: "float", Enabled: true},
	"WheelLongForceRL":  {Address: "/motion_ex/wheel_long_force/rl", ValueType: "float", Enabled: true},
	"WheelLongForceRR":  {Address: "/motion_ex/wheel_long_force/rr", ValueType: "float", Enabled: true},
	"WheelLongForceFL":  {Address: "/motion_ex/wheel_long_force/fl", ValueType: "float", Enabled: true},
	"WheelLongForceFR":  {Address: "/motion_ex/wheel_long_force/fr", ValueType: "float", Enabled: true},
	"WheelVertForceRL":  {Address: "/motion_ex/wheel_vert_force/rl", ValueType: "float", Enabled: true},
	"WheelVertForceRR":  {Address: "/motion_ex/wheel_vert_force/rr", ValueType: "float", Enabled: true},
	"WheelVertForceFL":  {Address: "/motion_ex/wheel_vert_force/fl", ValueType: "float", Enabled: true},
	"WheelVertForceFR":  {Address: "/motion_ex/wheel_vert_force/fr", ValueType: "float", Enabled: true},
	"WheelCamberRL":     {Address: "/motion_ex/wheel_camber/rl", ValueType: "float", Enabled: true},
	"WheelCamberRR":     {Address: "/motion_ex/wheel_camber/rr", ValueType: "float", Enabled: true},
	"WheelCamberFL":     {Address: "/motion_ex/wheel_camber/fl", ValueType: "float", Enabled: true},
	"WheelCamberFR":     {Address: "/motion_ex/wheel_camber/fr", ValueType: "float", Enabled: true},
	"WheelCamberGainRL": {Address: "/motion_ex/wheel_camber_gain/rl", ValueType: "float", Enabled: true},
	"WheelCamberGainRR": {Address: "/motion_ex/wheel_camber_gain/rr", ValueType: "float", Enabled: true},
	"WheelCamberGainFL": {Address: "/motion_ex/wheel_camber_gain/fl", ValueType: "float", Enabled: true},
	"WheelCamberGainFR": {Address: "/motion_ex/wheel_camber_gain/fr", ValueType: "float", Enabled: true},

	// Lap Data
	"LastLapTimeInMS":       {Address: "/lap/last_lap_time_ms", ValueType: "int", Enabled: true},
	"CurrentLapTimeInMS":    {Address: "/lap/current_lap_time_ms", ValueType: "int", Enabled: true},
	"LapDistance":           {Address: "/lap/lap_distance", ValueType: "float", Enabled: true},
	"TotalDistance":         {Address: "/lap/total_distance", ValueType: "float", Enabled: true},
	"CarPosition":           {Address: "/lap/car_position", ValueType: "int", Enabled: true},
	"CurrentLapNum":         {Address: "/lap/current_lap_num", ValueType: "int", Enabled: true},
	"PitStatus":             {Address: "/lap/pit_status", ValueType: "int", Enabled: true},
	"NumPitStops":           {Address: "/lap/num_pit_stops", ValueType: "int", Enabled: true},
	"Sector":                {Address: "/lap/sector", ValueType: "int", Enabled: true},
	"CurrentLapInvalid":     {Address: "/lap/current_lap_invalid", ValueType: "int", Enabled: true},
	"Penalties":             {Address: "/lap/penalties", ValueType: "int", Enabled: true},
	"TotalWarnings":         {Address: "/lap/total_warnings", ValueType: "int", Enabled: true},
	"CornerCuttingWarnings": {Address: "/lap/corner_cutting_warnings", ValueType: "int", Enabled: true},
	"GridPosition":          {Address: "/lap/grid_position", ValueType: "int", Enabled: true},

	// Car Status
	"TractionControl":         {Address: "/status/traction_control", ValueType: "int", Enabled: true},
	"AntiLockBrakes":          {Address: "/status/anti_lock_brakes", ValueType: "int", Enabled: true},
	"FuelMix":                 {Address: "/status/fuel_mix", ValueType: "int", Enabled: true},
	"FrontBrakeBias":          {Address: "/status/front_brake_bias", ValueType: "int", Enabled: true},
	"PitLimiterStatus":        {Address: "/status/pit_limiter", ValueType: "int", Enabled: true},
	"FuelInTank":              {Address: "/status/fuel_in_tank", ValueType: "float", Enabled: true},
	"FuelCapacity":            {Address: "/status/fuel_capacity", ValueType: "float", Enabled: true},
	"FuelRemainingLaps":       {Address: "/status/fuel_remaining_laps", ValueType: "float", Enabled: true},
	"MaxRPM":                  {Address: "/status/max_rpm", ValueType: "int", Enabled: true},
	"IdleRPM":                 {Address: "/status/idle_rpm", ValueType: "int", Enabled: true},
	"MaxGears":                {Address: "/status/max_gears", ValueType: "int", Enabled: true},
	"DRSAllowed":              {Address: "/status/drs_allowed", ValueType: "int", Enabled: true},
	"DRSActivationDistance":   {Address: "/status/drs_activation_distance", ValueType: "int", Enabled: true},
	"ActualTyreCompound":      {Address: "/status/actual_tyre_compound", ValueType: "int", Enabled: true},
	"VisualTyreCompound":      {Address: "/status/visual_tyre_compound", ValueType: "int", Enabled: true},
	"TyresAgeLaps":            {Address: "/status/tyres_age_laps", ValueType: "int", Enabled: true},
	"VehicleFIAFlags":         {Address: "/status/vehicle_fia_flags", ValueType: "int", Enabled: true},
	"EnginePowerICE":          {Address: "/status/engine_power_ice", ValueType: "float", Enabled: true},
	"EnginePowerMGUK":         {Address: "/status/engine_power_mguk", ValueType: "float", Enabled: true},
	"ERSStoreEnergy":          {Address: "/status/ers_store_energy", ValueType: "float", Enabled: true},
	"ERSDeployMode":           {Address: "/status/ers_deploy_mode", ValueType: "int", Enabled: true},
	"ERSHarvestedThisLapMGUK": {Address: "/status/ers_harvested_mguk", ValueType: "float", Enabled: true},
	"ERSHarvestedThisLapMGUH": {Address: "/status/ers_harvested_mguh", ValueType: "float", Enabled: true},
	"ERSDeployedThisLap":      {Address: "/status/ers_deployed", ValueType: "float", Enabled: true},
	"NetworkPaused":           {Address: "/status/network_paused", ValueType: "int", Enabled: true},

	// Car Damage
	"TyresWearRL":          {Address: "/damage/tyres_wear/rl", ValueType: "float", Enabled: true},
	"TyresWearRR":          {Address: "/damage/tyres_wear/rr", ValueType: "float", Enabled: true},
	"TyresWearFL":          {Address: "/damage/tyres_wear/fl", ValueType: "float", Enabled: true},
	"TyresWearFR":          {Address: "/damage/tyres_wear/fr", ValueType: "float", Enabled: true},
	"TyresDamageRL":        {Address: "/damage/tyres_damage/rl", ValueType: "int", Enabled: true},
	"TyresDamageRR":        {Address: "/damage/tyres_damage/rr", ValueType: "int", Enabled: true},
	"TyresDamageFL":        {Address: "/damage/tyres_damage/fl", ValueType: "int", Enabled: true},
	"TyresDamageFR":        {Address: "/damage/tyres_damage/fr", ValueType: "int", Enabled: true},
	"BrakesDamageRL":       {Address: "/damage/brakes_damage/rl", ValueType: "int", Enabled: true},
	"BrakesDamageRR":       {Address: "/damage/brakes_damage/rr", ValueType: "int", Enabled: true},
	"BrakesDamageFL":       {Address: "/damage/brakes_damage/fl", ValueType: "int", Enabled: true},
	"BrakesDamageFR":       {Address: "/damage/brakes_damage/fr", ValueType: "int", Enabled: true},
	"FrontLeftWingDamage":  {Address: "/damage/front_left_wing", ValueType: "int", Enabled: true},
	"FrontRightWingDamage": {Address: "/damage/front_right_wing", ValueType: "int", Enabled: true},
	"RearWingDamage":       {Address: "/damage/rear_wing", ValueType: "int", Enabled: true},
	"FloorDamage":          {Address: "/damage/floor", ValueType: "int", Enabled: true},
	"DiffuserDamage":       {Address: "/damage/diffuser", ValueType: "int", Enabled: true},
	"SidepodDamage":        {Address: "/damage/sidepod", ValueType: "int", Enabled: true},
	"DRSFault":             {Address: "/damage/drs_fault", ValueType: "int", Enabled: true},
	"ERSFault":             {Address: "/damage/ers_fault", ValueType: "int", Enabled: true},
	"GearBoxDamage":        {Address: "/damage/gearbox", ValueType: "int", Enabled: true},
	"EngineDamage":         {Address: "/damage/engine", ValueType: "int", Enabled: true},
	"EngineMGUHWear":       {Address: "/damage/engine_mguh_wear", ValueType: "int", Enabled: true},
	"EngineESWear":         {Address: "/damage/engine_es_wear", ValueType: "int", Enabled: true},
	"EngineCEWear":         {Address: "/damage/engine_ce_wear", ValueType: "int", Enabled: true},
	"EngineICEWear":        {Address: "/damage/engine_ice_wear", ValueType: "int", Enabled: true},
	"EngineMGUKWear":       {Address: "/damage/engine_mguk_wear", ValueType: "int", Enabled: true},
	"EngineTCWear":         {Address: "/damage/engine_tc_wear", ValueType: "int", Enabled: true},
	"EngineBlown":          {Address: "/damage/engine_blown", ValueType: "int", Enabled: true},
	"EngineSeized":         {Address: "/damage/engine_seized", ValueType: "int", Enabled: true},

	// Session
	"Session_Weather":                         {Address: "/session/weather", ValueType: "int", Enabled: true},
	"Session_TrackTemperature":                {Address: "/session/track_temperature", ValueType: "int", Enabled: true},
	"Session_AirTemperature":                  {Address: "/session/air_temperature", ValueType: "int", Enabled: true},
	"Session_TotalLaps":                       {Address: "/session/total_laps", ValueType: "int", Enabled: true},
	"Session_TrackLength":                     {Address: "/session/track_length", ValueType: "int", Enabled: true},
	"Session_SessionType":                     {Address: "/session/session_type", ValueType: "int", Enabled: true},
	"Session_TrackId":                         {Address: "/session/track_id", ValueType: "int", Enabled: true},
	"Session_Formula":                         {Address: "/session/formula", ValueType: "int", Enabled: true},
	"Session_SessionTimeLeft":                 {Address: "/session/session_time_left", ValueType: "int", Enabled: true},
	"Session_SessionDuration":                 {Address: "/session/session_duration", ValueType: "int", Enabled: true},
	"Session_PitSpeedLimit":                   {Address: "/session/pit_speed_limit", ValueType: "int", Enabled: true},
	"Session_GamePaused":                      {Address: "/session/game_paused", ValueType: "int", Enabled: true},
	"Session_IsSpectating":                    {Address: "/session/is_spectating", ValueType: "int", Enabled: true},
	"Session_SpectatorCarIndex":               {Address: "/session/spectator_car_index", ValueType: "int", Enabled: true},
	"Session_SliProNativeSupport":             {Address: "/session/sli_pro_native_support", ValueType: "int", Enabled: true},
	"Session_NumMarshalZones":                 {Address: "/session/num_marshal_zones", ValueType: "int", Enabled: true},
	"Session_SafetyCarStatus":                 {Address: "/session/safety_car_status", ValueType: "int", Enabled: true},
	"Session_NetworkGame":                     {Address: "/session/network_game", ValueType: "int", Enabled: true},
	"Session_NumWeatherForecastSamples":       {Address: "/session/num_weather_forecast_samples", ValueType: "int", Enabled: true},
	"Session_ForecastAccuracy":                {Address: "/session/forecast_accuracy", ValueType: "int", Enabled: true},
	"Session_AIDifficulty":                    {Address: "/session/ai_difficulty", ValueType: "int", Enabled: true},
	"Session_SeasonLinkIdentifier":            {Address: "/session/season_link_identifier", ValueType: "int", Enabled: true},
	"Session_WeekendLinkIdentifier":           {Address: "/session/weekend_link_identifier", ValueType: "int", Enabled: true},
	"Session_SessionLinkIdentifier":           {Address: "/session/session_link_identifier", ValueType: "int", Enabled: true},
	"Session_PitStopWindowIdealLap":           {Address: "/session/pit_stop_window_ideal_lap", ValueType: "int", Enabled: true},
	"Session_PitStopWindowLatestLap":          {Address: "/session/pit_stop_window_latest_lap", ValueType: "int", Enabled: true},
	"Session_PitStopRejoinPosition":           {Address: "/session/pit_stop_rejoin_position", ValueType: "int", Enabled: true},
	"Session_SteeringAssist":                  {Address: "/session/steering_assist", ValueType: "int", Enabled: true},
	"Session_BrakingAssist":                   {Address: "/session/braking_assist", ValueType: "int", Enabled: true},
	"Session_GearboxAssist":                   {Address: "/session/gearbox_assist", ValueType: "int", Enabled: true},
	"Session_PitAssist":                       {Address: "/session/pit_assist", ValueType: "int", Enabled: true},
	"Session_PitReleaseAssist":                {Address: "/session/pit_release_assist", ValueType: "int", Enabled: true},
	"Session_ERSAssist":                       {Address: "/session/ers_assist", ValueType: "int", Enabled: true},
	"Session_DRSAssist":                       {Address: "/session/drs_assist", ValueType: "int", Enabled: true},
	"Session_DynamicRacingLine":               {Address: "/session/dynamic_racing_line", ValueType: "int", Enabled: true},
	"Session_DynamicRacingLineType":           {Address: "/session/dynamic_racing_line_type", ValueType: "int", Enabled: true},
	"Session_GameMode":                        {Address: "/session/game_mode", ValueType: "int", Enabled: true},
	"Session_RuleSet":                         {Address: "/session/rule_set", ValueType: "int", Enabled: true},
	"Session_TimeOfDay":                       {Address: "/session/time_of_day", ValueType: "int", Enabled: true},
	"Session_SessionLength":                   {Address: "/session/session_length", ValueType: "int", Enabled: true},
	"Session_SpeedUnitsLeadPlayer":            {Address: "/session/speed_units_lead_player", ValueType: "int", Enabled: true},
	"Session_TemperatureUnitsLeadPlayer":      {Address: "/session/temperature_units_lead_player", ValueType: "int", Enabled: true},
	"Session_SpeedUnitsSecondaryPlayer":       {Address: "/session/speed_units_secondary_player", ValueType: "int", Enabled: true},
	"Session_TemperatureUnitsSecondaryPlayer": {Address: "/session/temperature_units_secondary_player", ValueType: "int", Enabled: true},
	"Session_NumSafetyCarPeriods":             {Address: "/session/num_safety_car_periods", ValueType: "int", Enabled: true},
	"Session_NumVirtualSafetyCarPeriods":      {Address: "/session/num_virtual_safety_car_periods", ValueType: "int", Enabled: true},
	"Session_NumRedFlagPeriods":               {Address: "/session/num_red_flag_periods", ValueType: "int", Enabled: true},
	"Session_EqualCarPerformance":             {Address: "/session/equal_car_performance", ValueType: "int", Enabled: true},
	"Session_RecoveryMode":                    {Address: "/session/recovery_mode", ValueType: "int", Enabled: true},
	"Session_FlashbackLimit":                  {Address: "/session/flashback_limit", ValueType: "int", Enabled: true},
	"Session_SurfaceType":                     {Address: "/session/surface_type", ValueType: "int", Enabled: true},
	"Session_LowFuelMode":                     {Address: "/session/low_fuel_mode", ValueType: "int", Enabled: true},
	"Session_RaceStarts":                      {Address: "/session/race_starts", ValueType: "int", Enabled: true},
	"Session_TyreTemperature":                 {Address: "/session/tyre_temperature", ValueType: "int", Enabled: true},
	"Session_PitLaneTyreSim":                  {Address: "/session/pit_lane_tyre_sim", ValueType: "int", Enabled: true},
	"Session_CarDamage":                       {Address: "/session/car_damage", ValueType: "int", Enabled: true},
	"Session_CarDamageRate":                   {Address: "/session/car_damage_rate", ValueType: "int", Enabled: true},
	"Session_Collisions":                      {Address: "/session/collisions", ValueType: "int", Enabled: true},
	"Session_CollisionsOffForFirstLapOnly":    {Address: "/session/collisions_off_for_first_lap_only", ValueType: "int", Enabled: true},
	"Session_MpUnsafePitRelease":              {Address: "/session/mp_unsafe_pit_release", ValueType: "int", Enabled: true},
	"Session_MpOffForGriefing":                {Address: "/session/mp_off_for_griefing", ValueType: "int", Enabled: true},
	"Session_CornerCuttingStringency":         {Address: "/session/corner_cutting_stringency", ValueType: "int", Enabled: true},
	"Session_ParcFermeRules":                  {Address: "/session/parc_ferme_rules", ValueType: "int", Enabled: true},
	"Session_PitStopExperience":               {Address: "/session/pit_stop_experience", ValueType: "int", Enabled: true},
	"Session_SafetyCar":                       {Address: "/session/safety_car", ValueType: "int", Enabled: true},
	"Session_SafetyCarExperience":             {Address: "/session/safety_car_experience", ValueType: "int", Enabled: true},
	"Session_FormationLap":                    {Address: "/session/formation_lap", ValueType: "int", Enabled: true},
	"Session_FormationLapExperience":          {Address: "/session/formation_lap_experience", ValueType: "int", Enabled: true},
	"Session_RedFlags":                        {Address: "/session/red_flags", ValueType: "int", Enabled: true},
	"Session_AffectsLicenceLevelSolo":         {Address: "/session/affects_licence_level_solo", ValueType: "int", Enabled: true},
	"Session_AffectsLicenceLevelMP":           {Address: "/session/affects_licence_level_mp", ValueType: "int", Enabled: true},
	"Session_NumSessionsInWeekend":            {Address: "/session/num_sessions_in_weekend", ValueType: "int", Enabled: true},
	"Session_Sector2LapDistanceStart":         {Address: "/session/sector2_lap_distance_start", ValueType: "float", Enabled: true},
	"Session_Sector3LapDistanceStart":         {Address: "/session/sector3_lap_distance_start", ValueType: "float", Enabled: true},
	// MarshalZones (first 3 for example)
	"Session_MarshalZone0_ZoneStart": {Address: "/session/marshal_zone0/zone_start", ValueType: "float", Enabled: true},
	"Session_MarshalZone0_ZoneFlag":  {Address: "/session/marshal_zone0/zone_flag", ValueType: "int", Enabled: true},
	"Session_MarshalZone1_ZoneStart": {Address: "/session/marshal_zone1/zone_start", ValueType: "float", Enabled: true},
	"Session_MarshalZone1_ZoneFlag":  {Address: "/session/marshal_zone1/zone_flag", ValueType: "int", Enabled: true},
	"Session_MarshalZone2_ZoneStart": {Address: "/session/marshal_zone2/zone_start", ValueType: "float", Enabled: true},
	"Session_MarshalZone2_ZoneFlag":  {Address: "/session/marshal_zone2/zone_flag", ValueType: "int", Enabled: true},
	// WeatherForecastSamples (first 3 for example)
	"Session_WeatherForecastSample0_SessionType":            {Address: "/session/weather_forecast_sample0/session_type", ValueType: "int", Enabled: true},
	"Session_WeatherForecastSample0_TimeOffset":             {Address: "/session/weather_forecast_sample0/time_offset", ValueType: "int", Enabled: true},
	"Session_WeatherForecastSample0_Weather":                {Address: "/session/weather_forecast_sample0/weather", ValueType: "int", Enabled: true},
	"Session_WeatherForecastSample0_TrackTemperature":       {Address: "/session/weather_forecast_sample0/track_temperature", ValueType: "int", Enabled: true},
	"Session_WeatherForecastSample0_TrackTemperatureChange": {Address: "/session/weather_forecast_sample0/track_temperature_change", ValueType: "int", Enabled: true},
	"Session_WeatherForecastSample0_AirTemperature":         {Address: "/session/weather_forecast_sample0/air_temperature", ValueType: "int", Enabled: true},
	"Session_WeatherForecastSample0_AirTemperatureChange":   {Address: "/session/weather_forecast_sample0/air_temperature_change", ValueType: "int", Enabled: true},
	"Session_WeatherForecastSample0_RainPercentage":         {Address: "/session/weather_forecast_sample0/rain_percentage", ValueType: "int", Enabled: true},

	// Participants (first 3 for example)
	"Participant0_Name":        {Address: "/participants/0/name", ValueType: "string", Enabled: true},
	"Participant0_DriverId":    {Address: "/participants/0/driver_id", ValueType: "int", Enabled: true},
	"Participant0_TeamId":      {Address: "/participants/0/team_id", ValueType: "int", Enabled: true},
	"Participant0_Nationality": {Address: "/participants/0/nationality", ValueType: "int", Enabled: true},
	"Participant1_Name":        {Address: "/participants/1/name", ValueType: "string", Enabled: true},
	"Participant1_DriverId":    {Address: "/participants/1/driver_id", ValueType: "int", Enabled: true},
	"Participant1_TeamId":      {Address: "/participants/1/team_id", ValueType: "int", Enabled: true},
	"Participant1_Nationality": {Address: "/participants/1/nationality", ValueType: "int", Enabled: true},
	"Participant2_Name":        {Address: "/participants/2/name", ValueType: "string", Enabled: true},
	"Participant2_DriverId":    {Address: "/participants/2/driver_id", ValueType: "int", Enabled: true},
	"Participant2_TeamId":      {Address: "/participants/2/team_id", ValueType: "int", Enabled: true},
	"Participant2_Nationality": {Address: "/participants/2/nationality", ValueType: "int", Enabled: true},
	// Car Setups (first 3 for example)
	"CarSetup0_FrontWing": {Address: "/carsetup/0/front_wing", ValueType: "int", Enabled: true},
	"CarSetup0_RearWing":  {Address: "/carsetup/0/rear_wing", ValueType: "int", Enabled: true},
	"CarSetup1_FrontWing": {Address: "/carsetup/1/front_wing", ValueType: "int", Enabled: true},
	"CarSetup1_RearWing":  {Address: "/carsetup/1/rear_wing", ValueType: "int", Enabled: true},
	"CarSetup2_FrontWing": {Address: "/carsetup/2/front_wing", ValueType: "int", Enabled: true},
	"CarSetup2_RearWing":  {Address: "/carsetup/2/rear_wing", ValueType: "int", Enabled: true},
	// Final Classification (first 3 for example)
	"FinalClassification0_Position": {Address: "/finalclassification/0/position", ValueType: "int", Enabled: true},
	"FinalClassification0_NumLaps":  {Address: "/finalclassification/0/num_laps", ValueType: "int", Enabled: true},
	"FinalClassification1_Position": {Address: "/finalclassification/1/position", ValueType: "int", Enabled: true},
	"FinalClassification1_NumLaps":  {Address: "/finalclassification/1/num_laps", ValueType: "int", Enabled: true},
	"FinalClassification2_Position": {Address: "/finalclassification/2/position", ValueType: "int", Enabled: true},
	"FinalClassification2_NumLaps":  {Address: "/finalclassification/2/num_laps", ValueType: "int", Enabled: true},
	// Lobby Info (first 3 for example)
	"LobbyPlayer0_Name":   {Address: "/lobby/0/name", ValueType: "string", Enabled: true},
	"LobbyPlayer0_TeamId": {Address: "/lobby/0/team_id", ValueType: "int", Enabled: true},
	"LobbyPlayer1_Name":   {Address: "/lobby/1/name", ValueType: "string", Enabled: true},
	"LobbyPlayer1_TeamId": {Address: "/lobby/1/team_id", ValueType: "int", Enabled: true},
	"LobbyPlayer2_Name":   {Address: "/lobby/2/name", ValueType: "string", Enabled: true},
	"LobbyPlayer2_TeamId": {Address: "/lobby/2/team_id", ValueType: "int", Enabled: true},
	// Session History (first 3 laps for example)
	"SessionHistory_Lap0_LapTimeInMS": {Address: "/sessionhistory/lap0/lap_time_ms", ValueType: "int", Enabled: true},
	"SessionHistory_Lap1_LapTimeInMS": {Address: "/sessionhistory/lap1/lap_time_ms", ValueType: "int", Enabled: true},
	"SessionHistory_Lap2_LapTimeInMS": {Address: "/sessionhistory/lap2/lap_time_ms", ValueType: "int", Enabled: true},
	// Tyre Sets (first 3 for example)
	"TyreSet0_ActualTyreCompound": {Address: "/tyreset/0/actual_tyre_compound", ValueType: "int", Enabled: true},
	"TyreSet0_VisualTyreCompound": {Address: "/tyreset/0/visual_tyre_compound", ValueType: "int", Enabled: true},
	"TyreSet1_ActualTyreCompound": {Address: "/tyreset/1/actual_tyre_compound", ValueType: "int", Enabled: true},
	"TyreSet1_VisualTyreCompound": {Address: "/tyreset/1/visual_tyre_compound", ValueType: "int", Enabled: true},
	"TyreSet2_ActualTyreCompound": {Address: "/tyreset/2/actual_tyre_compound", ValueType: "int", Enabled: true},
	"TyreSet2_VisualTyreCompound": {Address: "/tyreset/2/visual_tyre_compound", ValueType: "int", Enabled: true},
	// Time Trial
	"TimeTrial_PlayerBest_LapTimeInMS":   {Address: "/timetrial/player_best/lap_time_ms", ValueType: "int", Enabled: true},
	"TimeTrial_PersonalBest_LapTimeInMS": {Address: "/timetrial/personal_best/lap_time_ms", ValueType: "int", Enabled: true},
	"TimeTrial_Rival_LapTimeInMS":        {Address: "/timetrial/rival/lap_time_ms", ValueType: "int", Enabled: true},
	// Lap Positions (first 3 laps, first 3 cars for example)
	"LapPositions_Lap0_Car0": {Address: "/lappositions/lap0/car0", ValueType: "int", Enabled: true},
	"LapPositions_Lap0_Car1": {Address: "/lappositions/lap0/car1", ValueType: "int", Enabled: true},
	"LapPositions_Lap0_Car2": {Address: "/lappositions/lap0/car2", ValueType: "int", Enabled: true},
	"LapPositions_Lap1_Car0": {Address: "/lappositions/lap1/car0", ValueType: "int", Enabled: true},
	"LapPositions_Lap1_Car1": {Address: "/lappositions/lap1/car1", ValueType: "int", Enabled: true},
	"LapPositions_Lap1_Car2": {Address: "/lappositions/lap1/car2", ValueType: "int", Enabled: true},
	"LapPositions_Lap2_Car0": {Address: "/lappositions/lap2/car0", ValueType: "int", Enabled: true},
	"LapPositions_Lap2_Car1": {Address: "/lappositions/lap2/car1", ValueType: "int", Enabled: true},
	"LapPositions_Lap2_Car2": {Address: "/lappositions/lap2/car2", ValueType: "int", Enabled: true},
}

func InitOSCAddressesConfig() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	appDir := filepath.Join(configDir, "f1-telem-bridge")
	os.MkdirAll(appDir, 0755)
	oscAddressesConfigPath = filepath.Join(appDir, "osc_addresses.json")

	if _, err := os.Stat(oscAddressesConfigPath); os.IsNotExist(err) {
		SaveOSCAddressesConfig()
	} else {
		LoadOSCAddressesConfig()
	}
}

func SaveOSCAddressesConfig() {
	f, err := os.Create(oscAddressesConfigPath)
	if err != nil {
		log.Printf("[error] Could not create OSC addresses config file: %v", err)
		return
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(OSCAddresses); err != nil {
		log.Printf("[error] Could not encode OSC addresses config: %v", err)
	}
}

func LoadOSCAddressesConfig() {
	f, err := os.Open(oscAddressesConfigPath)
	if err != nil {
		log.Printf("[error] Could not open OSC addresses config file: %v", err)
		return
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&OSCAddresses); err != nil {
		log.Printf("[error] Could not decode OSC addresses config: %v", err)
	}
}

// API for getting/setting OSC address mapping
func handleOSCAddressesAPI(w http.ResponseWriter, r *http.Request) {
	setCORS(w, r)
	if os.Getenv("DEV") == "1" && r.Method == http.MethodOptions {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[panic] OSCAddresses API handler crashed: %v", r)
		}
	}()

	if r.Method == http.MethodGet {
		json.NewEncoder(w).Encode(OSCAddresses)
	} else if r.Method == http.MethodPost {
		json.NewDecoder(r.Body).Decode(&OSCAddresses)
		SaveOSCAddressesConfig()
		w.WriteHeader(http.StatusOK)
	}
}

package main

import (
	"log"

	"github.com/hypebeast/go-osc/osc"
)

// Centralized OSC send function
func sendOSC(address string, value interface{}) {
	if !Config.EnableOSC {
		return
	}
	// Convert unsupported types to supported OSC types
	switch v := value.(type) {
	case uint8:
		value = int32(v)
	case uint16:
		value = int32(v)
	case uint32:
		value = int32(v)
	case int8:
		value = int32(v)
	case int16:
		value = int32(v)
	case float64:
		value = float32(v)
	}
	client := osc.NewClient(Config.OSCAddr, Config.OSCPort)
	log.Printf("[debug] Sending OSC message: %s %v", address, value)
	msg := osc.NewMessage(address, value)
	if err := client.Send(msg); err != nil {
		log.Printf("[error] OSC send failed: OSC - unsupported type: %T", value)
	}
}

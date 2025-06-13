package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"reflect"
	"time"
)

// F1 25 UDP Packet Header
const (
	PacketMotion              = 0
	PacketSession             = 1
	PacketLapData             = 2
	PacketEvent               = 3
	PacketParticipants        = 4
	PacketCarSetups           = 5
	PacketCarTelemetry        = 6
	PacketCarStatus           = 7
	PacketFinalClassification = 8
	PacketLobbyInfo           = 9
	PacketCarDamage           = 10
	PacketSessionHistory      = 11
	PacketTyreSets            = 12
	PacketMotionEx            = 13
	PacketTimeTrial           = 14
	PacketLapPositions        = 15
)

type PacketHeader struct {
	PacketFormat            uint16
	GameYear                uint8
	GameMajorVersion        uint8
	GameMinorVersion        uint8
	PacketVersion           uint8
	PacketId                uint8
	SessionUID              uint64
	SessionTime             float32
	FrameIdentifier         uint32
	OverallFrameIdentifier  uint32
	PlayerCarIndex          uint8
	SecondaryPlayerCarIndex uint8
}

func StartUDPListener() {
	addr := net.UDPAddr{
		IP:   net.ParseIP(Config.UDPAddr),
		Port: Config.UDPPort,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Failed to listen on UDP: %v", err)
	}
	defer conn.Close()

	log.Printf("Listening for F1 UDP on %s:%d\n", Config.UDPAddr, Config.UDPPort)

	buf := make([]byte, 2048)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("UDP read error:", err)
			continue
		}
		// Log raw UDP data with timestamp and packet name
		packetID := buf[6]
		packetName := PacketNames[packetID]
		if Config.DebugOutput {
			log.Printf("[raw] %s | %s | from %s: %x", time.Now().Format("15:04:05.000"), packetName, addr, buf[:n])
		}
		if n > 6 {
		} else {
			log.Printf("[raw] %s | unknown | from %s: %x", time.Now().Format("15:04:05.000"), addr, buf[:n])
		}
		handleUDPPacket(buf[:n])
	}
}

type CarTelemetryData struct {
	Speed                   uint16     // m_speed
	Throttle                float32    // m_throttle
	Steer                   float32    // m_steer
	Brake                   float32    // m_brake
	Clutch                  uint8      // m_clutch
	Gear                    int8       // m_gear
	RPM                     uint16     // <-- Add this field for compatibility
	EngineRPM               uint16     // m_engineRPM
	DRS                     uint8      // m_drs
	RevLightsPercent        uint8      // m_revLightsPercent
	RevLightsBitValue       uint16     // m_revLightsBitValue
	BrakesTemperature       [4]uint16  // m_brakesTemperature
	TyresSurfaceTemperature [4]uint8   // m_tyresSurfaceTemperature
	TyresInnerTemperature   [4]uint8   // m_tyresInnerTemperature
	EngineTemperature       uint16     // m_engineTemperature
	TyresPressure           [4]float32 // m_tyresPressure
	SurfaceType             [4]uint8   // m_surfaceType
}

// -------------------- F1 25 UDP Packet Structs --------------------
// These structs match the official F1 25 UDP spec for all packet types

type CarMotionData struct {
	WorldPositionX     float32
	WorldPositionY     float32
	WorldPositionZ     float32
	WorldVelocityX     float32
	WorldVelocityY     float32
	WorldVelocityZ     float32
	WorldForwardDirX   int16
	WorldForwradDirY   int16
	WorldForwardDirZ   int16
	WorldRightDirX     int16
	WorldRightDirY     int16
	WorldRightDirZ     int16
	GForceLateral      float32
	GForceLongitudinal float32
	GForceVertical     float32
	Yaw                float32
	Pitch              float32
	Roll               float32
}

type PacketMotionData struct {
	Header        PacketHeader
	CarMotionData [22]CarMotionData
}

type MarshalZone struct {
	ZoneStart float32
	ZoneFlag  int8
}

type WeatherForecastSample struct {
	SessionType            uint8
	TimeOffset             uint8
	Weather                uint8
	TrackTemperature       int8
	TrackTemperatureChange int8
	AirTemperature         int8
	AirTemperatureChange   int8
	RainPercentage         uint8
}

type PacketSessionData struct {
	Header                          PacketHeader
	Weather                         uint8
	TrackTemperature                int8
	AirTemperature                  int8
	TotalLaps                       uint8
	TrackLength                     uint16
	SessionType                     uint8
	TrackId                         int8
	Formula                         uint8
	SessionTimeLeft                 uint16
	SessionDuration                 uint16
	PitSpeedLimit                   uint8
	GamePaused                      uint8
	IsSpectating                    uint8
	SpectatorCarIndex               uint8
	SliProNativeSupport             uint8
	NumMarshalZones                 uint8
	MarshalZones                    [21]MarshalZone
	SafetyCarStatus                 uint8
	NetworkGame                     uint8
	NumWeatherForecastSamples       uint8
	WeatherForecastSamples          [64]WeatherForecastSample
	ForecastAccuracy                uint8
	AIDifficulty                    uint8
	SeasonLinkIdentifier            uint32
	WeekendLinkIdentifier           uint32
	SessionLinkIdentifier           uint32
	PitStopWindowIdealLap           uint8
	PitStopWindowLatestLap          uint8
	PitStopRejoinPosition           uint8
	SteeringAssist                  uint8
	BrakingAssist                   uint8
	GearboxAssist                   uint8
	PitAssist                       uint8
	PitReleaseAssist                uint8
	ERSAssist                       uint8
	DRSAssist                       uint8
	DynamicRacingLine               uint8
	DynamicRacingLineType           uint8
	GameMode                        uint8
	RuleSet                         uint8
	TimeOfDay                       uint32
	SessionLength                   uint8
	SpeedUnitsLeadPlayer            uint8
	TemperatureUnitsLeadPlayer      uint8
	SpeedUnitsSecondaryPlayer       uint8
	TemperatureUnitsSecondaryPlayer uint8
	NumSafetyCarPeriods             uint8
	NumVirtualSafetyCarPeriods      uint8
	NumRedFlagPeriods               uint8
	EqualCarPerformance             uint8
	RecoveryMode                    uint8
	FlashbackLimit                  uint8
	SurfaceType                     uint8
	LowFuelMode                     uint8
	RaceStarts                      uint8
	TyreTemperature                 uint8
	PitLaneTyreSim                  uint8
	CarDamage                       uint8
	CarDamageRate                   uint8
	Collisions                      uint8
	CollisionsOffForFirstLapOnly    uint8
	MpUnsafePitRelease              uint8
	MpOffForGriefing                uint8
	CornerCuttingStringency         uint8
	ParcFermeRules                  uint8
	PitStopExperience               uint8
	SafetyCar                       uint8
	SafetyCarExperience             uint8
	FormationLap                    uint8
	FormationLapExperience          uint8
	RedFlags                        uint8
	AffectsLicenceLevelSolo         uint8
	AffectsLicenceLevelMP           uint8
	NumSessionsInWeekend            uint8
	WeekendStructure                [12]uint8
	Sector2LapDistanceStart         float32
	Sector3LapDistanceStart         float32
}

type LapData struct {
	LastLapTimeInMS              uint32
	CurrentLapTimeInMS           uint32
	Sector1TimeMSPart            uint16
	Sector1TimeMinutesPart       uint8
	Sector2TimeMSPart            uint16
	Sector2TimeMinutesPart       uint8
	DeltaToCarInFrontMSPart      uint16
	DeltaToCarInFrontMinutesPart uint8
	DeltaToRaceLeaderMSPart      uint16
	DeltaToRaceLeaderMinutesPart uint8
	LapDistance                  float32
	TotalDistance                float32
	SafetyCarDelta               float32
	CarPosition                  uint8
	CurrentLapNum                uint8
	PitStatus                    uint8
	NumPitStops                  uint8
	Sector                       uint8
	CurrentLapInvalid            uint8
	Penalties                    uint8
	TotalWarnings                uint8
	CornerCuttingWarnings        uint8
	NumUnservedDriveThroughPens  uint8
	NumUnservedStopGoPens        uint8
	GridPosition                 uint8
	DriverStatus                 uint8
	ResultStatus                 uint8
	PitLaneTimerActive           uint8
	PitLaneTimeInLaneInMS        uint16
	PitStopTimerInMS             uint16
	PitStopShouldServePen        uint8
	SpeedTrapFastestSpeed        float32
	SpeedTrapFastestLap          uint8
}

type LapDataPacket struct {
	Header               PacketHeader
	LapData              [22]LapData
	TimeTrialPBCarIdx    uint8
	TimeTrialRivalCarIdx uint8
}

type PacketEventData struct {
	Header          PacketHeader
	EventStringCode [4]uint8
	EventDetails    [24]byte // Union, handled per event type
}

type LiveryColour struct {
	Red   uint8
	Green uint8
	Blue  uint8
}

type ParticipantData struct {
	AIControlled    uint8
	DriverId        uint8
	NetworkId       uint8
	TeamId          uint8
	MyTeam          uint8
	RaceNumber      uint8
	Nationality     uint8
	Name            [32]byte
	YourTelemetry   uint8
	ShowOnlineNames uint8
	TechLevel       uint16
	Platform        uint8
	NumColours      uint8
	LiveryColours   [4]LiveryColour
}

type PacketParticipantsData struct {
	Header        PacketHeader
	NumActiveCars uint8
	Participants  [22]ParticipantData
}

type CarSetupData struct {
	FrontWing              uint8
	RearWing               uint8
	OnThrottle             uint8
	OffThrottle            uint8
	FrontCamber            float32
	RearCamber             float32
	FrontToe               float32
	RearToe                float32
	FrontSuspension        uint8
	RearSuspension         uint8
	FrontAntiRollBar       uint8
	RearAntiRollBar        uint8
	FrontSuspensionHeight  uint8
	RearSuspensionHeight   uint8
	BrakePressure          uint8
	BrakeBias              uint8
	EngineBraking          uint8
	RearLeftTyrePressure   float32
	RearRightTyrePressure  float32
	FrontLeftTyrePressure  float32
	FrontRightTyrePressure float32
	Ballast                uint8
	FuelLoad               float32
}

type PacketCarSetupData struct {
	Header             PacketHeader
	CarSetupData       [22]CarSetupData
	NextFrontWingValue float32
}

type CarStatusData struct {
	TractionControl         uint8
	AntiLockBrakes          uint8
	FuelMix                 uint8
	FrontBrakeBias          uint8
	PitLimiterStatus        uint8
	FuelInTank              float32
	FuelCapacity            float32
	FuelRemainingLaps       float32
	MaxRPM                  uint16
	IdleRPM                 uint16
	MaxGears                uint8
	DRSAllowed              uint8
	DRSActivationDistance   uint16
	ActualTyreCompound      uint8
	VisualTyreCompound      uint8
	TyresAgeLaps            uint8
	VehicleFIAFlags         int8
	EnginePowerICE          float32
	EnginePowerMGUK         float32
	ERSStoreEnergy          float32
	ERSDeployMode           uint8
	ERSHarvestedThisLapMGUK float32
	ERSHarvestedThisLapMGUH float32
	ERSDeployedThisLap      float32
	NetworkPaused           uint8
}

type PacketCarStatusData struct {
	Header        PacketHeader
	CarStatusData [22]CarStatusData
}

type FinalClassificationData struct {
	Position          uint8
	NumLaps           uint8
	GridPosition      uint8
	Points            uint8
	NumPitStops       uint8
	ResultStatus      uint8
	ResultReason      uint8
	BestLapTimeInMS   uint32
	TotalRaceTime     float64
	PenaltiesTime     uint8
	NumPenalties      uint8
	NumTyreStints     uint8
	TyreStintsActual  [8]uint8
	TyreStintsVisual  [8]uint8
	TyreStintsEndLaps [8]uint8
}

type PacketFinalClassificationData struct {
	Header             PacketHeader
	NumCars            uint8
	ClassificationData [22]FinalClassificationData
}

type LobbyInfoData struct {
	AIControlled    uint8
	TeamId          uint8
	Nationality     uint8
	Platform        uint8
	Name            [32]byte
	CarNumber       uint8
	YourTelemetry   uint8
	ShowOnlineNames uint8
	TechLevel       uint16
	ReadyStatus     uint8
}

type PacketLobbyInfoData struct {
	Header       PacketHeader
	NumPlayers   uint8
	LobbyPlayers [22]LobbyInfoData
}

type CarDamageData struct {
	TyresWear            [4]float32
	TyresDamage          [4]uint8
	BrakesDamage         [4]uint8
	TyreBlisters         [4]uint8
	FrontLeftWingDamage  uint8
	FrontRightWingDamage uint8
	RearWingDamage       uint8
	FloorDamage          uint8
	DiffuserDamage       uint8
	SidepodDamage        uint8
	DRSFault             uint8
	ERSFault             uint8
	GearBoxDamage        uint8
	EngineDamage         uint8
	EngineMGUHWear       uint8
	EngineESWear         uint8
	EngineCEWear         uint8
	EngineICEWear        uint8
	EngineMGUKWear       uint8
	EngineTCWear         uint8
	EngineBlown          uint8
	EngineSeized         uint8
}

type PacketCarDamageData struct {
	Header        PacketHeader
	CarDamageData [22]CarDamageData
}

type LapHistoryData struct {
	LapTimeInMS            uint32
	Sector1TimeMSPart      uint16
	Sector1TimeMinutesPart uint8
	Sector2TimeMSPart      uint16
	Sector2TimeMinutesPart uint8
	Sector3TimeMSPart      uint16
	Sector3TimeMinutesPart uint8
	LapValidBitFlags       uint8
}

type TyreStintHistoryData struct {
	EndLap             uint8
	TyreActualCompound uint8
	TyreVisualCompound uint8
}

type PacketSessionHistoryData struct {
	Header                PacketHeader
	CarIdx                uint8
	NumLaps               uint8
	NumTyreStints         uint8
	BestLapTimeLapNum     uint8
	BestSector1LapNum     uint8
	BestSector2LapNum     uint8
	BestSector3LapNum     uint8
	LapHistoryData        [100]LapHistoryData
	TyreStintsHistoryData [8]TyreStintHistoryData
}

type TyreSetData struct {
	ActualTyreCompound uint8
	VisualTyreCompound uint8
	Wear               uint8
	Available          uint8
	RecommendedSession uint8
	LifeSpan           uint8
	UsableLife         uint8
	LapDeltaTime       int16
	Fitted             uint8
}

type PacketTyreSetsData struct {
	Header      PacketHeader
	CarIdx      uint8
	TyreSetData [20]TyreSetData
	FittedIdx   uint8
}

type PacketMotionExData struct {
	Header                 PacketHeader
	SuspensionPosition     [4]float32
	SuspensionVelocity     [4]float32
	SuspensionAcceleration [4]float32
	WheelSpeed             [4]float32
	WheelSlipRatio         [4]float32
	WheelSlipAngle         [4]float32
	WheelLatForce          [4]float32
	WheelLongForce         [4]float32
	HeightOfCOGAboveGround float32
	LocalVelocityX         float32
	LocalVelocityY         float32
	LocalVelocityZ         float32
	AngularVelocityX       float32
	AngularVelocityY       float32
	AngularVelocityZ       float32
	AngularAccelerationX   float32
	AngularAccelerationY   float32
	AngularAccelerationZ   float32
	FrontWheelsAngle       float32
	WheelVertForce         [4]float32
	FrontAeroHeight        float32
	RearAeroHeight         float32
	FrontRollAngle         float32
	RearRollAngle          float32
	ChassisYaw             float32
	ChassisPitch           float32
	WheelCamber            [4]float32
	WheelCamberGain        [4]float32
}

type TimeTrialDataSet struct {
	CarIdx              uint8
	TeamId              uint8
	LapTimeInMS         uint32
	Sector1TimeInMS     uint32
	Sector2TimeInMS     uint32
	Sector3TimeInMS     uint32
	TractionControl     uint8
	GearboxAssist       uint8
	AntiLockBrakes      uint8
	EqualCarPerformance uint8
	CustomSetup         uint8
	Valid               uint8
}

type PacketTimeTrialData struct {
	Header                   PacketHeader
	PlayerSessionBestDataSet TimeTrialDataSet
	PersonalBestDataSet      TimeTrialDataSet
	RivalDataSet             TimeTrialDataSet
}

type PacketLapPositionsData struct {
	Header                PacketHeader
	NumLaps               uint8
	LapStart              uint8
	PositionForVehicleIdx [50][22]uint8
}

// Cleaner forwarding check using a map
var PacketForwardingConfig = map[uint8]bool{
	PacketMotion:              true,
	PacketSession:             true,
	PacketLapData:             true,
	PacketEvent:               true,
	PacketParticipants:        true,
	PacketCarSetups:           true,
	PacketCarTelemetry:        true,
	PacketCarStatus:           true,
	PacketFinalClassification: true,
	PacketLobbyInfo:           true,
	PacketCarDamage:           true,
	PacketSessionHistory:      true,
	PacketTyreSets:            true,
	PacketMotionEx:            true,
	PacketTimeTrial:           true,
	PacketLapPositions:        true,
}

var PacketNames = map[uint8]string{
	0:  "Motion",
	1:  "Session",
	2:  "LapData",
	3:  "Event",
	4:  "Participants",
	5:  "CarSetups",
	6:  "CarTelemetry",
	7:  "CarStatus",
	8:  "FinalClassification",
	9:  "LobbyInfo",
	10: "CarDamage",
	11: "SessionHistory",
	12: "TyreSets",
	13: "MotionEx",
	14: "TimeTrial",
	15: "LapPositions",
}

// Helper for decode, marshal, broadcast, and OSC forward
func decodeAndBroadcast[T any](data []byte, decodeFunc func([]byte) (T, error), packetName string, packetID uint8) {
	pkt, err := decodeFunc(data)
	if err != nil {
		log.Printf("[error] decode %s: %v", packetName, err)
		return
	}
	// Broadcast each field as "PacketName/FieldName value"
	v := reflect.ValueOf(pkt)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		broadcastStructFieldsToWS(v, packetName)
		sendStructFieldsToOSC(v)
	}
}

var lastWebSocketBroadcast time.Time

// Throttle and deduplicate for WebSocket and OSC
var (
	lastSentWS = make(map[string]struct {
		t time.Time
		v interface{}
	})
	lastSentOSC = make(map[string]struct {
		t time.Time
		v interface{}
	})
)

func getBroadcastInterval() time.Duration {
	rate := Config.BroadcastRateHz
	if rate <= 0 {
		return 500 * time.Millisecond // fallback default
	}
	return time.Second / time.Duration(rate)
}

// Helper for float comparison
func valuesEqual(a, b interface{}) bool {
	// Handle float32/float64 with epsilon
	switch va := a.(type) {
	case float32:
		vb, ok := b.(float32)
		if !ok {
			return false
		}
		return math.Abs(float64(va-vb)) < 1e-4
	case float64:
		vb, ok := b.(float64)
		if !ok {
			return false
		}
		return math.Abs(va-vb) < 1e-7
	default:
		return reflect.DeepEqual(a, b)
	}
}

func shouldSend(key string, value interface{}, lastSent map[string]struct {
	t time.Time
	v interface{}
}) bool {
	// Ignore zero values for float32/float64/int types
	switch v := value.(type) {
	case float32:
		if v == 0 {
			return false
		}
	case float64:
		if v == 0 {
			return false
		}
	case int:
		if v == 0 {
			return false
		}
	case int32:
		if v == 0 {
			return false
		}
	case int64:
		if v == 0 {
			return false
		}
	case uint:
		if v == 0 {
			return false
		}
	case uint32:
		if v == 0 {
			return false
		}
	case uint64:
		if v == 0 {
			return false
		}
	}
	entry, ok := lastSent[key]
	now := time.Now()
	interval := getBroadcastInterval()
	if ok {
		if now.Sub(entry.t) < interval {
			// Only send if value is different (with float tolerance)
			if valuesEqual(entry.v, value) {
				return false
			}
			// If value changed, still only send once per interval
			return false
		}
		if valuesEqual(entry.v, value) {
			return false
		}
	}
	return true
}

func updateLastSent(key string, value interface{}, lastSent map[string]struct {
	t time.Time
	v interface{}
}) {
	lastSent[key] = struct {
		t time.Time
		v interface{}
	}{time.Now(), value}
}

func broadcastStructFieldsToWS(v reflect.Value, packetName string) {
	typeOfV := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := typeOfV.Field(i)
		name := fieldType.Name
		if field.Kind() == reflect.Struct {
			broadcastStructFieldsToWS(field, packetName+"/"+name)
			continue
		}
		if field.Kind() == reflect.Array || field.Kind() == reflect.Slice {
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				key := fmt.Sprintf("%s/%s[%d]", packetName, name, j)
				if Config.DebugOutput {
					log.Printf("[debug] WebSocket key: %s value: %v", key, elem.Interface())
				}
				if shouldSend(key, elem.Interface(), lastSentWS) {
					msg := fmt.Sprintf("%s %v", key, elem.Interface())
					broadcast([]byte(msg))
					updateLastSent(key, elem.Interface(), lastSentWS)
				}
			}
			continue
		}
		key := fmt.Sprintf("%s/%s", packetName, name)
		if shouldSend(key, field.Interface(), lastSentWS) {
			msg := fmt.Sprintf("%s %v", key, field.Interface())
			broadcast([]byte(msg))
			updateLastSent(key, field.Interface(), lastSentWS)
		}
	}
}

func sendStructFieldsToOSC(v reflect.Value) {
	typeOfV := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := typeOfV.Field(i)
		name := fieldType.Name
		if entry, ok := OSCAddresses[name]; ok && entry.Enabled && Config.EnableOSC {
			key := entry.Address
			// Only send 0 if AllowZero is true for this address
			sendZero := false
			switch v := field.Interface().(type) {
			case float32, float64, int, int32, int64, uint, uint32, uint64:
				if v == 0 && !entry.AllowZero {
					continue
				}
				if v == 0 && entry.AllowZero {
					sendZero = true
				}
			}
			if shouldSend(key, field.Interface(), lastSentOSC) || sendZero {
				sendOSC(entry.Address, field.Interface())
				updateLastSent(key, field.Interface(), lastSentOSC)
			}
		}
		if field.Kind() == reflect.Struct {
			sendStructFieldsToOSC(field)
		}
		if field.Kind() == reflect.Array || field.Kind() == reflect.Slice {
			// Suffixes for 4-wheel arrays
			suffixes := []string{"RL", "RR", "FL", "FR"}
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				key := name
				if field.Len() == 4 && j < len(suffixes) {
					key = name + suffixes[j]
				}
				if entry, ok := OSCAddresses[key]; ok && entry.Enabled && Config.EnableOSC {
					oscKey := entry.Address
					if shouldSend(oscKey, elem.Interface(), lastSentOSC) {
						sendOSC(entry.Address, elem.Interface())
						updateLastSent(oscKey, elem.Interface(), lastSentOSC)
					}
				}
				if elem.Kind() == reflect.Struct {
					sendStructFieldsToOSC(elem)
				}
			}
		}
	}
}

func broadcastTelemetryFields(telemetry CarTelemetryData) {
	msg := fmt.Sprintf(
		"CarTelemetry/Speed %d | Throttle %.2f | Steer %.2f | Brake %.2f | Clutch %d | Gear %d | RPM %d",
		telemetry.Speed,
		telemetry.Throttle,
		telemetry.Steer,
		telemetry.Brake,
		telemetry.Clutch,
		telemetry.Gear,
		telemetry.RPM,
	)
	key := "CarTelemetry/summary"
	if shouldSend(key, msg, lastSentWS) {
		broadcast([]byte(msg))
		updateLastSent(key, msg, lastSentWS)
	}
	fields := map[string]interface{}{
		"Speed":     telemetry.Speed,
		"Throttle":  telemetry.Throttle,
		"Steer":     telemetry.Steer,
		"Brake":     telemetry.Brake,
		"Clutch":    telemetry.Clutch,
		"Gear":      telemetry.Gear,
		"EngineRPM": telemetry.EngineRPM, // Use EngineRPM as the key, not RPM
	}
	for k, v := range fields {
		if entry, ok := OSCAddresses[k]; ok && entry.Enabled && Config.EnableOSC {
			oscKey := entry.Address
			if shouldSend(oscKey, v, lastSentOSC) {
				sendOSC(entry.Address, v)
				updateLastSent(oscKey, v, lastSentOSC)
			}
		}
	}
}

// Main UDP handler dispatches based on PacketId
func handleUDPPacket(data []byte) {
	if len(data) < 24 {
		return
	}
	packetID := data[6] // FIX: packetId is at offset 6 per F1 25 spec
	if !PacketForwardingConfig[packetID] {
		return // Not enabled, skip processing
	}

	switch packetID {
	case PacketMotion:
		decodeAndBroadcast(data, decodeMotionPacket, "Motion", PacketMotion)
	case PacketSession:
		decodeAndBroadcast(data, decodeSessionPacket, "Session", PacketSession)
	case PacketLapData:
		decodeAndBroadcast(data, decodeLapDataPacket, "LapData", PacketLapData)
	case PacketEvent:
		decodeAndBroadcast(data, decodeEventPacket, "Event", PacketEvent)
	case PacketParticipants:
		decodeAndBroadcast(data, decodeParticipantsPacket, "Participants", PacketParticipants)
	case PacketCarSetups:
		decodeAndBroadcast(data, decodeCarSetupsPacket, "CarSetups", PacketCarSetups)
	case PacketCarTelemetry:
		telemetry := decodeCarTelemetryPacket(data)
		broadcastTelemetryFields(telemetry)
	case PacketCarStatus:
		decodeAndBroadcast(data, decodeCarStatusPacket, "CarStatus", PacketCarStatus)
	case PacketFinalClassification:
		decodeAndBroadcast(data, decodeFinalClassificationPacket, "FinalClassification", PacketFinalClassification)
	case PacketLobbyInfo:
		decodeAndBroadcast(data, decodeLobbyInfoPacket, "LobbyInfo", PacketLobbyInfo)
	case PacketCarDamage:
		decodeAndBroadcast(data, decodeCarDamagePacket, "CarDamage", PacketCarDamage)
	case PacketSessionHistory:
		decodeAndBroadcast(data, decodeSessionHistoryPacket, "SessionHistory", PacketSessionHistory)
	case PacketTyreSets:
		decodeAndBroadcast(data, decodeTyreSetsPacket, "TyreSets", PacketTyreSets)
	case PacketMotionEx:
		pkt, err := decodeMotionExPacket(data)
		if err != nil {
			log.Printf("[error] decodeMotionExPacket: %v", err)
			return
		}
		broadcastMotionExFields(pkt)
		// No JSON or forwardJSONToOSC here
	case PacketTimeTrial:
		decodeAndBroadcast(data, decodeTimeTrialPacket, "TimeTrial", PacketTimeTrial)
	case PacketLapPositions:
		decodeAndBroadcast(data, decodeLapPositionsPacket, "LapPositions", PacketLapPositions)
	}
}

func decodeCarTelemetryPacket(data []byte) CarTelemetryData {
	// Per F1 25 spec, header is 29 bytes, then 22 cars * 60 bytes each = 1352 bytes
	// Player car index is at offset 27 (m_playerCarIndex in header)
	if len(data) < 29+22*60 {
		log.Printf("[error] decodeCarTelemetryPacket: data too short (len=%d, need=%d)", len(data), 29+22*60)
		return CarTelemetryData{}
	}
	carIndex := int(data[27]) // m_playerCarIndex
	if carIndex < 0 || carIndex >= 22 {
		log.Printf("[error] decodeCarTelemetryPacket: invalid car index %d", carIndex)
		return CarTelemetryData{}
	}
	// Car telemetry data starts at offset 29
	carDataStart := 29 + carIndex*60
	carData := data[carDataStart : carDataStart+60]

	telemetry := CarTelemetryData{}
	telemetry.Speed = binary.LittleEndian.Uint16(carData[0:2])
	telemetry.Throttle = mathFromBits(carData[2:6])
	telemetry.Steer = mathFromBits(carData[6:10])
	telemetry.Brake = mathFromBits(carData[10:14])
	telemetry.Clutch = carData[14]
	telemetry.Gear = int8(carData[15])
	telemetry.EngineRPM = binary.LittleEndian.Uint16(carData[16:18])
	telemetry.DRS = carData[18]
	telemetry.RevLightsPercent = carData[19]
	telemetry.RevLightsBitValue = binary.LittleEndian.Uint16(carData[20:22])
	for i := 0; i < 4; i++ {
		telemetry.BrakesTemperature[i] = binary.LittleEndian.Uint16(carData[22+i*2 : 24+i*2])
	}
	for i := 0; i < 4; i++ {
		telemetry.TyresSurfaceTemperature[i] = carData[30+i]
	}
	for i := 0; i < 4; i++ {
		telemetry.TyresInnerTemperature[i] = carData[34+i]
	}
	telemetry.EngineTemperature = binary.LittleEndian.Uint16(carData[38:40])
	for i := 0; i < 4; i++ {
		telemetry.TyresPressure[i] = mathFromBits(carData[40+i*4 : 44+i*4])
	}
	for i := 0; i < 4; i++ {
		telemetry.SurfaceType[i] = carData[56+i]
	}
	return telemetry
}

func mathFromBits(b []byte) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(b))
}

func decodeMotionPacket(data []byte) (PacketMotionData, error) {
	var pkt PacketMotionData
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	if err != nil {
		log.Printf("[error] Failed to decode Motion Packet: %v", err)
	}

	return pkt, err
}

func decodeSessionPacket(data []byte) (PacketSessionData, error) {
	const expectedSize = 753
	var pkt PacketSessionData
	if len(data) < expectedSize {
		log.Printf("[error] PacketSessionData: data too short (got %d, want %d)", len(data), expectedSize)
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

func decodeLapDataPacket(data []byte) (LapDataPacket, error) {
	const expectedSize = 1285
	var pkt LapDataPacket
	if len(data) < expectedSize {
		log.Printf("[error] LapDataPacket: data too short (got %d, want %d)", len(data), expectedSize)
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

var lastEventShortLog int64

func decodeEventPacket(data []byte) (PacketEventData, error) {
	const expectedSize = 45
	var pkt PacketEventData
	if len(data) < expectedSize {
		if time.Now().Unix()-lastEventShortLog > 10 {
			// Add packet id and event string code to the error log
			var eventCode string
			if len(data) >= 11 {
				eventCode = string(data[9:13])
			}
			log.Printf("[error] decode Event: unexpected EOF | packetID=%d eventCode=%q (got %d, want %d)", PacketEvent, eventCode, len(data), expectedSize)
			lastEventShortLog = time.Now().Unix()
		}
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

func decodeParticipantsPacket(data []byte) (PacketParticipantsData, error) {
	const expectedSize = 1284
	var pkt PacketParticipantsData
	if len(data) < expectedSize {
		log.Printf("[error] PacketParticipantsData: data too short (got %d, want %d)", len(data), expectedSize)
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

func decodeCarSetupsPacket(data []byte) (PacketCarSetupData, error) {
	const expectedSize = 1133
	var pkt PacketCarSetupData
	if len(data) < expectedSize {
		log.Printf("[error] PacketCarSetupData: data too short (got %d, want %d)", len(data), expectedSize)
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

func decodeCarStatusPacket(data []byte) (PacketCarStatusData, error) {
	const expectedSize = 1239
	var pkt PacketCarStatusData
	if len(data) < expectedSize {
		log.Printf("[error] PacketCarStatusData: data too short (got %d, want %d)", len(data), expectedSize)
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

func decodeFinalClassificationPacket(data []byte) (PacketFinalClassificationData, error) {
	const expectedSize = 1042
	var pkt PacketFinalClassificationData
	if len(data) < expectedSize {
		log.Printf("[error] PacketFinalClassificationData: data too short (got %d, want %d)", len(data), expectedSize)
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

func decodeLobbyInfoPacket(data []byte) (PacketLobbyInfoData, error) {
	const expectedSize = 954
	var pkt PacketLobbyInfoData
	if len(data) < expectedSize {
		log.Printf("[error] PacketLobbyInfoData: data too short (got %d, want %d)", len(data), expectedSize)
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

func decodeCarDamagePacket(data []byte) (PacketCarDamageData, error) {
	const expectedSize = 1041
	var pkt PacketCarDamageData
	if len(data) < expectedSize {
		log.Printf("[error] PacketCarDamageData: data too short (got %d, want %d)", len(data), expectedSize)
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

func decodeSessionHistoryPacket(data []byte) (PacketSessionHistoryData, error) {
	const expectedSize = 1460
	var pkt PacketSessionHistoryData
	if len(data) < expectedSize {
		log.Printf("[error] PacketSessionHistoryData: data too short (got %d, want %d)", len(data), expectedSize)
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

func decodeTyreSetsPacket(data []byte) (PacketTyreSetsData, error) {
	const expectedSize = 231
	var pkt PacketTyreSetsData
	if len(data) < expectedSize {
		log.Printf("[error] PacketTyreSetsData: data too short (got %d, want %d)", len(data), expectedSize)
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

func decodeMotionExPacket(data []byte) (PacketMotionExData, error) {
	const expectedSize = 273
	var pkt PacketMotionExData
	if len(data) < expectedSize {
		log.Printf("[error] PacketMotionExData: data too short (got %d, want %d)", len(data), expectedSize)
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

func decodeTimeTrialPacket(data []byte) (PacketTimeTrialData, error) {
	const expectedSize = 101
	var pkt PacketTimeTrialData
	if len(data) < expectedSize {
		log.Printf("[error] PacketTimeTrialData: data too short (got %d, want %d)", len(data), expectedSize)
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

func decodeLapPositionsPacket(data []byte) (PacketLapPositionsData, error) {
	const expectedSize = 1131
	var pkt PacketLapPositionsData
	if len(data) < expectedSize {
		log.Printf("[error] PacketLapPositionsData: data too short (got %d, want %d)", len(data), expectedSize)
		return pkt, io.ErrUnexpectedEOF
	}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &pkt)
	return pkt, err
}

func broadcastMotionExFields(pkt PacketMotionExData) {
	wheels := []string{"RL", "RR", "FL", "FR"}
	fields := []struct {
		name   string
		values [4]float32
	}{
		{"WheelSpeed", pkt.WheelSpeed},
		{"WheelSlipRatio", pkt.WheelSlipRatio},
		{"WheelSlipAngle", pkt.WheelSlipAngle},
		{"WheelLatForce", pkt.WheelLatForce},
		{"WheelLongForce", pkt.WheelLongForce},
		{"WheelVertForce", pkt.WheelVertForce},
		{"WheelCamber", pkt.WheelCamber},
		{"WheelCamberGain", pkt.WheelCamberGain},
	}
	for _, field := range fields {
		for i, wheel := range wheels {
			key := field.name + wheel
			entry, ok := OSCAddresses[key]
			if !ok {
				continue
			}
			// Broadcast as plain text, not JSON
			msg := fmt.Sprintf("MotionEx/%s %v", key, field.values[i])
			broadcast([]byte(msg))
			if Config.EnableOSC && entry.Enabled {
				sendOSC(entry.Address, field.values[i])
			}
		}
	}
}

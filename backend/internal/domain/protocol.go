package domain

// Commands RPI -> ESP32
const (
	CmdMove      uint8 = 0x01
	CmdStop      uint8 = 0x02
	CmdEnable    uint8 = 0x03
	CmdGetStatus uint8 = 0x04
)

// Responces ESP32 → RPi
const (
	CmdStatus uint8 = 0x10
)

// Package sizes in bytes
const (
	PacketRegister  = 2  // motor_id + checksum
	PacketMove      = 10 // cmd + steps(4) + hz(4) + checksum
	PacketStop      = 2  // cmd + checksum
	PacketEnable    = 3  // cmd + enabled(1) + checksum
	PacketGetStatus = 2  // cmd + checksum
	PacketStatus    = 7  // cmd + enabled(1) + steps(4) + checksum
)

// MovePacket — RPi → ESP32
type MovePacket struct {
	Cmd      uint8
	Steps    int32
	Hz       float32
	Checksum uint8
}

// StatusPacket — ESP32 → RPi
type StatusPacket struct {
	Cmd      uint8
	Enabled  uint8
	Steps    int32
	Checksum uint8
}

// RegisterPacket — ESP32 → RPi on connect
type RegisterPacket struct {
	MotorID  uint8
	Checksum uint8
}

func Checksum(data []byte) uint8 {
	var x uint8
	for _, b := range data {
		x ^= b
	}
	return x
}

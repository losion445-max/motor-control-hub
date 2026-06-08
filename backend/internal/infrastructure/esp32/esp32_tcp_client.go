package esp32

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type EspMotorTCPClient struct {
	motorID int
	conn    net.Conn
	mu      sync.Mutex

	lastStatus domain.MotorStatus
	statusMu   sync.RWMutex
}

func NewTCPClient(motorID int, conn net.Conn) *EspMotorTCPClient {
	return &EspMotorTCPClient{
		motorID: motorID,
		conn:    conn,
	}
}

func (c *EspMotorTCPClient) ReadLoop(ctx context.Context) {
	buf := make([]byte, domain.PacketStatus)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if _, err := readFull(c.conn, buf); err != nil {
			log.Printf("[TCP] Motor %d read error: %v", c.motorID, err)
			return
		}

		if buf[0] != domain.CmdStatus {
			log.Printf("[TCP] Motor %d unexpected cmd: 0x%02x", c.motorID, buf[0])
			continue
		}

		// валидация checksum
		if domain.Checksum(buf[:6]) != buf[6] {
			log.Printf("[TCP] Motor %d invalid checksum in status", c.motorID)
			continue
		}

		steps := int32(binary.LittleEndian.Uint32(buf[2:6]))

		c.statusMu.Lock()
		c.lastStatus = domain.MotorStatus{
			MotorID:      c.motorID,
			Enabled:      buf[1] == 1,
			CurrentSteps: int(steps),
		}
		c.statusMu.Unlock()
	}
}

func (c *EspMotorTCPClient) Move(ctx context.Context, steps int, speedHz float64) error {
	buf := make([]byte, domain.PacketMove)
	buf[0] = domain.CmdMove
	copy(buf[1:5], int32ToBytes(int32(-steps)))
	copy(buf[5:9], float32ToBytes(float32(speedHz)))
	buf[9] = domain.Checksum(buf[:9])

	return c.send(buf)
}

func (c *EspMotorTCPClient) Stop(ctx context.Context) error {
	buf := make([]byte, domain.PacketStop)
	buf[0] = domain.CmdStop
	buf[1] = domain.Checksum(buf[:1])

	return c.send(buf)
}

func (c *EspMotorTCPClient) SetEnabled(ctx context.Context, enabled bool) error {
	buf := make([]byte, domain.PacketEnable)
	buf[0] = domain.CmdEnable
	if enabled {
		buf[1] = 1
	}
	buf[2] = domain.Checksum(buf[:2])

	return c.send(buf)
}

func (c *EspMotorTCPClient) GetStatus(ctx context.Context) (*domain.MotorStatus, error) {
	// запрашиваем статус
	buf := make([]byte, domain.PacketGetStatus)
	buf[0] = domain.CmdGetStatus
	buf[1] = domain.Checksum(buf[:1])

	if err := c.send(buf); err != nil {
		return nil, err
	}

	// возвращаем последний известный статус
	c.statusMu.RLock()
	status := c.lastStatus
	c.statusMu.RUnlock()

	return &status, nil
}

func (c *EspMotorTCPClient) GetConfig(ctx context.Context) (*domain.MotorConfig, error) {
	return &domain.MotorConfig{
		MotorID: c.motorID,
	}, nil
}

func (c *EspMotorTCPClient) send(buf []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.conn.Write(buf)
	if err != nil {
		return fmt.Errorf("motor %d send error: %w", c.motorID, err)
	}
	return nil
}

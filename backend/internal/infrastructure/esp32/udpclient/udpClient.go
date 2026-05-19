package udpclient

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type MotorUDPFactory struct{}

func NewMotorUDPFactory() domain.IMotorFactory {
	return &MotorUDPFactory{}
}

func (f *MotorUDPFactory) CreateMotor(cfg *domain.MotorConfig) (domain.IMotor, error) {
	client := NewMotorUDPClient(cfg)
	if client == nil {
		return nil, fmt.Errorf("failed to create UDP client for motor %d", cfg.MotorID)
	}
	return client, nil
}

type StreamPacket struct {
	Magic      uint8   // 0x5A
	MotorID    uint8   // 1..4
	StateCmd   uint8   // 0 - IDLE, 1 - STREAMING, 2 - ESTOP
	Seq        uint8   // index of package
	TargetStep int32   // 4 bytes (LittleEndian)
	SpeedRPS   float32 // 4 bytes (LittleEndian)
}

type MotorUDPClient struct {
	config     *domain.MotorConfig
	conn       *net.UDPConn
	targetAddr *net.UDPAddr
	closeChan  chan struct{}

	mu         sync.Mutex
	targetStep int32
	speedRPS   float32
	stateCmd   uint8
	seq        uint8
	enabled    bool
}

func NewMotorUDPClient(c *domain.MotorConfig) *MotorUDPClient {
	espAddr := fmt.Sprintf("%s:8888", c.CurrentIP)
	remoteAddr, err := net.ResolveUDPAddr("udp", espAddr)
	if err != nil {
		log.Printf("[UDP-CONFIG] Ошибка резолва адреса ESP32 %s: %v", espAddr, err)
		return nil
	}

	localAddr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		log.Printf("[UDP-CONFIG] Ошибка открытия локального сокета: %v", err)
		return nil
	}

	client := &MotorUDPClient{
		config:     c,
		conn:       conn,
		targetAddr: remoteAddr,
		closeChan:  make(chan struct{}),
		stateCmd:   0,
		enabled:    false,
		speedRPS:   5.0,
	}

	go client.startStreamingLoop(10 * time.Millisecond)

	return client
}

func (c *MotorUDPClient) startStreamingLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer c.conn.Close()

	for {
		select {
		case <-c.closeChan:
			return
		case <-ticker.C:
			c.mu.Lock()
			c.seq++

			packet := StreamPacket{
				Magic:      0x5A,
				MotorID:    uint8(c.config.MotorID),
				StateCmd:   c.stateCmd,
				Seq:        c.seq,
				TargetStep: c.targetStep,
				SpeedRPS:   c.speedRPS,
			}
			c.mu.Unlock()

			var buf bytes.Buffer
			if err := binary.Write(&buf, binary.LittleEndian, packet); err != nil {
				log.Printf("[UDP-MOTOR-%d] Ошибка сборки бинарного пакета: %v", c.config.MotorID, err)
				continue
			}

			_, err := c.conn.WriteToUDP(buf.Bytes(), c.targetAddr)
			if err != nil {
				log.Printf("[UDP-MOTOR-%d] Ошибка отправки пакета на %s: %v", c.config.MotorID, c.targetAddr, err)
			}
		}
	}
}

func (c *MotorUDPClient) Close() {
	close(c.closeChan)
}

func (c *MotorUDPClient) Move(ctx context.Context, steps int, speed float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.targetStep = int32(steps)
	c.speedRPS = float32(speed)
	c.stateCmd = 1 // Coordinate streaming mode

	return nil
}

func (c *MotorUDPClient) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stateCmd = 0 // stop // IDLE
	return nil
}

func (c *MotorUDPClient) SetEnabled(ctx context.Context, enabled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.enabled = enabled
	if !enabled {
		c.stateCmd = 0
	}
	return nil
}

func (c *MotorUDPClient) GetStatus(ctx context.Context) (*domain.MotorStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return &domain.MotorStatus{
		MotorID:      c.config.MotorID,
		Enabled:      c.enabled,
		Infinite:     false,
		CurrentSteps: int(c.targetStep),
		TargetSteps:  int(c.targetStep),
		SpeedRPS:     float64(c.speedRPS),
		RSSi:         -55,
	}, nil
}

func (c *MotorUDPClient) GetConfig(ctx context.Context) (*domain.MotorConfig, error) {
	return c.config, nil
}

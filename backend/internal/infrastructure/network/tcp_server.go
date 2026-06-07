package network

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/losion445-max/motor-control-hub/internal/domain"
	"github.com/losion445-max/motor-control-hub/internal/infrastructure/esp32"
)

type TCPServer struct {
	addr     string
	registry *MotorRegistry
}

func NewTCPServer(addr string, registry *MotorRegistry) *TCPServer {
	return &TCPServer{addr: addr, registry: registry}
}

func (s *TCPServer) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("tcp listen: %w", err)
	}
	defer ln.Close()

	log.Printf("[TCP] Listening on %s", s.addr)

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Printf("[TCP] Accept error: %v", err)
				continue
			}
		}
		go s.handleConn(ctx, conn)
	}
}

func (s *TCPServer) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, domain.PacketRegister)
	if _, err := readFull(conn, buf); err != nil {
		log.Printf("[TCP] Failed to read register packet: %v", err)
		return
	}

	if domain.Checksum(buf[:1]) != buf[1] {
		log.Printf("[TCP] Invalid checksum in register packet")
		return
	}

	motorID := int(buf[0])
	if motorID < 1 || motorID > 4 {
		log.Printf("[TCP] Invalid motor_id: %d", motorID)
		return
	}

	log.Printf("[TCP] Motor %d connected from %s", motorID, conn.RemoteAddr())

	client := esp32.NewTCPClient(motorID, conn)
	s.registry.Register(motorID, client)
	defer s.registry.Unregister(motorID)

	client.ReadLoop(ctx)

	log.Printf("[TCP] Motor %d disconnected", motorID)
}

func readFull(conn net.Conn, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := conn.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

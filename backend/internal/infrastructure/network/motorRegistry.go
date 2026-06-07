package network

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

const (
	ScanInterval = 5 * time.Second
	MotorCount   = 4
)

type MotorRegistry struct {
	mu      sync.RWMutex
	motors  [MotorCount]domain.IMotor
	online  [MotorCount]bool
	scanner domain.MotorDiscover
	factory func(cfg domain.MotorConfig) domain.IMotor
}

func NewMotorRegistry(
	scanner domain.MotorDiscover,
	factory func(cfg domain.MotorConfig) domain.IMotor,
) *MotorRegistry {
	return &MotorRegistry{
		scanner: scanner,
		factory: factory,
	}
}

func (r *MotorRegistry) Run(ctx context.Context) {
	r.scan(ctx)

	ticker := time.NewTicker(ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.healthCheck(ctx)
		}
	}
}

func (r *MotorRegistry) AllOnline() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, on := range r.online {
		if !on {
			return false
		}
	}
	return true
}

func (r *MotorRegistry) Motors() [MotorCount]domain.IMotor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.motors
}

func (r *MotorRegistry) scan(ctx context.Context) {
	configs, err := r.scanner.Discover(ctx)
	if err != nil {
		log.Printf("[REGISTRY] Scan error: %v", err)
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, cfg := range configs {
		idx := cfg.MotorID - 1
		if idx < 0 || idx >= MotorCount {
			log.Printf("[REGISTRY] Invalid motor ID %d, skipping", cfg.MotorID)
			continue
		}
		r.motors[idx] = r.factory(cfg)
		r.online[idx] = true
		log.Printf("[REGISTRY] Motor %d connected (%s)", cfg.MotorID, cfg.CurrentIP)
	}
}

func (r *MotorRegistry) healthCheck(ctx context.Context) {
	r.mu.RLock()
	motors := r.motors
	online := r.online
	r.mu.RUnlock()

	for i := 0; i < MotorCount; i++ {
		if motors[i] == nil {
			r.scan(ctx)
			return
		}

		if online[i] {
			_, err := motors[i].GetStatus(ctx)
			if err != nil {
				r.mu.Lock()
				r.online[i] = false
				r.mu.Unlock()
				log.Printf("[REGISTRY] Motor %d went offline: %v", i+1, err)
			}
		} else {
			r.scan(ctx)
			return
		}
	}
}

func (r *MotorRegistry) Register(motorID int, motor domain.IMotor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	idx := motorID - 1
	r.motors[idx] = motor
	r.online[idx] = true
	log.Printf("[REGISTRY] Motor %d registered via TCP", motorID)
}

func (r *MotorRegistry) Unregister(motorID int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	idx := motorID - 1
	r.motors[idx] = nil
	r.online[idx] = false
	log.Printf("[REGISTRY] Motor %d unregistered", motorID)
}

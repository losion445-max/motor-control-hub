package usecase

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

const MotorCount = 0

type MotorFactory func(cfg *domain.MotorConfig) domain.IMotor

func BootstrapMotors(ctx context.Context, scanner domain.MotorDiscover, factory MotorFactory) ([]domain.IMotor, error) {
	log.Println("[BOOTSTRAP] Looking for motors...")

	configs, err := scanner.Discover(ctx)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	if len(configs) != MotorCount {
		return nil, fmt.Errorf("critical error: expected %d motors, found %d", MotorCount, len(configs))
	}

	sort.Slice(configs, func(i, j int) bool {
		return configs[i].MotorID < configs[j].MotorID
	})

	motors := make([]domain.IMotor, 0, MotorCount)
	for i, cfg := range configs {
		expectedID := i + 1
		if cfg.MotorID != expectedID {
			return nil, fmt.Errorf("sequence error: expected ID %d, but got %d at position %d", expectedID, cfg.MotorID, i)
		}

		motors = append(motors, factory(&cfg))
	}

	log.Printf("[BOOTSTRAP] Successfully initialized %d motors", len(motors))
	return motors, nil
}

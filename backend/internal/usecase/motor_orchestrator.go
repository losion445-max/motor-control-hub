package usecase

import (
	"context"
	"log"
	"sync"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type MotorOrchestrator struct {
	motors     []domain.IMotor
	kinematics *KinematicsService
}

func NewMotorOrchestrator(motors []domain.IMotor, kinematics *KinematicsService) *MotorOrchestrator {
	if len(motors) != 4 {
		log.Panicf("Connected motors must be 4!")
	}

	return &MotorOrchestrator{
		motors:     motors,
		kinematics: kinematics,
	}
}

func (m *MotorOrchestrator) GetAllAggregatedConfig(ctx context.Context) ([]domain.MotorConfig, error) {
	configs := make([]domain.MotorConfig, len(m.motors))
	for i, config := range m.motors {
		configs[i] = config.GetConfig()
	}

	return configs, nil
}

func (m *MotorOrchestrator) GetAllAggregatedStatus(ctx context.Context) ([]*domain.MotorStatus, error) {
	statuses := make([]*domain.MotorStatus, len(m.motors))

	errors := make(chan error, len(m.motors))
	var wg sync.WaitGroup

	for i, motor := range m.motors {
		wg.Add(1)
		go func(idx int, m domain.IMotor) {
			defer wg.Done()

			status, err := m.GetStatus(ctx)
			if err != nil {
				errors <- err
				return
			}
			statuses[idx] = status
		}(i, motor)
	}

	wg.Wait()
	close(errors)
	if len(errors) > 0 {
		return nil, <-errors
	}

	return statuses, nil
}

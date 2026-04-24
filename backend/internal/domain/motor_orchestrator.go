package domain

import (
	"context"
	"log"
	"sync"
)

type MotorOrchestrator struct {
	motors []IMotor
}

func NewMotorOrchestrator(motors []IMotor) *MotorOrchestrator {
	if len(motors) != 4 {
		log.Panicf("Connected motors must be 4!")
	}

	return &MotorOrchestrator{
		motors: motors,
	}
}

func (m *MotorOrchestrator) GetAllAggregatedConfig(ctx context.Context) ([]MotorConfig, error) {
	configs := make([]MotorConfig, len(m.motors))
	for i, config := range m.motors {
		configs[i] = config.GetConfig()
	}

	return configs, nil
}

func (m *MotorOrchestrator) GetAllAggregatedStatus(ctx context.Context) ([]*MotorStatus, error) {
	statuses := make([]*MotorStatus, len(m.motors))

	errors := make(chan error, len(m.motors))
	var wg sync.WaitGroup

	for i, motor := range m.motors {
		wg.Add(1)
		go func(idx int, m IMotor) {
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

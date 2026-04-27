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

func (m *MotorOrchestrator) GoHome(ctx context.Context, speed float64) error {
	return m.kinematics.GoHome(ctx, speed)
}

func (m *MotorOrchestrator) GetFrameHeight() float64 {
	return m.kinematics.Height
}

func (m *MotorOrchestrator) GetFrameWidth() float64 {
	return m.kinematics.Width
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

func (m *MotorOrchestrator) MoveToPoint(ctx context.Context, x, y, speed float64) error {
	return m.kinematics.MoveTo(ctx, domain.Point{X: x, Y: y}, speed)
}

func (m *MotorOrchestrator) EmergencyStop(ctx context.Context) error {
	return m.kinematics.StopAll(ctx)
}

func (m *MotorOrchestrator) Calibrate(ctx context.Context, speed float64) error {
	return m.kinematics.Calibrate(ctx, speed)
}

// internal/usecase/orchestrator.go

func (m *MotorOrchestrator) GetCurrentPosition() domain.Point {
	return m.kinematics.currentPosition
}

func (m *MotorOrchestrator) IsCalibrated() bool {
	// Logic to check if Calibrate() has been called successfully
	return m.kinematics.currentPosition != (domain.Point{})
}

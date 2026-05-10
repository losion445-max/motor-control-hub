package usecase

import (
	"context"
	"log"
	"sync"

	"github.com/losion445-max/motor-control-hub/internal/config"
	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type MotorOrchestrator struct {
	motors     []domain.IMotor
	kinematics *KinematicsService
}

func NewMotorOrchestrator(motors []domain.IMotor, kinematics *KinematicsService) *MotorOrchestrator {
	if len(motors) != MotorCount {
		log.Panicf("[ORCHESTRATOR] Connected motors must be exactly 4!")
	}

	return &MotorOrchestrator{
		motors:     motors,
		kinematics: kinematics,
	}
}

func (m *MotorOrchestrator) GetAllAggregatedConfig(ctx context.Context) ([]domain.MotorConfig, error) {
	configs := make([]domain.MotorConfig, len(m.motors))
	for i, motor := range m.motors {
		conf, err := motor.GetConfig(ctx)
		if err != nil {
			log.Printf("[ORCHESTRATOR] Failed to get config from motor %d: %v", i, err)
			configs[i] = domain.MotorConfig{}
			continue
		}
		configs[i] = *conf
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

func (m *MotorOrchestrator) GetCurrentPosition() domain.Point {
	m.kinematics.mu.RLock()
	defer m.kinematics.mu.RUnlock()
	return m.kinematics.currentPosition
}

func (m *MotorOrchestrator) GoHome(ctx context.Context, speed float64) error {
	return m.kinematics.MoveTo(ctx, domain.Point{X: 0, Y: 0}, speed)
}

func (m *MotorOrchestrator) GetConfig() *config.GlobalConfig {
	return m.kinematics.cfg
}

func (m *MotorOrchestrator) Sync() {
	m.kinematics.SyncConfig()
}

func (m *MotorOrchestrator) IsCalibrated() bool {
	return m.kinematics.currentPosition != (domain.Point{})
}

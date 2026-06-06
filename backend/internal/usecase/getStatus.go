package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type MotorStatusResult struct {
	MotorID int                 `json:"motor_id"`
	Status  *domain.MotorStatus `json:"status"`
	Err     string              `json:"error,omitempty"`
}

type StatusResult struct {
	Timestamp    int64                `json:"timestamp"`
	Position     domain.Point2D       `json:"position"`
	IsCalibrated bool                 `json:"is_calibrated"`
	Motors       [4]MotorStatusResult `json:"motors"`
}

type GetStatus struct {
	registry   domain.IMotorRegistry
	kinematics domain.IKinematicsController
}

func NewGetStatus(registry domain.IMotorRegistry, kinematics domain.IKinematicsController) *GetStatus {
	return &GetStatus{registry: registry, kinematics: kinematics}
}

func (uc *GetStatus) Execute(ctx context.Context) (StatusResult, error) {
	motors := uc.registry.Motors()
	var wg sync.WaitGroup
	var results [4]MotorStatusResult

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if motors[idx] == nil {
				results[idx] = MotorStatusResult{MotorID: idx + 1, Err: "offline"}
				return
			}
			status, err := motors[idx].GetStatus(ctx)
			if err != nil {
				results[idx] = MotorStatusResult{MotorID: idx + 1, Err: err.Error()}
				return
			}
			results[idx] = MotorStatusResult{MotorID: idx + 1, Status: status}
		}(i)
	}

	wg.Wait()
	pos := uc.kinematics.CurrentPosition()
	return StatusResult{
		Timestamp:    time.Now().UnixMilli(),
		Position:     pos,
		IsCalibrated: pos != (domain.Point2D{}),
		Motors:       results,
	}, nil
}

package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type MotorStatusResult struct {
	MotorID int
	Status  *domain.MotorStatus
	Err     error
}

type StatusResult struct {
	Timestamp    int64                `json:"timestamp"`
	Position     domain.Point2D       `json:"position"`
	IsCalibrated bool                 `json:"is_calibrated"`
	Motors       [4]MotorStatusResult `json:"motors"`
}

type GetStatus struct {
	motors     [4]domain.IMotor
	kinematics domain.IKinematicsController
}

func NewGetStatus(motors [4]domain.IMotor, kinematics domain.IKinematicsController) *GetStatus {
	return &GetStatus{motors: motors, kinematics: kinematics}
}

func (uc *GetStatus) Execute(ctx context.Context) (StatusResult, error) {
	var (
		wg      sync.WaitGroup
		results [4]MotorStatusResult
	)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			status, err := uc.motors[idx].GetStatus(ctx)
			results[idx] = MotorStatusResult{
				MotorID: idx + 1,
				Status:  status,
				Err:     err,
			}
		}(i)
	}
	wg.Wait()

	var firstErr error
	for _, r := range results {
		if r.Err != nil {
			firstErr = r.Err
			break
		}
	}

	pos := uc.kinematics.CurrentPosition()

	return StatusResult{
		Timestamp:    time.Now().UnixMilli(),
		Position:     pos,
		IsCalibrated: pos != (domain.Point2D{}),
		Motors:       results,
	}, firstErr
}

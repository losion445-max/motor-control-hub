package usecase

import (
	"context"
	"fmt"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type MoveSingleMotor struct {
	motors [4]domain.IMotor
}

func NewMoveSingleMotor(motors [4]domain.IMotor) *MoveSingleMotor {
	return &MoveSingleMotor{motors: motors}
}

func (uc *MoveSingleMotor) Execute(ctx context.Context, motorID int, steps int, speedHz float64) error {
	idx := motorID - 1
	if idx < 0 || idx >= 4 {
		return fmt.Errorf("invalid motor id %d, must be 1-4", motorID)
	}
	return uc.motors[idx].Move(ctx, steps, speedHz)
}

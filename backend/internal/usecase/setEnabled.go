package usecase

import (
	"context"
	"fmt"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type SetEnabled struct {
	motors [4]domain.IMotor
}

func NewSetEnabled(motors [4]domain.IMotor) *SetEnabled {
	return &SetEnabled{motors: motors}
}

func (uc *SetEnabled) Execute(ctx context.Context, enabled bool) error {
	for i, motor := range uc.motors {
		if err := motor.SetEnabled(ctx, enabled); err != nil {
			return fmt.Errorf("motor %d: %w", i+1, err)
		}
	}
	return nil
}

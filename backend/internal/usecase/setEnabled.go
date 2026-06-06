package usecase

import (
	"context"
	"fmt"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type SetEnabled struct {
	registry domain.IMotorRegistry
}

func NewSetEnabled(registry domain.IMotorRegistry) *SetEnabled {
	return &SetEnabled{registry: registry}
}

func (uc *SetEnabled) Execute(ctx context.Context, enabled bool) error {
	motors := uc.registry.Motors()
	for i, motor := range motors {
		if motor == nil {
			return fmt.Errorf("motor %d offline", i+1)
		}
		if err := motor.SetEnabled(ctx, enabled); err != nil {
			return fmt.Errorf("motor %d: %w", i+1, err)
		}
	}
	return nil
}

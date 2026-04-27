package usecase

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/losion445-max/motor-control-hub/internal/domain"
	"github.com/losion445-max/motor-control-hub/internal/infrastructure/esp32"
)

func BootstrapMotors(ctx context.Context, scanner domain.MotorDiscover) ([]domain.IMotor, error) {
	log.Println("[BOOTSTRAP] Looking for motors...")
	configs, err := scanner.Discover(ctx)
	if err != nil {
		return nil, err
	}

	if len(configs) != 4 {
		return nil, fmt.Errorf("[BOOTSTRAP] critical error: expected 4 motors, found %d", len(configs))
	}

	var motors []domain.IMotor
	for _, cfg := range configs {
		motors = append(motors, esp32.NewMotorClient(cfg))
	}

	slices.SortFunc(motors, func(a, b domain.IMotor) int {
		return a.GetConfig().MotorID - b.GetConfig().MotorID
	})

	for i, m := range motors {
		if m.GetConfig().MotorID != (i + 1) {
			return nil, fmt.Errorf("[BOOTSTRAP] sequence error: expected motor ID %d at index %d, but got ID %d", i, i, m.GetConfig().MotorID)
		}
	}

	log.Println("[BOOTStRAP] Motors were successfully initialized and mapped to corners!")
	return motors, nil
}

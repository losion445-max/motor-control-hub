package usecase

import (
	"context"
	"log"

	"github.com/losion445-max/motor-control-hub/internal/domain"
	"github.com/losion445-max/motor-control-hub/internal/infrastructure/esp32"
)

func BootstrapMotors(ctx context.Context, scanner domain.MotorDiscover) ([]domain.IMotor, error) {
	log.Println("[BOOTSTRAP] Looking for motors...")
	configs, err := scanner.Discover(ctx)
	if err != nil {
		return nil, err
	}

	log.Printf("[BOOTSTRAP] Found only %d motors!", len(configs))

	var motors []domain.IMotor
	for _, cfg := range configs {
		motors = append(motors, esp32.NewMotorClient(cfg))
	}

	log.Println("[BOOTStRAP] Motors were successfully initialized!")
	return motors, nil
}

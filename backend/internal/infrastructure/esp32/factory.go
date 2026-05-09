package esp32

import "github.com/losion445-max/motor-control-hub/internal/domain"

func NewMotorClientFactory() func(cfg *domain.MotorConfig) domain.IMotor {
	return func(cfg *domain.MotorConfig) domain.IMotor {
		return NewMotorClient(cfg)
	}
}

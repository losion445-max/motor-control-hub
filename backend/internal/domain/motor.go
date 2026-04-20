package domain

import "context"

type MotorConfig struct {
	MotorID     int     `json:"motor_id"`
	StepPlus    int     `json:"step_plus"`
	StepMinus   int     `json:"step_minus"`
	DirPlus     int     `json:"dir_plus"`
	DirMinus    int     `json:"dir_minus"`
	StepsPerRev int     `json:"steps_per_rev"`
	PulleyMM    float64 `json:"pulley_mm"`
	CurrentIP   string  `json:"-"`
}

type MotorStatus struct {
	MotorID      int     `json:"motor_id"`
	Enabled      bool    `json:"enabled"`
	Infinite     bool    `json:"infinite"`
	CurrentSteps int     `json:"current_steps"`
	TargetSteps  int     `json:"target_steps"`
	SpeedRPS     float64 `json:"speed_rps"`
	RSSi         int     `json:"wifi_rssi"`
}

type IMotor interface {
	GetConfig() MotorConfig
	Move(ctx context.Context, steps int, speed float64) error
	Stop(ctx context.Context) error
	GetStatus(ctx context.Context) (*MotorStatus, error)
}

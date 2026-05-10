package config

type GlobalConfig struct {
	Kinematics struct {
		Width       float64 `json:"width"`
		Height      float64 `json:"height"`
		Diameter    float64 `json:"diameter"`
		StepsPerRev int     `json:"steps_per_rev"`
	} `json:"kinematics"`

	// Mapping: which ID responsible for what angle
	// Indexes: 0-TL, 1-TR, 2-BR, 3-BL
	MotorMapping [4]int `json:"motor_mapping"`

	// Defaults for motors
	Defaults struct {
		StepsPerRev int     `json:"steps_per_rev"`
		PulleyMM    float64 `json:"pulley_mm"`
	} `json:"defaults"`
}

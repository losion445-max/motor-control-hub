package domain

import "context"

// Millimeters
type Point2D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type MotorCommand struct {
	MotorID int
	Steps   int
	Hz      float64 // Times speed esp32
}

// Command for 4 motors in one tick
type Tick [4]MotorCommand

type WorkZone struct {
	Width  float64 // mm
	Height float64 // mm
}

// Split path in many pieces
type ITrajectoryPlanner interface {
	Plan(from, to Point2D, speedMmS float64) ([]Point2D, error)
}

// Motor commands for 1 ppoint
type IKinematicsController interface {
	Solve(next Point2D) (Tick, error)
	SetHome(pos Point2D)
	CurrentPosition() Point2D
	SetWorkZone(zone WorkZone)
}

// Put tick for all esp32
type IMotorDispatcher interface {
	Dispatch(ctx context.Context, ticks []Tick) error
	StopAll(ctx context.Context) error
}

type IMotorRegistry interface {
	Motors() [4]IMotor
	AllOnline() bool
}

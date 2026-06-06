package usecase

import (
	"context"
	"fmt"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type MoveTo struct {
	planner    domain.ITrajectoryPlanner
	kinematics domain.IKinematicsController
	dispatcher domain.IMotorDispatcher
	registry   domain.IMotorRegistry
}

func NewMoveTo(
	planner domain.ITrajectoryPlanner,
	kinematics domain.IKinematicsController,
	dispatcher domain.IMotorDispatcher,
	registry domain.IMotorRegistry,
) *MoveTo {
	return &MoveTo{
		planner:    planner,
		kinematics: kinematics,
		dispatcher: dispatcher,
		registry:   registry,
	}
}

func (uc *MoveTo) Execute(ctx context.Context, target domain.Point2D, speedMmS float64) error {
	if !uc.registry.AllOnline() {
		return fmt.Errorf("not all motors online, movement rejected")
	}
	from := uc.kinematics.CurrentPosition()
	points, err := uc.planner.Plan(from, target, speedMmS)
	if err != nil {
		return fmt.Errorf("plan: %w", err)
	}
	if len(points) == 0 {
		return nil
	}
	ticks := make([]domain.Tick, 0, len(points))
	for _, point := range points {
		tick, err := uc.kinematics.Solve(point)
		if err != nil {
			return fmt.Errorf("solve point (%.2f, %.2f): %w", point.X, point.Y, err)
		}
		ticks = append(ticks, tick)
	}
	return uc.dispatcher.Dispatch(ctx, ticks)
}

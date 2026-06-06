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
}

func NewMoveTo(
	planner domain.ITrajectoryPlanner,
	kinematics domain.IKinematicsController,
	dispatcher domain.IMotorDispatcher,
) *MoveTo {
	return &MoveTo{
		planner:    planner,
		kinematics: kinematics,
		dispatcher: dispatcher,
	}
}

func (uc *MoveTo) Execute(ctx context.Context, target domain.Point2D, speedMmS float64) error {
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

	if err := uc.dispatcher.Dispatch(ctx, ticks); err != nil {
		return fmt.Errorf("dispatch: %w", err)
	}

	return nil
}

package usecase

import (
	"context"
	"fmt"
	"log" // Добавлен стандартный логгер (или используйте ваш logger)

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
	log.Printf("[MoveTo] Execution started: Target(%.2f, %.2f), Speed: %.2f mm/s", target.X, target.Y, speedMmS)

	if !uc.registry.AllOnline() {
		log.Printf("[MoveTo] Error: Movement rejected, some motors are offline")
		return fmt.Errorf("not all motors online, movement rejected")
	}

	from := uc.kinematics.CurrentPosition()
	log.Printf("[MoveTo] Current position: (%.2f, %.2f)", from.X, from.Y)

	points, err := uc.planner.Plan(from, target, speedMmS)
	if err != nil {
		log.Printf("[MoveTo] Planner error: %v", err)
		return fmt.Errorf("plan: %w", err)
	}

	if len(points) == 0 {
		log.Printf("[MoveTo] Warning: Planner returned 0 points, movement skipped")
		return nil
	}

	log.Printf("[MoveTo] Path planned, calculated %d points", len(points))

	ticks := make([]domain.Tick, 0, len(points))
	for i, point := range points {
		tick, err := uc.kinematics.Solve(point)
		if err != nil {
			log.Printf("[MoveTo] Kinematics error at point %d (%.2f, %.2f): %v", i, point.X, point.Y, err)
			return fmt.Errorf("solve point (%.2f, %.2f): %w", point.X, point.Y, err)
		}
		ticks = append(ticks, tick)

		// Опционально: очень подробный лог для отладки кинематики (будьте осторожны с объемом!)
		// log.Printf("[MoveTo] Point %d solved: M1=%d, M2=%d, M3=%d, M4=%d", i, tick.M1, tick.M2, tick.M3, tick.M4)
	}

	log.Printf("[MoveTo] Dispatching %d ticks to motors...", len(ticks))
	err = uc.dispatcher.Dispatch(ctx, ticks)
	if err != nil {
		log.Printf("[MoveTo] Dispatcher error: %v", err)
		return fmt.Errorf("dispatch: %w", err)
	}

	log.Printf("[MoveTo] Execution successfully finished")
	return nil
}

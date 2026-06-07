package kinematics

import (
	"fmt"
	"log"
	"math"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

const (
	DefaultHz        = 50.0
	DefaultAccelMmS2 = 300.0
)

type TrajectoryPlanner struct {
	hz        float64
	accelMmS2 float64
}

func NewTrajectoryPlanner(hz, accelMmS2 float64) *TrajectoryPlanner {
	if hz <= 0 {
		hz = DefaultHz
	}
	if accelMmS2 <= 0 {
		accelMmS2 = DefaultAccelMmS2
	}
	log.Printf("[Planner] Initialized: Hz=%.2f, Accel=%.2f mm/s^2", hz, accelMmS2)
	return &TrajectoryPlanner{hz: hz, accelMmS2: accelMmS2}
}

func (p *TrajectoryPlanner) Plan(from, to domain.Point2D, speedMmS float64) ([]domain.Point2D, error) {
	if speedMmS <= 0 {
		log.Printf("[Planner] Error: Invalid speed %.2f", speedMmS)
		return nil, fmt.Errorf("speed must be positive, got %.2f", speedMmS)
	}

	dx := to.X - from.X
	dy := to.Y - from.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	log.Printf("[Planner] Planning path: (%.2f, %.2f) -> (%.2f, %.2f), Dist=%.2fmm, Speed=%.2fmm/s",
		from.X, from.Y, to.X, to.Y, distance, speedMmS)

	if distance < 0.1 {
		log.Printf("[Planner] Distance too small, skipping")
		return nil, nil
	}

	ux := dx / distance
	uy := dy / distance

	speeds := p.trapezoidSpeeds(distance, speedMmS)
	log.Printf("[Planner] Generated %d speed segments", len(speeds))

	points := make([]domain.Point2D, len(speeds))
	pos := 0.0
	dt := 1.0 / p.hz

	for i, v := range speeds {
		pos += v * dt
		if pos > distance {
			pos = distance
		}
		points[i] = domain.Point2D{
			X: from.X + ux*pos,
			Y: from.Y + uy*pos,
		}
	}

	log.Printf("[Planner] Successfully generated %d path points", len(points))
	return points, nil
}

func (p *TrajectoryPlanner) trapezoidSpeeds(distance, maxSpeed float64) []float64 {
	dt := 1.0 / p.hz
	a := p.accelMmS2

	accelDist := (maxSpeed * maxSpeed) / (2 * a)

	if accelDist*2 >= distance {
		accelDist = distance / 2
		maxSpeed = math.Sqrt(2 * a * accelDist)
		log.Printf("[Planner] Trapezoid adjusted: New MaxSpeed=%.2fmm/s, AccelDist=%.2fmm", maxSpeed, accelDist)
	}

	cruiseDist := distance - 2*accelDist
	log.Printf("[Planner] Trapezoid segments: Accel=%.2f, Cruise=%.2f, Decel=%.2f", accelDist, cruiseDist, accelDist)

	var speeds []float64
	pos := 0.0
	vel := 0.0

	for pos < distance-0.001 {
		switch {
		case pos < accelDist:
			vel = math.Min(vel+a*dt, maxSpeed)
		case pos < accelDist+cruiseDist:
			vel = maxSpeed
		default:
			vel = math.Max(vel-a*dt, 0.1)
		}

		remaining := distance - pos
		if remaining <= vel*dt {
			vel = remaining / dt
			pos += vel * dt
			speeds = append(speeds, vel)
			break
		}

		pos += vel * dt
		speeds = append(speeds, vel)
	}

	return speeds
}

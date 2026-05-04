package usecase

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

const MotorCount = 4

type KinematicsService struct {
	Width, Height float64
	Motors        []domain.IMotor

	mu                   sync.Mutex
	currentPosition      domain.Point
	currentAbsoluteSteps [MotorCount]int
}

func NewKinematicsService(width, height float64, motorInstances []domain.IMotor) (*KinematicsService, error) {
	if len(motorInstances) != MotorCount {
		return nil, fmt.Errorf("kinematics service requires exactly %d motors, got %d", MotorCount, len(motorInstances))
	}

	return &KinematicsService{
		Width:  width,
		Height: height,
		Motors: motorInstances,
	}, nil
}

func (s *KinematicsService) MoveTo(ctx context.Context, targetPos domain.Point, baseSpeed float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	target := s.clampPoint(targetPos)
	targetLengths := s.calculateIK(target)

	deltas := make([]int, MotorCount)
	targetSteps := make([]int, MotorCount)
	var maxDelta float64

	for i := 0; i < MotorCount; i++ {
		steps, err := s.mmToSteps(ctx, targetLengths[i], i)
		if err != nil {
			return fmt.Errorf("failed to calculate steps for motor %d: %w", i, err)
		}
		targetSteps[i] = steps
		deltas[i] = steps - s.currentAbsoluteSteps[i]

		absDelta := math.Abs(float64(deltas[i]))
		if absDelta > maxDelta {
			maxDelta = absDelta
		}
	}

	var wg sync.WaitGroup
	errChan := make(chan error, MotorCount)

	for i := 0; i < MotorCount; i++ {
		wg.Add(1)
		go func(idx int, delta int) {
			defer wg.Done()

			speed := baseSpeed
			if maxDelta > 0 {
				speed = baseSpeed * (math.Abs(float64(delta)) / maxDelta)
			}

			if err := s.Motors[idx].Move(ctx, delta, speed); err != nil {
				errChan <- fmt.Errorf("motor %d failed: %w", idx, err)
			}
		}(i, deltas[i])
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}

	s.currentAbsoluteSteps = [MotorCount]int(targetSteps)
	s.currentPosition = target
	return nil
}

func (s *KinematicsService) Calibrate(ctx context.Context, calibrationSpeed float64) error {
	s.mu.Lock()
	s.currentAbsoluteSteps = [MotorCount]int{}
	s.currentPosition = domain.Point{X: 0, Y: 0}
	s.mu.Unlock()

	return s.MoveTo(ctx, domain.Point{X: 0, Y: 0}, calibrationSpeed)
}

func (s *KinematicsService) StopAll(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, MotorCount)
	for i := 0; i < MotorCount; i++ {
		wg.Add(1)
		go func(m domain.IMotor) {
			defer wg.Done()
			if err := m.Stop(ctx); err != nil {
				errChan <- err
			}
		}(s.Motors[i])
	}
	wg.Wait()
	close(errChan)
	return <-errChan
}

func (s *KinematicsService) clampPoint(p domain.Point) domain.Point {
	return domain.Point{
		X: math.Max(0, math.Min(p.X, s.Width)),
		Y: math.Max(0, math.Min(p.Y, s.Height)),
	}
}

func (s *KinematicsService) mmToSteps(ctx context.Context, lengthMM float64, motorIdx int) (int, error) {
	config, err := s.Motors[motorIdx].GetConfig(ctx)
	if err != nil {
		return 0, err
	}
	steps := (lengthMM * float64(config.StepsPerRev)) / (config.PulleyMM * math.Pi)
	return int(math.Round(steps)), nil
}

func (s *KinematicsService) calculateIK(pos domain.Point) [MotorCount]float64 {
	x, y := pos.X, pos.Y
	return [MotorCount]float64{
		math.Sqrt(x*x + y*y),                                        // Top-Left
		math.Sqrt(math.Pow(s.Width-x, 2) + y*y),                     // Top-Right
		math.Sqrt(math.Pow(s.Width-x, 2) + math.Pow(s.Height-y, 2)), // Bottom-Right
		math.Sqrt(x*x + math.Pow(s.Height-y, 2)),                    // Bottom-Left
	}
}

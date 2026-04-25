package usecase

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type KinematicsService struct {
	Width, Height float64
	Motors        []domain.IMotor

	currentPosition domain.Point

	currentAbsoluteSteps [4]int

	motorConfigs [4]domain.MotorConfig
}

func NewKinematicsService(width, height float64, motorInstances []domain.IMotor) (*KinematicsService, error) {
	if len(motorInstances) != 4 {
		return nil, fmt.Errorf("kinematics service requires exactly 4 motor instances, got %d", len(motorInstances))
	}

	s := &KinematicsService{
		Width:  width,
		Height: height,
		Motors: motorInstances,
	}

	for i := 0; i < 4; i++ {
		config := motorInstances[i].GetConfig()
		s.motorConfigs[i] = config
	}

	return s, nil
}

func (s *KinematicsService) Calibrate(ctx context.Context, calibrationSpeed float64) error {
	targetLengthsAtHome := s.calculateIK(domain.Point{X: 0, Y: 0})

	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(motorIdx int, motor domain.IMotor, targetL float64) {
			defer wg.Done()

			stepsToMove := s.mmToSteps(targetL, motorIdx)

			err := motor.Move(ctx, stepsToMove, calibrationSpeed)
			if err != nil {
				errChan <- fmt.Errorf("motor %d calibration failed: %w", motorIdx, err)
				return
			}
			s.currentAbsoluteSteps[motorIdx] = stepsToMove
		}(i, s.Motors[i], targetLengthsAtHome[i])
	}
	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	s.currentPosition = domain.Point{X: 0, Y: 0}
	fmt.Println("KinematicsService: Calibration at (0,0) complete.")
	return nil
}

func (s *KinematicsService) MoveTo(ctx context.Context, targetPos domain.Point, baseSpeed float64) error {
	targetX := math.Max(0, math.Min(targetPos.X, s.Width))
	targetY := math.Max(0, math.Min(targetPos.Y, s.Height))
	clampedTargetPos := domain.Point{X: targetX, Y: targetY}

	targetLengths := s.calculateIK(clampedTargetPos)

	var wg sync.WaitGroup
	errChan := make(chan error, 4)
	maxDeltaSteps := 0

	motorStepsToMove := make([]int, 4)
	for i := 0; i < 4; i++ {
		absoluteTargetSteps := s.mmToSteps(targetLengths[i], i)

		deltaSteps := absoluteTargetSteps - s.currentAbsoluteSteps[i]
		motorStepsToMove[i] = deltaSteps

		if math.Abs(float64(deltaSteps)) > float64(maxDeltaSteps) {
			maxDeltaSteps = int(math.Abs(float64(deltaSteps)))
		}
	}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(motorIdx int, motor domain.IMotor, deltaSteps int) {
			defer wg.Done()

			motorSpeed := baseSpeed
			if maxDeltaSteps > 0 {
				motorSpeed = baseSpeed * (math.Abs(float64(deltaSteps)) / float64(maxDeltaSteps))
			}

			err := motor.Move(ctx, deltaSteps, motorSpeed)
			if err != nil {
				errChan <- fmt.Errorf("motor %d move failed: %w", motorIdx, err)
				return
			}
			s.currentAbsoluteSteps[motorIdx] += deltaSteps
		}(i, s.Motors[i], motorStepsToMove[i])
	}
	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	s.currentPosition = clampedTargetPos
	fmt.Printf("KinematicsService: Moved to (%.2f, %.2f) complete.", clampedTargetPos.X, clampedTargetPos.Y)
	return nil
}

func (s *KinematicsService) GoHome(ctx context.Context, speed float64) error {
	fmt.Println("KinematicsService: Going to home position (0,0)...")
	return s.MoveTo(ctx, domain.Point{X: 0, Y: 0}, speed)
}

func (s *KinematicsService) StopAll(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(motorIdx int, m domain.IMotor) {
			defer wg.Done()
			if err := m.Stop(ctx); err != nil {
				errChan <- fmt.Errorf("motor %d stop failed: %w", motorIdx, err)
			}
		}(i, s.Motors[i])
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	fmt.Println("KinematicsService: All motors stopped.")
	return nil
}

func (s *KinematicsService) mmToSteps(lengthMM float64, motorIdx int) int {
	config := s.motorConfigs[motorIdx]
	return int(math.Round((lengthMM * float64(config.StepsPerRev)) / (config.PulleyMM * math.Pi)))
}

func (s *KinematicsService) calculateIK(pos domain.Point) [4]float64 {
	var lengths [4]float64
	x := pos.X
	y := pos.Y

	lengths[0] = math.Sqrt(x*x + y*y)
	lengths[1] = math.Sqrt(math.Pow(s.Width-x, 2) + y*y)
	lengths[2] = math.Sqrt(math.Pow(s.Width-x, 2) + math.Pow(s.Height-y, 2))
	lengths[3] = math.Sqrt(x*x + math.Pow(s.Height-y, 2))
	return lengths
}

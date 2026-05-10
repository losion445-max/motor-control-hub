package usecase

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/losion445-max/motor-control-hub/internal/config"
	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type KinematicsService struct {
	cfg    *config.GlobalConfig
	Motors []domain.IMotor

	stepsPerMM float64

	mu                   sync.RWMutex
	currentPosition      domain.Point
	currentAbsoluteSteps [MotorCount]int
}

func NewKinematicsService(cfg *config.GlobalConfig, motorInstances []domain.IMotor) (*KinematicsService, error) {
	if len(motorInstances) != MotorCount {
		return nil, fmt.Errorf("kinematics service requires exactly %d motors, got %d", MotorCount, len(motorInstances))
	}

	s := &KinematicsService{
		cfg:    cfg,
		Motors: motorInstances,
	}

	s.SyncConfig()

	return s, nil
}

func (s *KinematicsService) SyncConfig() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stepsPerMM = float64(s.cfg.Kinematics.StepsPerRev) / (s.cfg.Kinematics.Diameter * math.Pi)
}

func (s *KinematicsService) MoveTo(ctx context.Context, targetPos domain.Point, baseSpeed float64) error {
	s.mu.RLock()
	target := s.clampPoint(targetPos)
	targetLengths := s.calculateIK(target)

	deltas := make([]int, MotorCount)
	targetSteps := make([]int, MotorCount)
	var maxDelta float64

	for i := 0; i < MotorCount; i++ {
		steps := int(math.Round(targetLengths[i] * s.stepsPerMM))
		targetSteps[i] = steps
		deltas[i] = steps - s.currentAbsoluteSteps[i]

		absDelta := math.Abs(float64(deltas[i]))
		if absDelta > maxDelta {
			maxDelta = absDelta
		}
	}
	s.mu.RUnlock()

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

	s.mu.Lock()
	s.currentAbsoluteSteps = [MotorCount]int(targetSteps)
	s.currentPosition = target
	s.mu.Unlock()

	return nil
}

func (s *KinematicsService) Calibrate(ctx context.Context, speed float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentAbsoluteSteps = [MotorCount]int{}
	s.currentPosition = domain.Point{X: 0, Y: 0}
	return nil
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

	if len(errChan) > 0 {
		return <-errChan
	}
	return nil
}

func (s *KinematicsService) clampPoint(p domain.Point) domain.Point {
	return domain.Point{
		X: math.Max(0, math.Min(p.X, s.cfg.Kinematics.Width)),
		Y: math.Max(0, math.Min(p.Y, s.cfg.Kinematics.Height)),
	}
}

func (s *KinematicsService) calculateIK(pos domain.Point) [MotorCount]float64 {
	x, y := pos.X, pos.Y
	w, h := s.cfg.Kinematics.Width, s.cfg.Kinematics.Height

	distTL := math.Sqrt(x*x + y*y)
	distTR := math.Sqrt(math.Pow(w-x, 2) + y*y)
	distBR := math.Sqrt(math.Pow(w-x, 2) + math.Pow(h-y, 2))
	distBL := math.Sqrt(x*x + math.Pow(h-y, 2))

	var lengths [MotorCount]float64

	lengths[s.cfg.MotorMapping[0]-1] = distTL
	lengths[s.cfg.MotorMapping[1]-1] = distTR
	lengths[s.cfg.MotorMapping[2]-1] = distBR
	lengths[s.cfg.MotorMapping[3]-1] = distBL

	return lengths
}

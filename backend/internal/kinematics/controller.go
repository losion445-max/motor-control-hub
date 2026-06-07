package kinematics

import (
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

const (
	MotorTL    = 0
	MotorTR    = 1
	MotorBR    = 2
	MotorBL    = 3
	MaxMotorHz = 50000.0
)

type KinematicsController struct {
	mu         sync.RWMutex
	zone       domain.WorkZone
	stepsPerMM [4]float64
	currentPos domain.Point2D
	currentMM  [4]float64
	calibrated bool
}

func NewKinematicsController(zone domain.WorkZone, configs [4]domain.MotorConfig) (*KinematicsController, error) {
	var stepsPerMM [4]float64
	for i, cfg := range configs {
		if cfg.PulleyMM <= 0 {
			return nil, fmt.Errorf("motor %d: pulley diameter must be positive", i)
		}
		stepsPerMM[i] = float64(cfg.StepsPerRev) / (cfg.PulleyMM * math.Pi)
		log.Printf("[Kinematics] Motor %d: StepsPerRev=%d, Pulley=%.2fmm, Ratio=%.2f step/mm",
			i+1, cfg.StepsPerRev, cfg.PulleyMM, stepsPerMM[i])
	}
	return &KinematicsController{
		zone:       zone,
		stepsPerMM: stepsPerMM,
	}, nil
}

func (c *KinematicsController) SetHome(pos domain.Point2D) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentPos = pos
	c.currentMM = c.ik(pos)
	c.calibrated = true
	log.Printf("[Kinematics] Home set to (%.2f, %.2f). Cable lengths: TL=%.2f TR=%.2f BR=%.2f BL=%.2f",
		pos.X, pos.Y, c.currentMM[0], c.currentMM[1], c.currentMM[2], c.currentMM[3])
}

func (c *KinematicsController) CurrentPosition() domain.Point2D {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentPos
}

func (c *KinematicsController) IsCalibrated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.calibrated
}

func (c *KinematicsController) Solve(next domain.Point2D) (domain.Tick, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	targetMM := c.ik(next)
	var tick domain.Tick

	for i := range 4 {
		deltaMM := targetMM[i] - c.currentMM[i]
		steps := int(math.Round(deltaMM * c.stepsPerMM[i]))
		hz := math.Abs(float64(steps)) * DefaultHz
		if hz > MaxMotorHz {
			hz = MaxMotorHz
		}
		tick[i] = domain.MotorCommand{
			MotorID: i + 1,
			Steps:   steps,
			Hz:      hz,
		}
		log.Printf("[Kinematics] Motor %d: delta=%.2fmm steps=%d hz=%.2f",
			i+1, deltaMM, steps, hz)
	}

	c.currentPos = next
	c.currentMM = targetMM

	return tick, nil
}

func (c *KinematicsController) ik(p domain.Point2D) [4]float64 {
	x, y := p.X, p.Y
	w, h := c.zone.Width, c.zone.Height
	return [4]float64{
		math.Sqrt(x*x + y*y),
		math.Sqrt((w-x)*(w-x) + y*y),
		math.Sqrt((w-x)*(w-x) + (h-y)*(h-y)),
		math.Sqrt(x*x + (h-y)*(h-y)),
	}
}

func (c *KinematicsController) SetWorkZone(zone domain.WorkZone) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.zone = zone
	log.Printf("[Kinematics] WorkZone updated: %.0fx%.0f", zone.Width, zone.Height)
}

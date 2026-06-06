package kinematics

import (
	"fmt"
	"math"
	"sync"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

const (
	MotorTL = 0
	MotorTR = 1
	MotorBR = 2
	MotorBL = 3
)

type KinematicsController struct {
	mu         sync.RWMutex
	zone       domain.WorkZone
	stepsPerMM [4]float64
	currentPos domain.Point2D
	currentMM  [4]float64
}

func NewKinematicsController(zone domain.WorkZone, configs [4]domain.MotorConfig) (*KinematicsController, error) {
	var stepsPerMM [4]float64
	for i, cfg := range configs {
		if cfg.PulleyMM <= 0 {
			return nil, fmt.Errorf("motor %d: pulley diameter must be positive", i)
		}
		stepsPerMM[i] = float64(cfg.StepsPerRev) / (cfg.PulleyMM * math.Pi)
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
}

func (c *KinematicsController) CurrentPosition() domain.Point2D {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentPos
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

		tick[i] = domain.MotorCommand{
			MotorID: i + 1,
			Steps:   steps,
			Hz:      hz,
		}
	}

	c.currentPos = next
	c.currentMM = targetMM

	return tick, nil
}

func (c *KinematicsController) ik(p domain.Point2D) [4]float64 {
	x, y := p.X, p.Y
	w, h := c.zone.Width, c.zone.Height
	return [4]float64{
		math.Sqrt(x*x + y*y),                 // TL
		math.Sqrt((w-x)*(w-x) + y*y),         // TR
		math.Sqrt((w-x)*(w-x) + (h-y)*(h-y)), // BR
		math.Sqrt(x*x + (h-y)*(h-y)),         // BL
	}
}

func (c *KinematicsController) SetWorkZone(zone domain.WorkZone) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.zone = zone
}

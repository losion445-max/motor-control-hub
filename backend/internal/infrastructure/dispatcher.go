package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/losion445-max/motor-control-hub/internal/domain"
	"github.com/losion445-max/motor-control-hub/internal/kinematics"
)

type MotorDispatcher struct {
	motors [4]domain.IMotor
}

func NewMotorDispatcher(motors [4]domain.IMotor) *MotorDispatcher {
	return &MotorDispatcher{motors: motors}
}

func (d *MotorDispatcher) Dispatch(ctx context.Context, ticks []domain.Tick) error {
	ticker := time.NewTicker(time.Duration(float64(time.Second) / kinematics.DefaultHz))
	defer ticker.Stop()

	for _, tick := range ticks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := d.sendTick(ctx, tick); err != nil {

				_ = d.StopAll(context.Background())
				return err
			}
		}
	}
	return nil
}

func (d *MotorDispatcher) sendTick(ctx context.Context, tick domain.Tick) error {
	ready := make(chan struct{})

	var wg sync.WaitGroup
	errs := make(chan error, 4)

	for i := range 4 {
		wg.Add(1)
		go func(idx int, cmd domain.MotorCommand) {
			defer wg.Done()
			<-ready
			if err := d.motors[idx].Move(ctx, cmd.Steps, cmd.Hz); err != nil {
				errs <- fmt.Errorf("motor %d: %w", cmd.MotorID, err)
			}
		}(i, tick[i])
	}

	close(ready)
	wg.Wait()
	close(errs)

	var first error
	for err := range errs {
		if first == nil {
			first = err
		}
	}
	return first
}

func (d *MotorDispatcher) StopAll(ctx context.Context) error {
	var wg sync.WaitGroup
	errs := make(chan error, 4)

	for i := range 4 {
		wg.Add(1)
		go func(m domain.IMotor) {
			defer wg.Done()
			if err := m.Stop(ctx); err != nil {
				errs <- err
			}
		}(d.motors[i])
	}

	wg.Wait()
	close(errs)

	var first error
	for err := range errs {
		if first == nil {
			first = err
		}
	}
	return first
}

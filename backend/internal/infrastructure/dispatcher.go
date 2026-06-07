package infrastructure

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/losion445-max/motor-control-hub/internal/domain"
	"github.com/losion445-max/motor-control-hub/internal/infrastructure/network"
	"github.com/losion445-max/motor-control-hub/internal/kinematics"
)

type MotorDispatcher struct {
	registry *network.MotorRegistry
}

func NewMotorDispatcher(registry *network.MotorRegistry) *MotorDispatcher {
	return &MotorDispatcher{registry: registry}
}

func (d *MotorDispatcher) Dispatch(ctx context.Context, ticks []domain.Tick) error {
	log.Printf("[Dispatcher] Starting movement sequence with %d ticks", len(ticks))

	ticker := time.NewTicker(time.Duration(float64(time.Second) / kinematics.DefaultHz))
	defer ticker.Stop()

	for i, tick := range ticks {
		select {
		case <-ctx.Done():
			log.Printf("[Dispatcher] Context cancelled at tick %d", i)
			return ctx.Err()
		case <-ticker.C:
			// Логируем содержимое тика перед отправкой
			log.Printf("[Dispatcher] Tick %d: M1(s=%d, h=%.2f) M2(s=%d, h=%.2f) M3(s=%d, h=%.2f) M4(s=%d, h=%.2f)",
				i, tick[0].Steps, tick[0].Hz, tick[1].Steps, tick[1].Hz, tick[2].Steps, tick[2].Hz, tick[3].Steps, tick[3].Hz)

			if err := d.sendTick(ctx, tick); err != nil {
				log.Printf("[Dispatcher] CRITICAL: Error at tick %d: %v", i, err)
				_ = d.StopAll(context.Background())
				return err
			}
		}
	}
	log.Println("[Dispatcher] Movement sequence finished successfully")
	return nil
}

func (d *MotorDispatcher) sendTick(ctx context.Context, tick domain.Tick) error {
	motors := d.registry.Motors()
	ready := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, 4)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(idx int, cmd domain.MotorCommand) {
			defer wg.Done()
			<-ready

			if motors[idx] == nil {
				errs <- fmt.Errorf("motor %d is not connected", idx+1)
				return
			}

			// Логируем запрос к мотору
			if cmd.Steps != 0 {
				log.Printf("[Motor %d] Moving: steps=%d, hz=%.2f", cmd.MotorID, cmd.Steps, cmd.Hz)
			}

			if err := motors[idx].Move(ctx, cmd.Steps, cmd.Hz); err != nil {
				errs <- fmt.Errorf("motor %d execution error: %w", cmd.MotorID, err)
			}
		}(i, tick[i])
	}

	close(ready)
	wg.Wait()
	close(errs)

	var firstErr error
	for err := range errs {
		if firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (d *MotorDispatcher) StopAll(ctx context.Context) error {
	log.Println("[Dispatcher] EMERGENCY STOP initiated")
	motors := d.registry.Motors()
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if motors[idx] != nil {
				if err := motors[idx].Stop(ctx); err != nil {
					log.Printf("[Dispatcher] Failed to stop motor %d: %v", idx+1, err)
				}
			}
		}(i)
	}

	wg.Wait()
	return nil
}

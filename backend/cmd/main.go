package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/losion445-max/motor-control-hub/internal/config"
	"github.com/losion445-max/motor-control-hub/internal/domain"
	"github.com/losion445-max/motor-control-hub/internal/infrastructure"
	"github.com/losion445-max/motor-control-hub/internal/infrastructure/esp32"
	"github.com/losion445-max/motor-control-hub/internal/infrastructure/network"
	"github.com/losion445-max/motor-control-hub/internal/kinematics"
	"github.com/losion445-max/motor-control-hub/internal/usecase"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("[MAIN] Starting Motor Control Hub...")

	cfg := &config.GlobalConfig{}
	cfg.Kinematics.Width = 1000.0
	cfg.Kinematics.Height = 1000.0
	cfg.Kinematics.Diameter = 90.0
	cfg.Kinematics.StepsPerRev = 10000

	scanner := network.NewARPScanner("wlan1", 500*time.Millisecond)
	motorSlice, err := usecase.BootstrapMotors(ctx, scanner, esp32.NewMotorClientFactory())
	if err != nil {
		log.Fatalf("[MAIN] Failed to bootstrap motors: %v", err)
	}

	var motors [4]domain.IMotor
	copy(motors[:], motorSlice)

	zone := domain.WorkZone{
		Width:  cfg.Kinematics.Width,
		Height: cfg.Kinematics.Height,
	}

	var motorConfigs [4]domain.MotorConfig
	for i := range motors {
		mc, err := motors[i].GetConfig(ctx)
		if err != nil {
			log.Fatalf("[MAIN] Failed to get config from motor %d: %v", i+1, err)
		}
		motorConfigs[i] = *mc
	}

	controller, err := kinematics.NewKinematicsController(zone, motorConfigs)
	if err != nil {
		log.Fatalf("[MAIN] Failed to init kinematics controller: %v", err)
	}

	planner := kinematics.NewTrajectoryPlanner(kinematics.DefaultHz, kinematics.DefaultAccelMmS2)
	dispatcher := infrastructure.NewMotorDispatcher(motors)

	// 4. юзкейсы
	moveTo := usecase.NewMoveTo(planner, controller, dispatcher)
	setHome := usecase.NewSetHome(controller)
	stopAll := usecase.NewStopAll(dispatcher)
	getStatus := usecase.NewGetStatus(motors, controller)
	setEnabled := usecase.NewSetEnabled(motors)
	moveSingleMotor := usecase.NewMoveSingleMotor(motors)
	getConfig := usecase.NewGetConfig(cfg, motors)
	updateConfig := usecase.NewUpdateConfig(cfg, controller)

	// 5. HTTP
	mux := http.NewServeMux()
	handler := network.NewMotorHandler(
		moveTo,
		setHome,
		stopAll,
		getStatus,
		setEnabled,
		moveSingleMotor,
		getConfig,
		updateConfig,
	)
	handler.MapRoutes(mux)

	log.Println("[MAIN] Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

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
	log.Println("[MAIN] Starting Motor Control Hub...")

	// 1. конфиг
	cfg := &config.GlobalConfig{}
	cfg.Kinematics.Width = 1000.0
	cfg.Kinematics.Height = 1000.0
	cfg.Kinematics.Diameter = 90.0
	cfg.Kinematics.StepsPerRev = 10000

	// 2. registry — стартует сразу, моторы подключаются в фоне
	scanner := network.NewARPScanner("wlan1", 500*time.Millisecond)
	registry := network.NewMotorRegistry(scanner, func(cfg domain.MotorConfig) domain.IMotor {
		return esp32.NewMotorClient(&cfg)
	})
	go registry.Run(context.Background())

	zone := domain.WorkZone{
		Width:  cfg.Kinematics.Width,
		Height: cfg.Kinematics.Height,
	}

	defaultMotorConfig := domain.MotorConfig{
		StepsPerRev: cfg.Kinematics.StepsPerRev,
		PulleyMM:    cfg.Kinematics.Diameter,
	}
	var motorConfigs [4]domain.MotorConfig
	for i := range motorConfigs {
		motorConfigs[i] = defaultMotorConfig
	}

	controller, err := kinematics.NewKinematicsController(zone, motorConfigs)
	if err != nil {
		log.Fatalf("[MAIN] Failed to init kinematics controller: %v", err)
	}

	planner := kinematics.NewTrajectoryPlanner(kinematics.DefaultHz, kinematics.DefaultAccelMmS2)
	dispatcher := infrastructure.NewMotorDispatcher(registry)

	// 4. юзкейсы
	moveTo := usecase.NewMoveTo(planner, controller, dispatcher, registry)
	setHome := usecase.NewSetHome(controller)
	stopAll := usecase.NewStopAll(dispatcher)
	getStatus := usecase.NewGetStatus(registry, controller)
	setEnabled := usecase.NewSetEnabled(registry)
	moveSingleMotor := usecase.NewMoveSingleMotor(registry)
	getConfig := usecase.NewGetConfig(cfg, registry)
	updateConfig := usecase.NewUpdateConfig(cfg, controller)

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

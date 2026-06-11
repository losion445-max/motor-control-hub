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

	cfg := &config.GlobalConfig{}
	cfg.Kinematics.Width = 1200.0 
	cfg.Kinematics.Height = 2400
	cfg.Kinematics.Diameter = 68.7
	cfg.Kinematics.StepsPerRev = 10000

	scanner := network.NewARPScanner("wlan1", 500*time.Millisecond)
	registry := network.NewMotorRegistry(scanner, func(cfg domain.MotorConfig) domain.IMotor {
		return esp32.NewMotorClient(&cfg)
	})

	go registry.Run(context.Background())

	tcpServer := network.NewTCPServer(":9000", registry)
	go func() {
		if err := tcpServer.Run(context.Background()); err != nil {
			log.Printf("[TCP] Server error: %v", err)
		}
	}()

	zone := domain.WorkZone{
		Width:   cfg.Kinematics.Width,
		Height:  cfg.Kinematics.Height,
		Margin:  150,
		ZOffset: 200,
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
	controller.SetHome(domain.Point2D{X: 0.0, Y: 0.0})
	if err != nil {
		log.Fatalf("[MAIN] Failed to init kinematics controller: %v", err)
	}

	planner := kinematics.NewTrajectoryPlanner(kinematics.DefaultHz, kinematics.DefaultAccelMmS2)
	dispatcher := infrastructure.NewMotorDispatcher(registry)

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

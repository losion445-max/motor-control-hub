package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/losion445-max/motor-control-hub/internal/config"
	"github.com/losion445-max/motor-control-hub/internal/infrastructure/esp32"
	"github.com/losion445-max/motor-control-hub/internal/infrastructure/network"
	"github.com/losion445-max/motor-control-hub/internal/usecase"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("[MAIN] Starting Motor Control Hub...")

	appConfig := &config.GlobalConfig{}
	appConfig.Kinematics.Width = 1000.0
	appConfig.Kinematics.Height = 1000.0
	appConfig.Kinematics.Diameter = 90.0
	appConfig.Kinematics.StepsPerRev = 10000
	appConfig.MotorMapping = [4]int{2, 3, 4, 1}

	scanner := network.NewARPScanner("wlan1", 500*time.Millisecond)
	motors, err := usecase.BootstrapMotors(ctx, scanner, esp32.NewMotorClientFactory())
	if err != nil {
		log.Fatalf("[MAIN] Failed to initialize system: %v", err)
	}
	log.Printf("[MAIN] System initialized with %d active motors", len(motors))

	kinematics, err := usecase.NewKinematicsService(appConfig, motors)
	if err != nil {
		log.Fatalf("[MAIN] Failed to initialize Kinematics service: %v", err)
	}

	motorOrchestrator := usecase.NewMotorOrchestrator(motors, kinematics)

	mux := http.NewServeMux()
	handler := network.NewMotorHandler(motorOrchestrator)
	handler.MapRoutes(mux)

	log.Println("[MAIN] Control Hub listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

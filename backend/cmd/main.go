package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/losion445-max/motor-control-hub/internal/infrastructure/network"
	"github.com/losion445-max/motor-control-hub/internal/usecase"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("[MAIN] Starting Motor Control Hub...")

	scanner := network.NewARPScanner("wlan1")

	motors, err := usecase.BootstrapMotors(ctx, scanner)
	if err != nil {
		log.Fatalf("[MAIN] Failed to initialize system: %v", err)
	}

	log.Printf("[MAIN] System initialized with %d active motors", len(motors))

	for _, m := range motors {
		status, _ := m.GetStatus(ctx)
		log.Printf("[MAIN] Motor %d is at position %d", status.MotorID, status.CurrentSteps)
	}

	kinematics, err := usecase.NewKinematicsService(10, 10, motors)
	if err != nil {
		log.Fatalf("[MAIN] Failed to initialize Kinematics service: %v", err)
	}

	motorOrchestrator := usecase.NewMotorOrchestrator(motors, kinematics)

	mux := http.NewServeMux()
	handler := network.NewMotorHandler(motorOrchestrator)
	handler.MapRoutes(mux)

	log.Println("Control Hub listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))

}

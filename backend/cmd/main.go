package main

import (
	"context"
	"fmt"
	"log"
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

	c, err := motorOrchestrator.GetAllAggregatedConfig(ctx)
	if err != nil {
		log.Fatalf("[MAIN] Failed to get aggregated config %v", err)
	}
	s, err := motorOrchestrator.GetAllAggregatedStatus(ctx)
	if err != nil {
		log.Fatalf("[MAIN] Failed to get aggregated status: %v", err)
	}

	fmt.Println(c)
	fmt.Println(s)

}

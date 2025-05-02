package main

import (
	"fmt"
	"os"
	"strconv"

	"calculator/internal/agent"
	calculator "calculator/internal/grpc/calculator"

	"google.golang.org/grpc"
)

func main() {
	computingPower := 1
	if v := os.Getenv("COMPUTING_POWER"); v != "" {
		if cp, err := strconv.Atoi(v); err == nil {
			computingPower = cp
		}
	}

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := calculator.NewCalculatorAgentServiceClient(conn)
	for i := 0; i < computingPower; i++ {
		go agent.Worker(i, client)
	}

	fmt.Printf("Agent started with %d workers\n", computingPower)
	select {}
}
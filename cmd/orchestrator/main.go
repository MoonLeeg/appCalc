package main

import (
	"calculator/internal/database"
	pb "calculator/internal/grpc/calculator"
	"calculator/internal/orchestrator"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"
)

const (
	httpPort     = ":8080"
	grpcPort     = ":50051"
	dbPath       = "calculator.db"
	jwtSecretEnv = "JWT_SECRET"
)

func main() {
	fmt.Println("Запуск Оркестратора...")

	dbStore, err := database.NewStore(dbPath)
	if err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer dbStore.Close()

	if err := dbStore.InitDB(); err != nil {
		log.Fatalf("Ошибка миграции БД: %v", err)
	}
	fmt.Println("База данных инициализирована.")

	jwtSecret := os.Getenv(jwtSecretEnv)
	if jwtSecret == "" {
		log.Fatalf("Переменная окружения %s не установлена!", jwtSecretEnv)
	}
	authService := orchestrator.NewAuthService(dbStore, jwtSecret)
	schedulerService := orchestrator.NewScheduler(dbStore)
	grpcServerInstance := orchestrator.NewCalculatorGRPCServer(dbStore, schedulerService.GetOperationTimes(), schedulerService)

	httpHandlers := orchestrator.NewHTTPHandlers(authService, dbStore, schedulerService)

	go func() {
		lis, err := net.Listen("tcp", grpcPort)
		if err != nil {
			log.Fatalf("Ошибка прослушивания gRPC порта %s: %v", grpcPort, err)
		}
		s := grpc.NewServer()
		pb.RegisterCalculatorAgentServiceServer(s, grpcServerInstance)

		fmt.Printf("gRPC сервер слушает на %s\n", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Ошибка gRPC сервера: %v", err)
		}
	}()

	router := http.NewServeMux()

	router.HandleFunc("/api/v1/register", httpHandlers.RegisterHandler)
	router.HandleFunc("/api/v1/login", httpHandlers.LoginHandler)

	router.Handle("/api/v1/calculate", authService.JWTMiddleware(http.HandlerFunc(httpHandlers.CalculateHandler)))
	router.Handle("/api/v1/expressions", authService.JWTMiddleware(http.HandlerFunc(httpHandlers.ExpressionsHandler)))
	router.Handle("/api/v1/expressions/", authService.JWTMiddleware(http.HandlerFunc(httpHandlers.ExpressionsHandler))) // Для путей с ID

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./web/static/index.html")
		} else {
			http.NotFound(w, r)
		}
	})

	fmt.Printf("HTTP сервер слушает на %s\n", httpPort)
	corsRouter := orchestrator.EnableCORS(router)
	if err := http.ListenAndServe(httpPort, corsRouter); err != nil {
		log.Fatalf("Ошибка старта HTTP сервера: %v", err)
	}

	fmt.Println("Оркестратор остановлен.")
}
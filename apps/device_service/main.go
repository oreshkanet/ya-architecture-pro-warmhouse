package main

import (
	"context"
	"device-service/internal/client"
	"device-service/internal/delivery/http"
	"device-service/internal/repository"
	"device-service/internal/service"
	"errors"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbURL := os.Getenv("DATABASE_URL")
	dbName := os.Getenv("DATABASE_NAME")
	warmURL := os.Getenv("WARM_URL")

	debugStr := os.Getenv("DEBUG")
	isDebug, _ := strconv.ParseBool(debugStr)

	// Инициализация сервисов
	pgxPool, err := pgxpool.New(ctx, dbURL+"/"+dbName+"?sslmode=disable")
	if err != nil {
		log.Fatal("pgx connect:", err)
	}
	defer pgxPool.Close()

	warmClient := client.NewWarmServiceClient(warmURL)
	locationRepo := repository.NewLocationRepository(pgxPool)
	deviceRepo := repository.NewDeviceRepository(pgxPool)
	locationService := service.NewLocationService(locationRepo, deviceRepo)
	deviceService := service.NewDeviceService(locationRepo, deviceRepo, warmClient)

	// Создаём HTTP-сервер
	httpServer := http.NewHttpService("8082", isDebug)

	// Регистрируем маршруты
	handler := http.NewHandler(deviceService, locationService)
	handler.InitRoutes(httpServer.Engine())

	// Канал для сигналов ОС
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Println("Запуск HTTP-сервера на :8082")
		if err := httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка запуска HTTP-сервера: %v", err)
		}
	}()

	// Ждём сигнала завершения
	<-sigCh
	log.Println("Получен сигнал завершения, остановка сервера...")

	if err := httpServer.Stop(ctx); err != nil {
		log.Fatalf("Ошибка остановки HTTP-сервера: %v", err)
	}

	log.Println("Сервер успешно остановлен")
}

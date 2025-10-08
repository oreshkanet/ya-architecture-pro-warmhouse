package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"telemetry-service/internal/delivery/consumer"
	"telemetry-service/internal/delivery/http"
	"telemetry-service/internal/repository"
	"telemetry-service/internal/service/aggregator"
	"telemetry-service/internal/service/telemetry"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	debugStr := os.Getenv("DEBUG")
	isDebug, _ := strconv.ParseBool(debugStr)

	rmqURL := os.Getenv("RABBITMQ_URL")
	rmqQueue := os.Getenv("RABBITMQ_QUEUE")
	clhURL := os.Getenv("CLICKHOUSE_URL")
	clhDb := os.Getenv("CLICKHOUSE_DATABASE")

	// ClickHouse
	repo, err := repository.NewClickHouseRepo(clhURL, clhDb)
	if err != nil {
		log.Fatal("ClickHouse init failed:", err)
	}

	// Aggregation Engine
	engine := aggregator.NewAggregationEngine(repo)

	// RabbitMQ Consumer
	rmq, err := consumer.NewRabbitMQConsumer(rmqURL, rmqQueue, engine)
	if err != nil {
		log.Fatal("RabbitMQ init failed:", err)
	}
	go rmq.Start()

	// Создаём HTTP-сервер
	httpServer := http.NewHttpService("8089", isDebug)

	telemetryService := telemetry.NewTelemetryService(repo)

	// Регистрируем маршруты
	handler := http.NewHandler(telemetryService)
	handler.InitRoutes(httpServer.Engine())

	// Канал для сигналов ОС
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Println("Запуск HTTP-сервера на :8089")
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

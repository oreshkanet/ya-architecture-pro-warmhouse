package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"warm-service/internal/client/telemetry"
	"warm-service/internal/delivery/http"
	"warm-service/internal/repository"
	"warm-service/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/streadway/amqp"
)

const migrations = "file://migrations"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbURL := os.Getenv("DATABASE_URL")
	dbName := os.Getenv("DATABASE_NAME")
	rmqURL := os.Getenv("RABBITMQ_URL")
	monolithURL := os.Getenv("MONOLITH_URL")

	debugStr := os.Getenv("DEBUG")
	isDebug, _ := strconv.ParseBool(debugStr)

	// Подключение к RabbitMQ
	rabbitConn, err := amqp.Dial(rmqURL)
	if err != nil {
		log.Fatal("RabbitMQ connect:", err)
	}
	defer rabbitConn.Close()

	publisher, err := telemetry.NewPublisher(rabbitConn)
	if err != nil {
		log.Fatal("RabbitMQ channel:", err)
	}

	// Инициализация сервисов
	pgxPool, err := pgxpool.New(ctx, dbURL+"/"+dbName)
	if err != nil {
		log.Fatal("pgx connect:", err)
	}
	defer pgxPool.Close()

	repo := repository.NewWarmRepo(pgxPool)
	warmService := service.NewWarmService(repo, publisher, monolithURL)

	// Создаём HTTP-сервер
	httpServer := http.NewHttpService("8084", isDebug)

	// Регистрируем маршруты
	handler := http.NewHandler(warmService)
	handler.InitRoutes(httpServer.Engine())

	// Канал для сигналов ОС
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	warmService.StartTelemetryCollection(ctx)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Println("Запуск HTTP-сервера на :8084")
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

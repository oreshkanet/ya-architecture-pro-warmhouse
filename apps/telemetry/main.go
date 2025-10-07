package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"warm/internal/delivery/http"
)

func main() {
	debugStr := os.Getenv("DEBUG")
	isDebug, _ := strconv.ParseBool(debugStr)

	// Создаём HTTP-сервер
	httpServer := http.NewHttpService("8081", isDebug) // указываем порт с двоеточием

	// Регистрируем маршруты
	handler := http.NewHandler()
	handler.InitRoutes(httpServer.Engine()) // ← нужно передать *gin.Engine

	// Канал для сигналов ОС
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Println("Запуск HTTP-сервера на :8081")
		if err := httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка запуска HTTP-сервера: %v", err)
		}
	}()

	// Ждём сигнала завершения
	<-sigCh
	log.Println("Получен сигнал завершения, остановка сервера...")

	// Плавная остановка
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Stop(ctx); err != nil {
		log.Fatalf("Ошибка остановки HTTP-сервера: %v", err)
	}

	log.Println("Сервер успешно остановлен")
}

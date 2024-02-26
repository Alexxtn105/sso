package main

import (
	"fmt"
	"grpc-service-ref/internal/config"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// TODO инициализировать объект конфига
	cfg := config.MustLoad()
	fmt.Println("Конфигурация загружена:\n", cfg)

	// TODO инициализировать логгер
	log := setupLogger(cfg.Env)
	fmt.Println("Логгер загружен:\n", log)

	// TODO инициализировать приложение (app)

	// TODO запустить gRPC-сервер приложения
}

// настройка логгера
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {

	//если локальный запуск
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)

	//запуск на удаленном dev-сервере
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)

	//запуск в продакшен
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)

	}

	return log
}

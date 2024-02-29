// `cmd/sso/main.go`
// точка входа в приложение
package main

import (
	"fmt"
	"log/slog"
	"os"

	"grpc-service-ref/internal/app"
	"grpc-service-ref/internal/config"
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
	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	// TODO запустить gRPC-сервер приложения
	application.GRPCServer.MustRun()

	// !!!
	// Вместо строчки application.GRPCServer.MustRun()
	// можете научить своё основное приложение автоматически запускать все внутренние,
	// а не дёргать запуск внутренних в main().
	// То есть, выглядеть это будет так:
	//
	// application.MustRun()

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

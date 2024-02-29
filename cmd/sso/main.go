// `cmd/sso/main.go`
// точка входа в приложение
package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	// TODO запустить gRPC-сервер приложения (вариант без GracefulStop)
	//application.GRPCServer.MustRun()

	// !!!
	// Вместо строчки application.GRPCServer.MustRun()
	// можете научить своё основное приложение автоматически запускать все внутренние,
	// а не дёргать запуск внутренних в main().
	// То есть, выглядеть это будет так:
	//
	// application.MustRun()

	//запускаем сервер как горутину для дальнейшего GracefulStop
	go func() {
		application.GRPCServer.MustRun()
	}()

	//Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	// Waiting for SIGINT (pkill -2) or SIGTERM
	<-stop
	// initiate graceful shutdown
	application.GRPCServer.Stop() // Assuming GRPCServer has Stop() method for graceful shutdown
	log.Info("Gracefully stopped")

	application.Storage.Close()
	log.Info("Storage closed")

	// TODO: Далее предлагаю вам самостоятельно написать
	// аналогичный метод Stop() для sqlite-реализации Storage.
	// Там это делается тоже одной строчкой:
	// s.db.Close()
	// При этом, конечно же, придется добавить Storage
	// в структуру App основного приложения (internal/app/app.go).
	// При желании, можете обернуть хранилище
	// в отдельное приложение StorageApp — это хороший подход.
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

// internal/app/app.go
package app

import (
	//"log/slog"
	//"time"

	grpcapp "grpc-service-ref/internal/app/grpc"
	"grpc-service-ref/internal/services/auth"
	"grpc-service-ref/internal/storage/sqlite"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {

	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	// А именно, трижды передаваемый storage.
	// Увы, таковы издержки минималистичных интерфейсов.
	// Но подумайте о том, что не во всех случаях реализациями этих интерфейсов
	// может быть storage, это даёт нам больше гибкости.
	// В любом случае, если эта концепция вам не по душе,
	// вы всегда вольны сделать по своему.
	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}

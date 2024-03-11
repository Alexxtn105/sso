// internal/grpc/auth/server.go
package auth

import (
	"context"
	"errors"
	"grpc-service-ref/internal/services/auth"
	"grpc-service-ref/internal/storage"

	// Подключаем сгенерированный код (имя ssov1 взято из контракта)
	ssov1 "github.com/Alexxtn105/protos/gen/go/sso"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// serverAPI Эта структура (ssov1.UnimplementedAuthServer) представляет собой некую пустую имплементацию всех методов gRPC сервиса.
// Использование этой структуры помогает обеспечить обратную совместимость при изменении .proto файла.
// Если мы добавим новый метод в наш .proto файл и заново сгенерируем код, но не реализуем этот метод в serverAPI,
// то благодаря встраиванию UnimplementedAuthServer наш код все равно будет компилироваться,
// а новый метод просто вернет ошибку "Not implemented".
type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

// Интерфейс, который мы передавали в grpcApp
type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)

	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)

	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

// регистрация serverAPI в gRPC-сервере
func Register(gRPCServer *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth})
}

// RPC-метод логина
// Обратите внимание, что возвращаемую ошибку
// мы создаем с помощью специальной функции status.Error
// из библиотеки grpc/status.
// Это нужно для того, чтобы формат ошибки был понятен любому grpc-клиенту.
// Кроме того, мы присваиваем этой ошибке код из пакета grpc/codes —
// это тоже необходимо для совместимости с клиентами.
// К примеру, если не подошел пароль или мы не нашли пользователя в БД, это — codes.InvalidArgument,
// если же БД вернула неожиданную ошибку, это уже codes.Internal.
func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {

	// Вынес в validateLogin
	/*
		if req.Email == "" {
			return nil, status.Error(codes.InvalidArgument, "email is required")
		}

		if req.Password == "" {
			return nil, status.Error(codes.InvalidArgument, "password is required")
		}

		if req.GetAppId() == 0 {
			return nil, status.Error(codes.InvalidArgument, "app_id is required")
		}
	*/
	err := validateLogin(req)
	if err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		// Ошибку auth.ErrInvalidCredentials мы создадим ниже
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

// RPC-метод регистрации
func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		// Ошибку storage.ErrUserExists мы создадим ниже
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &ssov1.RegisterResponse{UserId: uid}, nil
}

// RPC-метод получения статуса администратора по ИД пользователя
func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "failed to check admin status")
	}

	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func validateRegister(req *ssov1.RegisterRequest) error {

	return nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetAppId() == 0 {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}

	return nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {

	return nil
}

// TODO: сделать "ручку" для изменения статуса администратора

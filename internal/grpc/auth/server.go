// internal/grpc/auth/server.go
package auth

import (
	"context"
	"errors"
	"grpc-service-ref/internal/storage"

	// Подключаем сгенерированный код
	ssov1 "github.com/Alexxtn105/protos/gen/go/sso"

	//"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	//Эта структура (ssov1.UnimplementedAuthServer) представляет собой некую пустую имплементацию всех методов gRPC сервиса.
	//Использование этой структуры помогает обеспечить обратную совместимость при изменении .proto файла.
	//Если мы добавим новый метод в наш .proto файл и заново сгенерируем код, но не реализуем этот метод в serverAPI,
	//то благодаря встраиванию UnimplementedAuthServer наш код все равно будет компилироваться,
	//а новый метод просто вернет ошибку "Not implemented".
	ssov1.UnimplementedAuthServer
	auth Auth
}

// Тот самый интерфейс, котрый мы передавали в grpcApp
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
}

// регистрация serverAPI в gRPC-сервере
func Register(gRPCServer *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth})
}

// RPC-метод логина
func (s *serverAPI) Login(
	ctx context.Context,
	in *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {

	//Обратите внимание, что возвращаемую ошибку
	//мы создаем с помощью специальной функции status.Error
	//из библиотеки grpc/status.
	//Это нужно для того, чтобы формат ошибки был понятен любому grpc-клиенту.
	//Кроме того, мы присваиваем этой ошибке код из пакета grpc/codes —
	//это тоже необходимо для совместимости с клиентами.
	//К примеру, если не подошел пароль или мы не нашли пользователя в БД, это — codes.InvalidArgument,
	//если же БД вернула неожиданную ошибку, это уже codes.Internal.
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is requred")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is requred")
	}

	if in.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "app_id is requred")
	}

	token, err := s.auth.Login(ctx, in.GetEmail(), in.GetPassword(), int(in.GetAppId()))
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
	in *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.auth.RegisterNewUser(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		// Ошибку storage.ErrUserExists мы создадим ниже
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &ssov1.RegisterResponse{UserId: uid}, nil
}

// internal/app/grpc/app.go

// отдельное приложение (internal/app/grpc) для gRPC-сервера вместе со всеми зависимостями
package grpcapp

import (
	//"context"
	"log/slog"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
)

// Структура, которая будет представлять приложение gRPC-сервера
// и интерфейс для сервисного слоя — в нашем случае это только Auth,
// но потенциально сервисов может быть больше.
type App struct {
	log         *slog.Logger
	gRRPCServer *grpc.Server
	port        int //порт на котором будет работать gRPC-сервер
}

// New creates new gRPC server app.
// Для конструктора используем библиотеку grpc-ecosystem/go-grpc-middleware
// содержащую готовые реализации некоторых полезных интерсепторов
// Установка:
// go get github.com/grpc-ecosystem/go-grpc-middleware/v2@v2.0.0
// Один из параметров authgrpc.Auth - это интерфейс сервисного слоя, не путать gRPC-сервисом Auth. Его мы напишем чуть ниже.
func New(log *slog.Logger, authService authgrpc.Auth, port int) *App {
	// TODO: создать gRPCServer и подключить к нему интерсепторы
	// пример создания сервера:
	// gRPCServer := grpc.NewServer(opts)
	// На вход он принимает различные опции, и в нашем случае это будут только интерсепторы:
	// инфо по интецепторам: https://grpc.io/blog/grpc-web-interceptor/
	//Интерсептор gRPC это, в некотором смысле, аналог Middleware из мира HTTP / REST серверов.
	// То есть, это функция, которая вызывается перед и/или после обработки RPC-вызова
	// на стороне сервера или клиента. С помощью интерсепторов мы можем выполнять
	// различные полезные действия (например, логирование запросов, аутентификацию, авторизация и др.),
	// не изменяя основной логики обработки RPC.
	// Итак, создаём новый сервер с единственным интерсептором
	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(),
	))

	// У нас пока всего один интерсептор, и я его обернул grpc.ChainUnaryInterceptor — это некий враппер,
	// который принимает в качестве аргументов набор интерсепторов,
	// а когда приходит одиночный запрос (Unary), запускает все эти интерсепторы поочерёдно
	// (об этом говорит слово Chain в названии).

	// Интерсептор recovery.UnaryServerInterceptor восстановит и обработает панику,
	// если она случится внутри хэндлера.
	// Полезная штука, ведь мы не хотим, чтобы паника в одном запросе уронила нам весь сервис,
	// остановив обработку даже корректных запросов.
	// Вообще, восстановление паники, это порой дискутивная тема, и если вам такой подход не нравится,
	// можете просто не добавлять этот интерсептор.

	// Регистрируем наш gRPC-сервис Auth, об этом будет ниже
	authgrpc.Register(gRPCServer, authService)

	// Вернуть объект App со всеми необходимыми полями
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}
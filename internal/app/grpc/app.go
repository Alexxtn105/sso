// internal/app/grpc/app.go

// отдельное приложение (internal/app/grpc/app.go) для gRPC-сервера вместе со всеми зависимостями
// cmd/sso/main.go — это была лишь точка входа
// Такой подход делает код main намного проще,
// и, главное — даёт возможность создавать экземпляр приложения в других местах — например, в тестах, что облегчает тестирование.
// При этом, gRPC-сервер мы завернём в ещё одно отдельное приложение (internal/app/grpc) вместе со всеми зависимостями.
package grpcapp

import (
	//"context"
	"context"
	"fmt"
	"log/slog"
	"net"

	authgrpc "grpc-service-ref/internal/grpc/auth"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Структура, которая будет представлять приложение gRPC-сервера
// и интерфейс для сервисного слоя — в нашем случае это только Auth,
// но потенциально сервисов может быть больше.
type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int //порт на котором будет работать gRPC-сервер
}

// New creates new gRPC server app.
// Для конструктора используем библиотеку grpc-ecosystem/go-grpc-middleware
// содержащую готовые реализации некоторых полезных интерсепторов
// Установка:
// go get github.com/grpc-ecosystem/go-grpc-middleware/v2@v2.0.0
// Один из параметров authgrpc.Auth - это интерфейс сервисного слоя, НЕ ПУТАТЬ с gRPC-сервисом Auth!!! Его мы напишем чуть ниже.
func New(log *slog.Logger, authService authgrpc.Auth, port int) *App {
	// TODO: создать gRPCServer и подключить к нему интерсепторы
	// пример создания сервера:
	// gRPCServer := grpc.NewServer(opts)
	// На вход он принимает различные опции, и в нашем случае это будут только интерсепторы:
	// инфо по интецепторам: https://grpc.io/blog/grpc-web-interceptor/
	// Интерсептор gRPC это, в некотором смысле, аналог Middleware из мира HTTP / REST серверов.
	// То есть, это функция, которая вызывается перед и/или после обработки RPC-вызова
	// на стороне сервера или клиента. С помощью интерсепторов мы можем выполнять
	// различные полезные действия (например, логирование запросов, аутентификацию, авторизация и др.),
	// не изменяя основной логики обработки RPC.

	// создаем опции интерцептора - для восстановления после паники
	// Интерсептор recovery.UnaryServerInterceptor восстановит и обработает панику,
	// если она случится внутри хэндлера.
	// Полезная штука, ведь мы не хотим, чтобы паника в одном запросе уронила нам весь сервис,
	// остановив обработку даже корректных запросов.
	// Вообще, восстановление паники, это порой дискутивная тема, и если вам такой подход не нравится,
	// можете просто не добавлять этот интерсептор.
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			//логируем информацию о панике с уровнем Error
			log.Error("Recovered from panic", slog.Any("panic", p))

			//можно либо честно вернуть клиенту содержимое паники
			// либо ответить - "internal error", если не хотим делиться внутренностями
			return status.Errorf(codes.Internal, "internal error")
		}),

		// TODO - сюда можно еще САМОСТОЯТЕЛЬНО добавить опции (например метрики и алерты)
		// ...
	}

	// Создаем опции для еще одного важного интерсептора,
	// который будет логировать все входящие запросы и ответы.
	// Это бывает очень полезно для поиске хвостов в случае поломок и дебага.
	// Для этого мы также возьмём готовое решение из пакета:
	// logging.UnaryServerInterceptor(log, opts) (пример импорта см. ниже).
	//На вход он принимает логгер и опции.
	// К сожалению, мы не можем просто передать наш текущий логгер,
	// т.к. у его метода Log() немного отличается сигнатура,
	// от той которую требует хэндлер
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived,
			logging.PayloadSent,
		),
	}

	// создаём новый сервер с единственным интерсептором и опциями
	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		//интерцептор восстановления после паники
		recovery.UnaryServerInterceptor(recoveryOpts...),
		//интерцептор логирования
		logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),

		//Помимо этих двух, могут также понадобиться интерсепторы для следующих целей:
		//трейсинг, метрики, алерты, авторизация и др.
		//Но в текущем проекте нам этого достаточно.
	))

	// Регистрируем gRPC-сервис Auth
	// Эта функция регистрирует реализацию сервиса аутентификации (authService)
	// на нашем gRPC сервере (gRPCServer).
	// В контексте gRPC это обычно означает, что сервер будет знать,
	// как обрабатывать входящие RPC-запросы, связанные с этим сервисом аутентификации,
	// потому что реализация этого сервиса (методы, которые она предоставляет)
	// теперь связаны с сервером gRPC.
	authgrpc.Register(gRPCServer, authService)

	// Вернуть объект App со всеми необходимыми полями
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

// обертка для интерцептора логгера, поскольку текущий логгер (logging) отличается сигнатурой от используемомго (slog)
// имеем сигнатуру: Log(context.Context, slog.Level, string, ...any)
// необходима: 		Log(context.Context, logging.Level, string, ...any)
// Здесь мы просто конвертируем имеющуюся функцию Log() в аналогичную из пакета интерсептора.
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		// TODO: САМОСТОЯТЕЛЬНО Замаскировать пароли в логах!!!
		// Потому что это потенциальная джыра в безопасности!!!
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

// MustRun runs gRPC server and panics if any error occurs.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	// Создаем Listener, который будет слушать TCP-сообщения,
	// адресованные нашему gRPC-серверу

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	//запускаем обработчик grpc-сообщений
	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

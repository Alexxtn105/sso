// tests/suite/suite.go
package suite

import (
	"context"
	"grpc-service-ref/internal/config"
	"net"
	"os"
	"strconv"
	"testing"

	ssov1 "github.com/Alexxtn105/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// В структуре ниже:
//*testing.T — объект для управлением тестом, подробнее можно почитать тут
//Cfg — обычный объект конфига, тот же что используется при запуске приложения из cmd
//AuthClient — gRPC-клиент нашего Auth-сервера, основной компонент Suit'а, с его помощью будем отправлять запросы в тестируемое приложение

type Suite struct {
	*testing.T                  // Потребуется для вызова методов *testing.T
	Cfg        *config.Config   // Конфигурация приложения
	AuthClient ssov1.AuthClient // Клиент для взаимодействия с gRPC-сервером Auth
}

const (
	grpcHost = "localhost"
)

func configPath() string {
	const key = "CONFIG_PATH"

	if v := os.Getenv(key); v != "" {
		return v
	}

	return "../config/local_tests.yaml"
}

// New creates new test suite.
func New(t *testing.T) (context.Context, *Suite) {

	t.Helper() // Функция будет восприниматься как вспомогательная для тестов
	// t.Helper() — помечает функцию как вспомогательную, это нужно для формирования
	//правильного вывода тестов, особенно при их падении.
	//А именно, что-то сфэйлится внутри такого хелпера,
	//в выводе будет указана родительская функция
	//(но в трейсе текущая функция будет, конечно), что очень удобно при отладке.

	t.Parallel() // Разрешаем параллельный запуск тестов

	// Читаем конфиг из файла
	cfg := config.MustLoadPath(configPath())
	// Основной родительский контекст
	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	// Когда тесты пройдут, закрываем контекст
	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	// Адрес нашего gRPC-сервера
	grpcAddress := net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))

	// Создаем клиент
	cc, err := grpc.DialContext(context.Background(),
		grpcAddress,
		// Используем insecure-коннект для тестов
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	//И вот тут мы можем снова прочувствовать крутость кодогенерации из Protobuf:
	// всего одной строчкой мы создаём готовый клиент для нашего gRPC-сервера:
	authClient := ssov1.NewAuthClient(cc)
	//Методами этого клиенту будут методы сервиса, описанные в контракте, т.е.:
	// authClient.Login() и authClient.Register()

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: authClient,
	}
}

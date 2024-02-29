// tests/auth_register_login_test.go
package tests

// Для генерации случайных данных я использую библиотеку brianvoe/gofakeit
// — очень крутая штука, может сгенерировать огромное количество различных видов данных.
// Для ее установки:
// go get github.com/brianvoe/gofakeit/v6@v6.23.2

import (
	"grpc-service-ref/tests/suite"
	"testing"
	"time"

	ssov1 "github.com/Alexxtn105/protos/gen/go/sso"
	"github.com/golang-jwt/jwt/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brianvoe/gofakeit/v6"
)

const (
	appID     = 1             // ID приложения, которое мы создали миграцией
	appSecret = "test-secret" // Секретный ключ приложения

	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	// Создаём Suite
	ctx, st := suite.New(t)

	//генерим случайные имя и пароль (с помощью библиотеки github.com/brianvoe/gofakeit/v6)
	email := gofakeit.Email()
	pass := randomFakePassword()

	// Сначала зарегистрируем нового пользователя, которого будем логинить
	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	// Это вспомогательный запрос, поэтому делаем лишь минимальные проверки
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	// А это основная проверка
	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    appID,
	})
	require.NoError(t, err)

	// ---------------------------ПРОВЕРКА РЕЗУЛЬТАТОВ
	// Получаем токен из ответа
	token := respLogin.GetToken()
	require.NotEmpty(t, token) // Проверяем, что он не пустой

	// Отмечаем время, в которое бы выполнен логин.
	// Это понадобится для проверки TTL токена
	loginTime := time.Now()

	// Парсим и валидируем токен
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return []byte(appSecret), nil
	})
	// Если ключ окажется невалидным, мы получим соответствующую ошибку
	require.NoError(t, err)

	// Преобразуем к типу jwt.MapClaims, в котором мы сохраняли данные
	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Проверяем содержимое токена
	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 1

	// Проверяем, что TTL токена примерно соответствует нашим ожиданиям.
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
	// Время, которое мы отмечаем и сохраняем в переменную loginTime
	// может быть немного неточным, т.к. точное время определяем сервер Auth,
	// а для наших тестов он, как мы помним, черная коробка.
	// То есть, может быть лаг между моментом обработки запроса сервером
	// и получением нами ответа.
	// Для тестов нас это устроит, но нам придется использовать функцию
	// assert.InDelta для проверки значения exp (expiration time).
	// А именно, мы берём сохраненный выше loginTime, добавляем к нему st.Cfg.TokenTTL,
	// преобразуем это в UnixTimestamp (в котором мы изначально сохраняли),
	// и проверяем, что поле exp соответствует этому значению
	// с точностью до одной секунды.
}

// генерация случайного пароля (используется библиотека github.com/brianvoe/gofakeit/v6)
func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}

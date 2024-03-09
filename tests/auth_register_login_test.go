// tests/auth_register_login_test.go
package tests

// Для генерации случайных данных я использую библиотеку brianvoe/gofakeit
// — очень крутая штука, может сгенерировать огромное количество различных видов данных.
// Для ее установки:
// go get github.com/brianvoe/gofakeit/v6@v6.23.2

// ЛИКБЕЗ по пакету "github.com/stretchr/testify" (взято с https://habr.com/ru/companies/joom/articles/666440/)
// В testify есть два основных пакета с проверками — assert и require.
// Набор проверок в них идентичен, но фейл require-проверки означает прерывание выполнения теста, а assert-проверки — нет.
// Когда мы пишем тест, мы хотим, чтобы неудачный запуск выдал нам как можно больше информации о текущем (неправильном) поведении программы.
// Но если у нас есть череда проверок с require, неудачный запуск сообщит нам только о первом несоответствии.
// Поэтому имеет смысл пользоваться require-проверками только если дальнейшее выполнение теста в случае невыполнения условия лишено смысла.
// Например, когда мы проверяем отсутствие ошибки, или валидируем длину списка, в который полезем дальше по коду теста.
// Используйте подходящие проверки:
// ❌	require.Nil(t, err)
// ✅	require.NoError(t, err)

// ❌	assert.Equal(t, 300.0, float64(price.Amount))
// ✅	assert.EqualValues(t, 300.0, price.Amount)

// ❌	assert.Equal(t, 0, len(result.Errors))
// ✅	assert.Empty(t, result.Errors)

// ❌	require.Equal(t, len(expected), len(result)
// 	     sort.Slice(expected, ...)
// 	     sort.Slice(result, ...)
// 	     for i := range result {
// 		     assert.Equal(t, expected[i], result[i])
// 	     }
// ✅	assert.ElementsMatch(t, expected, result)
//
// Аналогично, тест по умолчанию считается упавшим в случае паники,
// но использование assert.NotPanics() помогает будущему читателю теста понять, что вы проверяете именно её отсутствие.

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
	emptyAppID = 0
	appID      = 1             // ID приложения, которое мы создали миграцией
	appSecret  = "test-secret" // Секретный ключ приложения

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

// Что пользователь может сделать не так?
// Например, он может попытаться зарегистрироваться несколько раз с одинаковым логином.
// Наше приложение не должно такое позволять, и должно отвечать правильной ошибкой.
// И уж тем более, оно не должно падать с паникой, например.
// Тест может выглядеть, например, так:
func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	//первая попытка должна быть успешной
	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.NoError(t, err)                  // ошибки быть не должно
	require.NotEmpty(t, respReg.GetUserId()) // UserID должен быть не пустым

	// Вторая попытка - должен быть фейл
	respReg, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.Error(t, err)                               //должна быть ошибка
	assert.Empty(t, respReg.GetUserId())                //user_id должен быть пуст
	assert.ErrorContains(t, err, "user already exists") //текст ошибки должен содержать
}

// Также пользователь может присылать неверные данные на вход.
// Вариаций подобных кейсов довольно много, а проверки примерно одинаковые,
// поэтому мы напишем табличные тесты (если не знакомы с ними, очень советую ознакомиться).
// Начнём с регистрации:
func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Register with empty Password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password is required",
		},
		{
			name:        "Register with empty Email",
			email:       "",
			password:    randomFakePassword(),
			expectedErr: "email is required",
		},
		{
			name:        "Register with both empty",
			email:       "",
			password:    "",
			expectedErr: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		appID       int32
		expectedErr string
	}{
		{
			name:        "Login with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			appID:       appID,
			expectedErr: "password is required",
		},
		{
			name:        "Login with Empty Email",
			email:       "",
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "email is required",
		},
		{
			name:        "Login with Both Empty Email and Password",
			email:       "",
			password:    "",
			appID:       appID,
			expectedErr: "email is required",
		},
		{
			name:        "Login with Non-Matching Password",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "invalid email or password",
		},
		{
			name:        "Login without AppID",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appID:       emptyAppID,
			expectedErr: "app_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    gofakeit.Email(),
				Password: randomFakePassword(),
			})
			require.NoError(t, err)

			_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

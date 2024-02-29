// internal/lib/jwt/jwt.go
package jwt

import (
	"grpc-service-ref/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// NewToken creates new JWT token for given user app
func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodPS256.SigningMethodRSA)

	//добавляем в токен всю необходимую информацию
	claims := token.Claims.(jwt.MapClaims) //утверждение типа интерфейса. Проверямый тип - jwt.MapClaims, значение token.Claims. Это что-то типа преобразования типа
	claims["uid"] = user.ID
	claims["email"] = user.Email
	//В ней мы задаём срок действия (TTL) токена в виде конкретной временной метки, до которой он будет считаться валидным. После этого дедлайна токен будет считаться "протухшим", на стороне клиента мы его не будем принимать.
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["app_id"] = app.ID

	//подписываем токен, используя секретный ключ приложения
	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", jwt.ErrECDSAVerification
	}

	return tokenString, nil
}

package services

import (
	"errors"
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

// Определяем пользовательские ошибки для обработки JWT.
var (
	ErrTokenIsInvalid= errors.New("токен недействителен")
	ErrTokenIsExpired= errors.New("токен истёк")
)

// JWTService представляет сервис для работы с JWT токенами.
type JWTService struct {
	authSecretKey string // Секретный ключ, используемый для подписи и валидации токенов
}

// NewJWTService создает новый экземпляр JWTService с заданным секретным ключом.
func NewJWTService(authSecretKey string) *JWTService {
	return &JWTService{authSecretKey: authSecretKey}
}

// GenerateJWT генерирует JWT токен для указанного субъекта с заданным временем истечения (24 часа).
func (j *JWTService) GenerateJWT(subject string) (string, error) {
	now := time.Now() // Текущее время для метки времени токена

	// Создаем токен с подписью HMAC и стандартными полями
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": subject,             // Субъект токена (обычно идентификатор пользователя)
		"iat": now.Unix(),          // Время выпуска токена (issued at)
		"exp": now.Add(24 * time.Hour).Unix(), // Время истечения токена (expires at)
	})

	// Подписываем токен секретным ключом
	tokenString, err := token.SignedString([]byte(j.authSecretKey))
	if err != nil {
		return "", fmt.Errorf("error while generating token: %w", err) // Обрабатываем ошибку генерации токена
	}

	return tokenString, nil
}

// ValidateToken проверяет валидность и срок действия JWT токена.
func (j *JWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
	// Разбираем и проверяем токен с учетом кастомных claims (здесь используется стандартный набор claims).
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что метод подписи является HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"]) // Ошибка при неверном методе подписи
		}
		return []byte(j.authSecretKey), nil // Возвращаем секретный ключ для валидации подписи
	})

	// Обрабатываем возможные ошибки валидации токена
	if err != nil {
		// Проверяем на конкретные ошибки истечения срока действия токена
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenIsExpired
		}

		return nil, fmt.Errorf("error while validating token: %w", err) // Общая ошибка валидации
	}

	// Проверяем валидность токена
	if !parsedToken.Valid {
		return nil, ErrTokenIsInvalid // Ошибка, если токен недействителен
	}

	return parsedToken, nil // Возвращаем разобранный и проверенный токен
}

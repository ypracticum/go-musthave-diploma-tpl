package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/models"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/services"
)

// userFieldType определяет тип для ключа, используемого для хранения данных пользователя в контексте.
type userFieldType string

// userField является ключом для хранения информации о пользователе в контексте запроса.
const userField userFieldType = "userField"

// AuthMiddlewareConfig представляет конфигурацию middleware для аутентификации.
type AuthMiddlewareConfig struct {
	excludePaths []string // Пути, которые будут исключены из проверки аутентификации.
}

// AuthMiddleware создает новую конфигурацию middleware для аутентификации.
func AuthMiddleware() *AuthMiddlewareConfig {
	return &AuthMiddlewareConfig{}
}

// WithExcludedPaths устанавливает пути, которые будут исключены из проверки аутентификации.
func (a *AuthMiddlewareConfig) WithExcludedPaths(paths ...string) *AuthMiddlewareConfig {
	a.excludePaths = paths
	return a
}

// Middleware возвращает middleware для аутентификации, используя установленную конфигурацию.
func (a *AuthMiddlewareConfig) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, является ли текущий путь исключенным из проверки аутентификации.
		for _, path := range a.excludePaths {
			if strings.HasPrefix(r.URL.Path, path) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Извлекаем сервисы аутентификации и JWT из контекста запроса.
		authService := GetServiceFromContext[models.AuthService](w, r, AuthServiceKey)
		jwtService := GetServiceFromContext[models.JWTService](w, r, JwtServiceKey)

		// Получаем заголовок Authorization.
		authHeader := r.Header.Get("Authorization")

		// Проверяем наличие заголовка Authorization.
		if authHeader == "" {
			http.Error(w, "Требуется заголовок Authorization", http.StatusUnauthorized)
			return
		}

		// Извлекаем токен из заголовка Authorization.
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			http.Error(w, "Токен Bearer пуст", http.StatusUnauthorized)
			return
		}

		// Валидируем токен с помощью JWT-сервиса.
		token, err := (*jwtService).ValidateToken(tokenString)
		if err != nil {
			// Обрабатываем различные ошибки валидации токена.
			if errors.Is(err, services.ErrTokenIsInvalid) {
				http.Error(w, "Неверный токен", http.StatusUnauthorized)
				return
			}

			if errors.Is(err, services.ErrTokenIsExpired) {
				http.Error(w, "Токен истёк", http.StatusUnauthorized)
				return
			}

			// Обрабатываем другие ошибки при валидации токена.
			http.Error(w, fmt.Sprintf("Произошла ошибка при проверке токена: %s", err.Error()), http.StatusUnauthorized)
			return
		}

		// Извлекаем логин пользователя из токена.
		login, err := token.Claims.GetSubject()
		if err != nil {
			http.Error(w, fmt.Sprintf("Произошла ошибка при чтении поля sub: %s", err.Error()), http.StatusUnauthorized)
			return
		}

		// Получаем пользователя из базы данных по логину.
		user, err := (*authService).GetUser(r.Context(), login)
		if err != nil {
			// Обрабатываем ошибку, если пользователь не найден.
			if errors.Is(err, services.ErrUserIsNotExist) {
				http.Error(w, fmt.Sprintf("Пользователь с логином %s не существует", login), http.StatusConflict)
				return
			}

			// Обрабатываем другие ошибки при получении пользователя.
			http.Error(w, fmt.Sprintf("Произошла ошибка при проверке логина пользователя: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// Добавляем информацию о пользователе в контекст запроса и передаем управление следующему обработчику.
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userField, user)))
	})
}

// GetUserFromContext извлекает информацию о пользователе из контекста запроса.
// В случае ошибки возвращает HTTP 500 и nil.
func GetUserFromContext(w http.ResponseWriter, r *http.Request) *models.User {
	user, ok := r.Context().Value(userField).(*models.User)

	if !ok {
		http.Error(w, "Не удалось получить пользователя из контекста", http.StatusInternalServerError)
		return nil
	}

	return user
}

package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/middlewares"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/models"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/services"
)

// IsUnknownUserDataValid проверяет, что в данных пользователя присутствуют логин и пароль.
func IsUnknownUserDataValid(data models.UnknownUser) bool {
	return data.Login != nil && data.Password != nil
}

// Register обрабатывает запрос на регистрацию нового пользователя и возвращает JWT токен при успешной регистрации.
func Register(w http.ResponseWriter, r *http.Request) {
	// Извлекаем данные пользователя из тела запроса.
	data := middlewares.GetParsedJSONData[models.UnknownUser](w, r)

	// Получаем сервисы аутентификации и JWT из контекста запроса.
	authService := middlewares.GetServiceFromContext[models.AuthService](w, r, middlewares.AuthServiceKey)
	jwtService := middlewares.GetServiceFromContext[models.JWTService](w, r, middlewares.JwtServiceKey)

	// Проверяем, что запрос содержит логин и пароль.
	if !IsUnknownUserDataValid(data) {
		http.Error(w, "Запрос не содержит логин или пароль", http.StatusBadRequest)
		return
	}

	// Генерируем JWT токен.
	token, err := (*jwtService).GenerateJWT(*data.Login)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при генерации JWT токена: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Регистрируем пользователя.
	if err := (*authService).Register(r.Context(), data); err != nil {
		// Обрабатываем ошибку, если пользователь уже зарегистрирован.
		if errors.Is(err, services.ErrUserIsAlreadyRegistered) {
			http.Error(w, "Пользователь уже зарегистрирован", http.StatusConflict)
			return
		}

		// Обрабатываем другие ошибки при регистрации.
		http.Error(w, fmt.Sprintf("Произошла ошибка при регистрации: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Устанавливаем токен в заголовок ответа.
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
}

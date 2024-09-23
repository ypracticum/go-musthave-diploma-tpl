package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/middlewares"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/models"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/services"
)

// Login обрабатывает запрос на вход пользователя и возвращает JWT токен при успешной авторизации.
func Login(w http.ResponseWriter, r *http.Request) {
	// Извлекаем данные пользователя из тела запроса.
	data := middlewares.GetParsedJSONData[models.UnknownUser](w, r)

	// Получаем сервисы аутентификации и JWT из контекста запроса.
	authService := middlewares.GetServiceFromContext[models.AuthService](w, r, middlewares.AuthServiceKey)
	jwtService := middlewares.GetServiceFromContext[models.JWTService](w, r, middlewares.JwtServiceKey)

	// Проверяем, что запрос содержит логин и пароль.
	if ok := IsUnknownUserDataValid(data); !ok {
		http.Error(w, "Запрос не содержит логин или пароль", http.StatusBadRequest)
		return
	}

	// Пытаемся аутентифицировать пользователя.
	if err := (*authService).Login(r.Context(), data); err != nil {
		// Обрабатываем ошибку, если пользователь не существует.
		if errors.Is(err, services.ErrUserIsNotExist) {
			http.Error(w, fmt.Sprintf("Пользователь с логином %s не существует", *data.Login), http.StatusUnauthorized)
			return
		}

		// Обрабатываем ошибку, если пароль неверный.
		if errors.Is(err, services.ErrPasswordIsIncorrect) {
			http.Error(w, "Неверный пароль", http.StatusUnauthorized)
			return
		}

		// Обрабатываем другие ошибки при аутентификации.
		http.Error(w, fmt.Sprintf("Произошла ошибка при входе: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Генерируем JWT токен для успешной аутентификации.
	token, err := (*jwtService).GenerateJWT(*data.Login)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при генерации JWT токена: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Устанавливаем токен в заголовок ответа.
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
}

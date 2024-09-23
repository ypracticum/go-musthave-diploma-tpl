package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/models"
)

// Тип для ключей контекста, используемых для хранения сервисов.
type key int

// Константы для ключей различных сервисов, которые будут храниться в контексте.
const (
	AuthServiceKey key = iota
	JwtServiceKey
	OrderServiceKey
	AccrualServiceKey
	BalanceServiceKey
)

// ServiceInjectorMiddleware - middleware для инъекции сервисов в контекст запроса.
// Принимает на вход необходимые сервисы и добавляет их в контекст.
func ServiceInjectorMiddleware(
	authService models.AuthService,
	jwtService models.JWTService,
	orderService models.OrderService,
	accrualService models.AccrualService,
	balanceService models.BalanceService,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Добавляем сервисы в контекст запроса
			ctx := context.WithValue(r.Context(), AuthServiceKey, authService)
			ctx = context.WithValue(ctx, JwtServiceKey, jwtService)
			ctx = context.WithValue(ctx, OrderServiceKey, orderService)
			ctx = context.WithValue(ctx, AccrualServiceKey, accrualService)
			ctx = context.WithValue(ctx, BalanceServiceKey, balanceService)

			// Передаем управление следующему обработчику с обновленным контекстом
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetServiceFromContext - извлекает сервис из контекста по заданному ключу.
// В случае ошибки возвращает HTTP 500 и сообщение об ошибке.
func GetServiceFromContext[Service interface{}](w http.ResponseWriter, r *http.Request, serviceKey key) *Service {
	// Извлекаем сервис из контекста по ключу
	foundService, ok := r.Context().Value(serviceKey).(Service)

	if !ok {
		// Если сервис не найден, возвращаем ошибку
		http.Error(w, fmt.Sprintf("Сервис не найден в контексте по ключу %v", serviceKey), http.StatusInternalServerError)
		return nil
	}

	return &foundService
}

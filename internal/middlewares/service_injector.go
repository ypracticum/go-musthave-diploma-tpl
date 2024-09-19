package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/models"
)

type key int

const (
	AuthServiceKey key = iota
	JwtServiceKey
	OrderServiceKey
	AccrualServiceKey
	BalanceServiceKey
)

func ServiceInjectorMiddleware(
	authService models.AuthService,
	jwtService models.JWTService,
	orderService models.OrderService,
	accrualService models.AccrualService,
	balanceService models.BalanceService,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), AuthServiceKey, authService)
			ctx = context.WithValue(ctx, JwtServiceKey, jwtService)
			ctx = context.WithValue(ctx, OrderServiceKey, orderService)
			ctx = context.WithValue(ctx, AccrualServiceKey, accrualService)
			ctx = context.WithValue(ctx, BalanceServiceKey, balanceService)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetServiceFromContext[Service interface{}](w http.ResponseWriter, r *http.Request, serviceKey key) *Service {
	foundService, ok := r.Context().Value(serviceKey).(Service)

	if !ok {
		http.Error(w, fmt.Sprintf("Service wasn't found in context by key %v", serviceKey), http.StatusInternalServerError)
		return nil
	}

	return &foundService
}

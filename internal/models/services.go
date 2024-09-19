package models

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

//go:generate mockgen -destination=mocks/mock_auth.go . AuthService
type AuthService interface {
	Register(ctx context.Context, user UnknownUser) error

	Login(ctx context.Context, user UnknownUser) error

	GetUser(ctx context.Context, login string) (*User, error)
}

//go:generate mockgen -destination=mocks/mock_jwt.go . JWTService
type JWTService interface {
	GenerateJWT(subject string) (string, error)

	ValidateToken(token string) (*jwt.Token, error)
}

//go:generate mockgen -destination=mocks/mock_order.go . OrderService
type OrderService interface {
	VerifyOrderID(orderID string) bool

	CreateOrder(ctx context.Context, orderID, userID string) error

	GetOrders(ctx context.Context, userID string) ([]Order, error)
}

//go:generate mockgen -destination=mocks/mock_accrual.go . AccrualService
type AccrualService interface {
	CalculateAccrual(orderID string)

	StartCalculationAccruals(ctx context.Context) error
}

//go:generate mockgen -destination=mocks/mock_balance.go . BalanceService
type BalanceService interface {
	GetUserBalance(ctx context.Context, userID string) (Balance, error)

	CreateWithdrawal(ctx context.Context, orderID, userID string, amount float64) error

	GetWithdrawalFlow(ctx context.Context, userID string) ([]WithdrawalFlowItem, error)
}

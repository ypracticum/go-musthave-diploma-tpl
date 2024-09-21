package services

import (
	"context"
	"errors"
	"sort"
	"strconv"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/database"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/models"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/utils"
)

var (
	ErrDuplicateOrder               = errors.New("order is duplicated")
	ErrDuplicateOrderByOriginalUser = errors.New("order is duplicated by the same user")
)

type OrderService struct {
	storage orderStorage
}

type orderStorage interface {
	CreateOrder(ctx context.Context, orderID string, userID string) error

	FindOrder(ctx context.Context, orderID string) (*database.OrderDB, error)

	FindOrdersWithAccrual(ctx context.Context, userID string) (*[]database.OrderWithAccrualDB, error)
}

func NewOrderService(storage orderStorage) *OrderService {
	return &OrderService{storage}
}

func (o *OrderService) VerifyOrderID(orderID string) bool {
	var sum int
	var alternate bool

	for i := len(orderID) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(orderID[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2

			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

func (o *OrderService) CreateOrder(ctx context.Context, orderID, userID string) error {
	if err := o.storage.CreateOrder(ctx, orderID, userID); err != nil {
		if !errors.Is(err, database.ErrDuplicateOrder) {
			return err
		}

		order, errOrder := o.storage.FindOrder(ctx, orderID)

		if errOrder != nil {
			return errOrder
		}

		if order.UserID == userID {
			return ErrDuplicateOrderByOriginalUser
		}

		return ErrDuplicateOrder
	}

	return nil
}

func (o *OrderService) GetOrders(ctx context.Context, userID string) ([]models.Order, error) {
	orders, err := o.storage.FindOrdersWithAccrual(ctx, userID)

	if err != nil {
		return []models.Order{}, err
	}

	if orders == nil {
		return []models.Order{}, nil
	}

	result := make([]models.Order, len(*orders))

	for i, order := range *orders {
		accrual := order.Accrual
		result[i] = models.Order{
			ID:         order.ID,
			Status:     order.Status.OrderStatus,
			UploadedAt: utils.RFC3339Date{Time: order.UploadedAt},
			Accrual:    &accrual,
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].UploadedAt.Time.Before(result[j].UploadedAt.Time)
	})

	return result, nil
}

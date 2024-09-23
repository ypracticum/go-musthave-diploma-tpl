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

// Определяем ошибки, связанные с заказами
var (
	ErrDuplicateOrder               = errors.New("заказ уже существует")                     // Ошибка для дублирующего заказа
	ErrDuplicateOrderByOriginalUser = errors.New("заказ уже был создан этим пользователем") // Ошибка для повторного заказа от того же пользователя
)

// OrderService представляет сервис для работы с заказами
type OrderService struct {
	storage orderStorage // Хранилище данных для работы с заказами
}

// Интерфейс хранилища для работы с заказами
type orderStorage interface {
	CreateOrder(ctx context.Context, orderID string, userID string) error
	FindOrder(ctx context.Context, orderID string) (*database.OrderDB, error)
	FindOrdersWithAccrual(ctx context.Context, userID string) (*[]database.OrderWithAccrualDB, error)
}

// NewOrderService создает новый экземпляр OrderService
func NewOrderService(storage orderStorage) *OrderService {
	return &OrderService{storage: storage}
}

// VerifyOrderID проверяет корректность номера заказа с использованием алгоритма Луна
func (o *OrderService) VerifyOrderID(orderID string) bool {
	var sum int
	var alternate bool

	// Идем по цифрам заказа с конца
	for i := len(orderID) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(orderID[i]))
		if err != nil {
			return false // Если не удается преобразовать символ в число, возвращаем false
		}

		if alternate {
			digit *= 2 // Удваиваем каждую вторую цифру

			if digit > 9 {
				digit -= 9 // Если результат больше 9, вычитаем 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0 // Проверяем, делится ли сумма на 10
}

// CreateOrder создает новый заказ и проверяет, не существует ли уже такой заказ
func (o *OrderService) CreateOrder(ctx context.Context, orderID, userID string) error {
	// Пытаемся создать заказ
	if err := o.storage.CreateOrder(ctx, orderID, userID); err != nil {
		// Если ошибка не связана с дублированием заказа, возвращаем ее
		if !errors.Is(err, database.ErrDuplicateOrder) {
			return err
		}

		// Если заказ уже существует, ищем его данные
		order, errOrder := o.storage.FindOrder(ctx, orderID)
		if errOrder != nil {
			return errOrder
		}

		// Если заказ был создан тем же пользователем, возвращаем соответствующую ошибку
		if order.UserID == userID {
			return ErrDuplicateOrderByOriginalUser
		}

		// Иначе возвращаем ошибку дублирования заказа
		return ErrDuplicateOrder
	}

	return nil
}

// GetOrders возвращает список заказов пользователя, отсортированный по дате загрузки
func (o *OrderService) GetOrders(ctx context.Context, userID string) ([]models.Order, error) {
	// Ищем заказы с информацией о начислениях
	orders, err := o.storage.FindOrdersWithAccrual(ctx, userID)
	if err != nil {
		return []models.Order{}, err
	}

	// Если заказы не найдены, возвращаем пустой список
	if orders == nil {
		return []models.Order{}, nil
	}

	// Преобразуем данные из базы в формат моделей
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

	// Сортируем заказы по дате загрузки
	sort.Slice(result, func(i, j int) bool {
		return result[i].UploadedAt.Time.Before(result[j].UploadedAt.Time)
	})

	return result, nil
}

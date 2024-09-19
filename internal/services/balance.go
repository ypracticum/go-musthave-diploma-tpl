package services

import (
	"context"
	"errors"
	"sort"

	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/database"
	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/models"
	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/utils"
)

// BalanceService предоставляет методы для управления балансом пользователя
type BalanceService struct {
	storage balanceStorage
}

// balanceStorage определяет интерфейс для работы с хранилищем баланса и операций
type balanceStorage interface {
	FindAccrualFlow(ctx context.Context, userID string) ([]database.AccrualFlowItemDB, error)
	CreateWithdrawal(ctx context.Context, orderID, userID string, amount float64) error
	FindWithdrawalFlow(ctx context.Context, userID string) ([]database.WithdrawalFlowItemDB, error)
}

// NewBalanceService создает новый экземпляр BalanceService с заданным хранилищем
func NewBalanceService(storage balanceStorage) *BalanceService {
	return &BalanceService{storage: storage}
}

// GetUserBalance возвращает текущий баланс пользователя с учетом всех операций
func (b *BalanceService) GetUserBalance(ctx context.Context, userID string) (models.Balance, error) {
	// Получение информации о начислениях
	accrualFlow, err := b.storage.FindAccrualFlow(ctx, userID)
	if err != nil {
		return models.Balance{}, wrapError("не удалось получить данные о начислениях", err)
	}

	// Получение информации о выводах средств
	withdrawalFlow, err := b.storage.FindWithdrawalFlow(ctx, userID)
	if err != nil {
		return models.Balance{}, wrapError("не удалось получить данные о выводах", err)
	}

	var current float64
	var withdrawn float64

	// Подсчет начислений
	for _, item := range accrualFlow {
		current += item.Amount
	}

	// Подсчет выведенных средств
	for _, item := range withdrawalFlow {
		withdrawn += item.Amount
	}

	// Возврат результата
	return models.Balance{
		Current:   current - withdrawn,
		Withdrawn: withdrawn,
	}, nil
}

// CreateWithdrawal создает запись о выводе средств
func (b *BalanceService) CreateWithdrawal(ctx context.Context, orderID, userID string, amount float64) error {
	// Валидация ввода
	if orderID == "" || userID == "" || amount <= 0 {
		return errors.New("недействительные параметры вывода")
	}

	// Создание записи о выводе средств
	err := b.storage.CreateWithdrawal(ctx, orderID, userID, amount)
	if err != nil {
		return wrapError("не удалось создать запись о выводе средств", err)
	}

	return nil
}

// GetWithdrawalFlow возвращает список выводов средств пользователя, отсортированный по времени
func (b *BalanceService) GetWithdrawalFlow(ctx context.Context, userID string) ([]models.WithdrawalFlowItem, error) {
	// Получение данных о выводах
	withdrawalFlow, err := b.storage.FindWithdrawalFlow(ctx, userID)
	if err != nil {
		return nil, wrapError("не удалось получить данные о выводах", err)
	}

	// Если данных нет, возвращаем пустой список
	if len(withdrawalFlow) == 0 {
		return []models.WithdrawalFlowItem{}, nil
	}

	// Преобразование данных в модель и сортировка
	result := make([]models.WithdrawalFlowItem, len(withdrawalFlow))
	for i, item := range withdrawalFlow {
		result[i] = models.WithdrawalFlowItem{
			OrderID:     item.OrderID,
			Sum:         item.Amount,
			ProcessedAt: utils.RFC3339Date{Time: item.ProcessedAt},
		}
	}

	// Сортировка по дате обработки
	sort.Slice(result, func(i, j int) bool {
		return result[i].ProcessedAt.Time.Before(result[j].ProcessedAt.Time)
	})

	return result, nil
}

// wrapError добавляет описание к ошибке
func wrapError(message string, err error) error {
	if err != nil {
		return errors.New(message + ": " + err.Error())
	}
	return errors.New(message)
}

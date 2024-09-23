package services

import (
	"context"
	"sort"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/database"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/models"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/utils"
)

type BalanceService struct {
	storage balanceStorage
}

type balanceStorage interface {
	FindAccrualFlow(ctx context.Context, userID string) (*[]database.AccrualFlowItemDB, error)

	CreateWithdrawal(ctx context.Context, orderID, userID string, amount float64) error

	FindWithdrawalFlow(ctx context.Context, userID string) (*[]database.WithdrawalFlowItemDB, error)
}

func NewBalanceService(storage balanceStorage) *BalanceService {
	return &BalanceService{storage: storage}
}

func (b *BalanceService) GetUserBalance(ctx context.Context, userID string) (models.Balance, error) {
	accrualFlow, err := b.storage.FindAccrualFlow(ctx, userID)

	if err != nil {
		return models.Balance{}, err
	}

	withdrawalFlow, err := b.storage.FindWithdrawalFlow(ctx, userID)

	if err != nil {
		return models.Balance{}, err
	}

	var current float64 = 0
	var withdrawn float64 = 0

	if accrualFlow != nil {
		for _, item := range *accrualFlow {
			current += item.Amount
		}
	}

	if withdrawalFlow != nil {
		for _, item := range *withdrawalFlow {
			withdrawn += item.Amount
		}
	}

	return models.Balance{Current: current - withdrawn, Withdrawn: withdrawn}, nil
}

func (b *BalanceService) CreateWithdrawal(ctx context.Context, orderID, userID string, amount float64) error {
	if err := b.storage.CreateWithdrawal(ctx, orderID, userID, amount); err != nil {
		return err
	}

	return nil
}

func (b *BalanceService) GetWithdrawalFlow(ctx context.Context, userID string) ([]models.WithdrawalFlowItem, error) {
	withdrawalFlow, err := b.storage.FindWithdrawalFlow(ctx, userID)

	if err != nil {
		return []models.WithdrawalFlowItem{}, err
	}

	if withdrawalFlow == nil {
		return []models.WithdrawalFlowItem{}, nil
	}

	result := make([]models.WithdrawalFlowItem, len(*withdrawalFlow))

	for i, item := range *withdrawalFlow {
		result[i] = models.WithdrawalFlowItem{
			OrderID:     item.OrderID,
			Sum:         item.Amount,
			ProcessedAt: utils.RFC3339Date{Time: item.ProcessedAt},
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ProcessedAt.Time.Before(result[j].ProcessedAt.Time)
	})

	return result, nil
}

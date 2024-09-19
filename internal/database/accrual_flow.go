package database

import (
	"context"
	"fmt"
	"time"
)

const (
	// InsertAccrualQuery используется для вставки новой записи начисления
	InsertAccrualQuery = `
		INSERT INTO
			accrual_flow (order_id, amount)
		VALUES ($1, $2)
	`

	// SelectAccrualFlowQuery используется для выборки записей начислений по user_id
	SelectAccrualFlowQuery = `
		SELECT
			af.order_id,
			af.amount,
			af.processed_at
		FROM
			accrual_flow af
		LEFT JOIN orders o ON af.order_id = o.id
		WHERE
			o.user_id = $1
	`
)

// AccrualFlowItemDB представляет собой структуру данных для записи начисления из базы данных
type AccrualFlowItemDB struct {
	OrderID     string    // Идентификатор заказа
	Amount      float64   // Сумма начисления
	ProcessedAt time.Time // Время обработки начисления
}

// CreateAccrual добавляет новое начисление в базу данных
func (d *Database) CreateAccrual(ctx context.Context, orderID string, amount float64) error {
	// Выполняем SQL-запрос для вставки нового начисления
	_, err := d.db.Exec(ctx, InsertAccrualQuery, orderID, amount)
	if err != nil {
		return fmt.Errorf("не удалось создать начисление: %w", err)
	}

	return nil
}

// FindAccrualFlow возвращает список начислений для указанного пользователя	
func (d *Database) FindAccrualFlow(ctx context.Context, userID string) (*[]AccrualFlowItemDB, error) {
	var result []AccrualFlowItemDB

	// Выполняем SQL-запрос для выборки начислений по user_id
	rows, err := d.db.Query(ctx, SelectAccrualFlowQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос начислений: %w", err)
	}
	defer rows.Close()
	// Проходим по всем строкам результата запроса
	for rows.Next() {
		var item AccrualFlowItemDB

		// Сканируем значения из текущей строки в структуру AccrualFlowItemDB
		if err := rows.Scan(&item.OrderID, &item.Amount, &item.ProcessedAt); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки начислений: %w", err)
		}

		// Добавляем запись в результирующий срез
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после чтения строк начислений: %w", err)
	}


	return &result, nil
}

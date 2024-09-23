package database

import (
	"context"
	"fmt"
	"time"
)

// SQL-запросы для работы с выводом средств
const (
	// InsertWithdrawalQuery используется для вставки новой записи вывода средств
	InsertWithdrawalQuery = `
		INSERT INTO
			withdrawal_flow (order_id, user_id, amount)
		VALUES ($1, $2, $3)
	`

	// SelectWithdrawalFlowQuery используется для выборки записей вывода средств по user_id
	SelectWithdrawalFlowQuery = `
		SELECT
			order_id,
			amount,
			processed_at
		FROM
			withdrawal_flow
		WHERE
			user_id = $1
	`
)

// WithdrawalFlowItemDB представляет собой структуру данных для записи вывода средств из базы данных
type WithdrawalFlowItemDB struct {
	OrderID     string    // Идентификатор заказа
	Amount      float64   // Сумма вывода средств
	ProcessedAt time.Time // Время обработки вывода средств
}

// CreateWithdrawal добавляет новый вывод средств в базу данных
func (d *Database) CreateWithdrawal(ctx context.Context, orderID, userID string, amount float64) error {
	// Выполняем SQL-запрос для вставки нового вывода средств
	_, err := d.db.Exec(ctx, InsertWithdrawalQuery, orderID, userID, amount)
	if err != nil {
		return fmt.Errorf("не удалось создать вывод средств: %w", err)
	}

	return nil
}

// FindWithdrawalFlow возвращает список выводов средств для указанного пользователя
func (d *Database) FindWithdrawalFlow(ctx context.Context, userID string) (*[]WithdrawalFlowItemDB, error) {
	var result []WithdrawalFlowItemDB

	// Выполняем SQL-запрос для выборки выводов средств по user_id
	rows, err := d.db.Query(ctx, SelectWithdrawalFlowQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос вывода средств: %w", err)
	}
	defer rows.Close()

	// Проходим по всем строкам результата запроса
	for rows.Next() {
		var item WithdrawalFlowItemDB

		// Сканируем значения из текущей строки в структуру WithdrawalFlowItemDB
		if err := rows.Scan(&item.OrderID, &item.Amount, &item.ProcessedAt); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки вывода средств: %w", err)
		}

		// Добавляем запись в результирующий срез
		result = append(result, item)
	}

	// Проверяем наличие ошибок после итерации по строкам
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после чтения строк вывода средств: %w", err)
	}

	return  &result, nil
}

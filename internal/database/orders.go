package database

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Определение пользовательских ошибок
var (
	ErrDuplicateOrder = errors.New("заказ уже существует") // Ошибка дублирования заказа
)

// SQL-запросы для работы с заказами
const (
	InsertOrderQuery = `
		INSERT INTO
			orders (id, user_id)
		VALUES ($1, $2)
	`
	SelectOrderQuery = `
		SELECT
		    id,
			user_id,
			status,
			uploaded_at
		FROM
		    orders
		WHERE
		    id = $1
	`
	SelectOrdersWithAccrualQuery = `
		SELECT
			o.id,
			user_id,
			status,
			uploaded_at,
			SUM(coalesce(amount, 0))
		FROM
		    orders o
			LEFT JOIN accrual_flow af ON o.id = af.order_id
		WHERE
		    user_id = $1
		GROUP BY 
		    o.id
	`
	UpdateOrderStatusQuery = `
		UPDATE
			orders
		SET
			status = $2
		WHERE
		    id = $1
	`
	SelectAllUnprocessedOrdersQuery = `
		SELECT
			id,
			user_id,
			status,
			uploaded_at
		FROM
		    orders
		WHERE
		    status NOT IN ('INVALID', 'PROCESSED')
	`
)

// Структура для хранения информации о заказе
type OrderDB struct {
	ID         string        // Идентификатор заказа
	UserID     string        // Идентификатор пользователя
	Status     OrderStatusDB // Статус заказа
	UploadedAt time.Time     // Дата и время загрузки
}

// Структура для заказа с информацией о начислениях
type OrderWithAccrualDB struct {
	OrderDB
	Accrual float64 // Сумма начислений по заказу
}

// Определение статуса заказа с возможностью преобразования в/из базы данных
type OrderStatusDB struct {
	models.OrderStatus
}

// Реализация интерфейса sql.Scanner для чтения статуса заказа из базы данных
func (s *OrderStatusDB) Scan(value interface{}) error {
	strVal, ok := value.(string)
	if !ok {
		return fmt.Errorf("статус заказа должен быть строкой, а не %T", value)
	}

	*s = OrderStatusDB{models.OrderStatus(strVal)}
	return nil
}

// Реализация интерфейса driver.Valuer для преобразования статуса заказа в строку перед записью в базу данных
func (s OrderStatusDB) Value() (driver.Value, error) {
	return string(s.OrderStatus), nil
}

// Создание нового заказа
func (d *Database) CreateOrder(ctx context.Context, orderID, userID string) error {
	_, err := d.db.Exec(ctx, InsertOrderQuery, orderID, userID)
	if err != nil {
		var e *pgconn.PgError
		// Проверяем, не является ли ошибка нарушением уникальности (дубликат заказа)
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			return ErrDuplicateOrder
		}
		return fmt.Errorf("ошибка создания заказа: %w", err)
	}

	return nil
}

// Поиск заказа по его ID
func (d *Database) FindOrder(ctx context.Context, orderID string) (*OrderDB, error) {
	order := &OrderDB{}

	err := d.db.QueryRow(ctx, SelectOrderQuery, orderID).
		Scan(&order.ID, &order.UserID, &order.Status, &order.UploadedAt)
	if err != nil {
		// Если заказ не найден, возвращаем nil без ошибки
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка поиска заказа: %w", err)
	}

	return order, nil
}

// Поиск всех заказов пользователя с начислениями
func (d *Database) FindOrdersWithAccrual(ctx context.Context, userID string) (*[]OrderWithAccrualDB, error) {
	var result []OrderWithAccrualDB

	rows, err := d.db.Query(ctx, SelectOrdersWithAccrualQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска заказов с начислениями: %w", err)
	}
	defer rows.Close()

	// Обрабатываем каждую строку результата
	for rows.Next() {
		var item OrderWithAccrualDB
		if err := rows.Scan(&item.ID, &item.UserID, &item.Status, &item.UploadedAt, &item.Accrual); err != nil {
			return nil, fmt.Errorf("ошибка обработки строки с заказом: %w", err)
		}
		result = append(result, item)
	}

	// Проверка на ошибки при итерации по строкам
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по строкам: %w", err)
	}

	return &result, nil
}

// Обновление статуса заказа
func (d *Database) UpdateOrderStatus(ctx context.Context, orderID string, status OrderStatusDB) error {
	_, err := d.db.Exec(ctx, UpdateOrderStatusQuery, orderID, status)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса заказа: %w", err)
	}
	return nil
}

// Поиск всех необработанных заказов
func (d *Database) FindAllUnprocessedOrders(ctx context.Context) (*[]OrderDB, error) {
	var result []OrderDB

	rows, err := d.db.Query(ctx, SelectAllUnprocessedOrdersQuery)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска необработанных заказов: %w", err)
	}
	defer rows.Close()

	// Обрабатываем каждую строку результата
	for rows.Next() {
		var item OrderDB
		if err := rows.Scan(&item.ID, &item.UserID, &item.Status, &item.UploadedAt); err != nil {
			return nil, fmt.Errorf("ошибка обработки строки с заказом: %w", err)
		}
		result = append(result, item)
	}

	// Проверка на ошибки при итерации по строкам
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по строкам: %w", err)
	}

	return &result, nil
}

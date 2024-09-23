package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/database"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/logger"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/models"
	"go.uber.org/zap"
)

// AccrualService представляет сервис для обработки начислений.
type AccrualService struct {
	storage          accrualStorage         // Хранилище для работы с данными начислений.
	jobQueueService  accrualJobQueueService // Сервис очереди задач для управления заданиями.
	externalEndpoint string                 // Внешний endpoint для получения данных начислений.
}

// Интерфейс для хранилища начислений.
type accrualStorage interface {
	UpdateOrderStatus(ctx context.Context, orderID string, status database.OrderStatusDB) error
	CreateAccrual(ctx context.Context, orderID string, amount float64) error
	FindAllUnprocessedOrders(ctx context.Context) (*[]database.OrderDB, error)
}

// Интерфейс для сервиса очереди задач начислений.
type accrualJobQueueService interface {
	Enqueue(job Job)                          // Добавление задания в очередь.
	ScheduleJob(job Job, delay time.Duration) // Планирование задания с задержкой.
	PauseAndResume(delay time.Duration)       // Приостановка и возобновление обработки заданий.
}

// Конструктор для создания нового экземпляра AccrualService.
func NewAccrualService(storage accrualStorage, jobQueueService accrualJobQueueService, externalEndpoint string) *AccrualService {
	return &AccrualService{
		storage:          storage,
		jobQueueService:  jobQueueService,
		externalEndpoint: externalEndpoint,
	}
}

// CalculateAccrual инициирует процесс расчета начислений для указанного заказа.
func (as *AccrualService) CalculateAccrual(orderID string) {
	// Добавляем задание в очередь для обработки начислений.
	as.jobQueueService.Enqueue(func(ctx context.Context) {
		data, retryAfter, err := as.fetchAccrualData(orderID)

		if err != nil {
			if errors.Is(err, errNoOrder) {
				logger.Log.Info("Заказ не зарегистрирован", zap.String("orderID", orderID))
				// TODO: Возможно, нужно добавить повторное добавление задания в очередь.
				return
			}

			logger.Log.Error("Не удалось получить данные начислений", zap.Error(err))
			// TODO: Добавить повторное добавление задания с таймаутом.
			return
		}

		if retryAfter > 0 {
			logger.Log.Info("Получен Retry-After", zap.Int("retryAfter", retryAfter), zap.String("orderID", orderID))
			as.jobQueueService.PauseAndResume(time.Second * time.Duration(retryAfter))
			as.jobQueueService.Enqueue(func(ctx context.Context) {
				as.CalculateAccrual(orderID)
			})
			logger.Log.Info("Добавлено новое задание после паузы", zap.Int("retryAfter", retryAfter), zap.String("orderID", orderID))
			return
		}

		logger.Log.Info("Получены данные начислений",
			zap.String("orderID", orderID),
			zap.String("status", string(data.Status)),
		)

		switch data.Status {
		case AccrualStatusRegistered:
			// Планируем новое задание через минуту.
			as.jobQueueService.ScheduleJob(func(ctx context.Context) {
				as.CalculateAccrual(orderID)
			}, time.Minute)
			logger.Log.Info("Добавлено запланированное задание", zap.String("orderID", orderID))

		case AccrualStatusProcessed, AccrualStatusProcessing, AccrualStatusInvalid:
			// Обновляем статус заказа.
			err := as.storage.UpdateOrderStatus(ctx, orderID, database.OrderStatusDB{OrderStatus: models.OrderStatus(data.Status)})
			if err != nil {
				logger.Log.Error("Не удалось обновить статус заказа", zap.Error(err))
				return
			}

			logger.Log.Info("Статус заказа обновлен",
				zap.String("orderID", orderID),
				zap.String("status", string(data.Status)),
			)

			// Если начисления есть, сохраняем их.
			if data.Accrual != nil {
				err := as.storage.CreateAccrual(ctx, orderID, *data.Accrual)
				if err != nil {
					logger.Log.Error("Не удалось создать начисление", zap.Error(err))
					return
				}

				logger.Log.Info("Начисление сохранено",
					zap.String("orderID", orderID),
					zap.String("status", string(data.Status)),
					zap.Float64("accrual", *data.Accrual),
				)
			}

		default:
			logger.Log.Error("Статус не определен", zap.String("status", string(data.Status)))
		}
	})
}

// StartCalculationAccruals запускает процесс расчета начислений для всех необработанных заказов.
func (as *AccrualService) StartCalculationAccruals(ctx context.Context) error {
	// Получаем все необработанные заказы из хранилища.
	orders, err := as.storage.FindAllUnprocessedOrders(ctx)
	if err != nil {
		return fmt.Errorf("ошибка при поиске необработанных заказов: %w", err)
	}

	if orders == nil {
		return nil
	}

	// Для каждого заказа инициируем расчет начислений.
	for _, order := range *orders {
		as.CalculateAccrual(order.ID)
	}

	return nil
}

// Определение возможных статусов начислений.
type accrualOrderStatus string

const (
	AccrualStatusRegistered accrualOrderStatus = "REGISTERED" // Зарегистрирован.
	AccrualStatusInvalid    accrualOrderStatus = "INVALID"    // Недействителен.
	AccrualStatusProcessing accrualOrderStatus = "PROCESSING" // В процессе обработки.
	AccrualStatusProcessed  accrualOrderStatus = "PROCESSED"  // Обработан.
)

// accrualDataResponse представляет структуру ответа от внешнего сервиса начислений.
type accrualDataResponse struct {
	ID      string             `json:"order"`             // Идентификатор заказа.
	Status  accrualOrderStatus `json:"status"`            // Статус начисления.
	Accrual *float64           `json:"accrual,omitempty"` // Сумма начисления (опционально).
}

var (
	errNoOrder = errors.New("заказ не зарегистрирован")  // Ошибка: заказ не зарегистрирован.
	errServer  = errors.New("внутренняя ошибка сервера") // Ошибка: внутренняя ошибка сервера.
)

const defaultRetryAfterDuration = 60 // Стандартная задержка повторной попытки в секундах.

// fetchAccrualData получает данные начислений для указанного заказа из внешнего сервиса.
func (as *AccrualService) fetchAccrualData(orderID string) (data *accrualDataResponse, retryAfter int, err error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/orders/%s", as.externalEndpoint, orderID), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("не удалось отправить GET-запрос: %w", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusNoContent:
		// Если статус 204 No Content, заказ не зарегистрирован.
		return nil, 0, errNoOrder
	case http.StatusTooManyRequests:
		// Если статус 429 Too Many Requests, получаем Retry-After из заголовка.
		retryAfterHeader := res.Header.Get("Retry-After")
		retryAfter, err := strconv.Atoi(retryAfterHeader)
		if err != nil {
			retryAfter = defaultRetryAfterDuration
		}
		return nil, retryAfter, nil
	case http.StatusInternalServerError:
		// Если статус 500 Internal Server Error, возвращаем соответствующую ошибку.
		return nil, 0, errServer
	}

	// Читаем тело ответа.
	var parsedData accrualDataResponse
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(res.Body); err != nil {
		return nil, 0, fmt.Errorf("не удалось прочитать тело ответа: %w", err)
	}

	// Распаковываем JSON-данные.
	if err := json.Unmarshal(buf.Bytes(), &parsedData); err != nil {
		return nil, 0, fmt.Errorf("не удалось распаковать данные JSON: %w", err)
	}

	return &parsedData, 0, nil
}

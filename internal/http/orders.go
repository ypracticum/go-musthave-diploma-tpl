package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/middlewares"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/models"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/services"
)

// CreateOrder обрабатывает HTTP-запрос на создание нового заказа.
// Метод извлекает данные из запроса, валидирует их, создает заказ и инициирует начисления.
func CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Извлекаем идентификатор заказа из данных запроса.
	orderID := middlewares.GetParsedTextData(w, r)

	// Проверяем, что идентификатор заказа не пустой.
	if len(orderID) == 0 {
		http.Error(w, "Идентификатор заказа не указан", http.StatusUnprocessableEntity)
		return
	}

	// Получаем сервисы заказа и начислений из контекста запроса.
	orderService := middlewares.GetServiceFromContext[models.OrderService](w, r, middlewares.OrderServiceKey)
	accrualService := middlewares.GetServiceFromContext[models.AccrualService](w, r, middlewares.AccrualServiceKey)

	// Проверяем валидность идентификатора заказа.
	if !(*orderService).VerifyOrderID(orderID) {
		http.Error(w, "Идентификатор заказа недействителен", http.StatusUnprocessableEntity)
		return
	}

	// Получаем информацию о текущем пользователе из контекста запроса.
	user := middlewares.GetUserFromContext(w, r)

	// Пытаемся создать новый заказ.
	err := (*orderService).CreateOrder(r.Context(), orderID, user.ID)
	if err != nil {
		// Обрабатываем ошибку дублирования заказа для того же пользователя.
		if errors.Is(err, services.ErrDuplicateOrderByOriginalUser) {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Обрабатываем ошибку дублирования заказа другим пользователем.
		if errors.Is(err, services.ErrDuplicateOrder) {
			http.Error(w, "Этот заказ уже был создан другим пользователем", http.StatusConflict)
			return
		}

		// Обрабатываем другие возможные ошибки при создании заказа.
		http.Error(w, fmt.Sprintf("Произошла ошибка при создании заказа: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Инициируем расчет начислений для созданного заказа.
	(*accrualService).CalculateAccrual(orderID)

	// Отправляем статус принятия запроса.
	w.WriteHeader(http.StatusAccepted)
}

// GetOrders обрабатывает HTTP-запрос на получение списка заказов пользователя.
// Метод извлекает идентификатор пользователя из контекста, получает список заказов и возвращает его в формате JSON.
func GetOrders(w http.ResponseWriter, r *http.Request) {
	// Получаем сервис заказа из контекста запроса.
	orderService := middlewares.GetServiceFromContext[models.OrderService](w, r, middlewares.OrderServiceKey)

	// Получаем информацию о текущем пользователе из контекста запроса.
	user := middlewares.GetUserFromContext(w, r)

	// Пытаемся получить список заказов пользователя.
	orders, err := (*orderService).GetOrders(r.Context(), user.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Произошла ошибка при получении заказов: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Если у пользователя нет заказов, возвращаем статус "Нет контента".
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	
	// Кодируем список заказов в формат JSON и отправляем в ответе.
	middlewares.EncodeJSONResponse(w, orders)
}

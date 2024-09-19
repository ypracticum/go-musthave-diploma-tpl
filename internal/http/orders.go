package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/models"
	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/services"

	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/middlewares"
)

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	orderID := middlewares.GetParsedTextData(w, r)

	if len(orderID) == 0 {
		http.Error(w, "Order id is empty", http.StatusUnprocessableEntity)
		return
	}

	orderService := middlewares.GetServiceFromContext[models.OrderService](w, r, middlewares.OrderServiceKey)
	accrualService := middlewares.GetServiceFromContext[models.AccrualService](w, r, middlewares.AccrualServiceKey)

	if !(*orderService).VerifyOrderID(orderID) {
		http.Error(w, "Order id is invalid", http.StatusUnprocessableEntity)
		return
	}

	user := middlewares.GetUserFromContext(w, r)

	if err := (*orderService).CreateOrder(r.Context(), orderID, user.ID); err != nil {
		if errors.Is(err, services.ErrDuplicateOrderByOriginalUser) {
			w.WriteHeader(http.StatusOK)
			return
		}

		if errors.Is(err, services.ErrDuplicateOrder) {
			http.Error(w, "Order was created by another user", http.StatusConflict)
			return
		}

		http.Error(w, fmt.Sprintf("Error occurred during creating order: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	(*accrualService).CalculateAccrual(orderID)

	w.WriteHeader(http.StatusAccepted)
}

func GetOrders(w http.ResponseWriter, r *http.Request) {
	orderService := middlewares.GetServiceFromContext[models.OrderService](w, r, middlewares.OrderServiceKey)
	user := middlewares.GetUserFromContext(w, r)

	orders, err := (*orderService).GetOrders(r.Context(), user.ID)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error occurred during getting orders: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	middlewares.EncodeJSONResponse(w, orders)
}

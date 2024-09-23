package router

import (
	"fmt"
	"net/http"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/middlewares"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/models"
)

// GetBalance получает баланс авторизованного пользователя
func GetBalance(w http.ResponseWriter, r *http.Request) {
	balanceService := middlewares.GetServiceFromContext[models.BalanceService](w, r, middlewares.BalanceServiceKey)
	user := middlewares.GetUserFromContext(w, r)

	balance, err := (*balanceService).GetUserBalance(r.Context(), user.ID)

	if err != nil {
		http.Error(w, fmt.Sprintf("При получении баланса произошла ошибка: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	middlewares.EncodeJSONResponse(w, balance)
}

// CreateWithdrawal создает новую выплату для авторизованного пользователя
func CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	data := middlewares.GetParsedJSONData[models.Withdrawal](w, r)

	if data.ID == nil || data.Sum == nil {
		http.Error(w, "В запросе отсутствует номер заказа или сумма", http.StatusBadRequest)
		return
	}

	if len(*data.ID) == 0 {
		http.Error(w, "Номер заказа пустой", http.StatusUnprocessableEntity)
		return
	}

	orderService := middlewares.GetServiceFromContext[models.OrderService](w, r, middlewares.OrderServiceKey)
	balanceService := middlewares.GetServiceFromContext[models.BalanceService](w, r, middlewares.BalanceServiceKey)

	if !(*orderService).VerifyOrderID(*data.ID) {
		http.Error(w, "Номер заказа неверный", http.StatusUnprocessableEntity)
		return
	}

	user := middlewares.GetUserFromContext(w, r)
	balance, err := (*balanceService).GetUserBalance(r.Context(), user.ID)

	if err != nil {
		http.Error(w, fmt.Sprintf("При получении баланса произошла ошибка: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if balance.Current < *data.Sum {
		http.Error(w, "Недостаточно средств на балансе", http.StatusPaymentRequired)
		return
	}

	if err := (*balanceService).CreateWithdrawal(r.Context(), *data.ID, user.ID, *data.Sum); err != nil {
		http.Error(w, fmt.Sprintf("При создании выплаты произошла ошибка: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals получает историю выплат авторизованного пользователя
func GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	balanceService := middlewares.GetServiceFromContext[models.BalanceService](w, r, middlewares.BalanceServiceKey)
	user := middlewares.GetUserFromContext(w, r)

	withdrawalFlow, err := (*balanceService).GetWithdrawalFlow(r.Context(), user.ID)

	if err != nil {
		http.Error(w, fmt.Sprintf("При получении истории выплат произошла ошибка: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if len(withdrawalFlow) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	middlewares.EncodeJSONResponse(w, withdrawalFlow)
}

package models

import "github.com/Renal37/go-musthave-diploma-tpl/internal/utils"

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdrawal struct {
	ID  *string  `json:"order"`
	Sum *float64 `json:"sum"`
}

type WithdrawalFlowItem struct {
	OrderID     string            `json:"order"`
	Sum         float64           `json:"sum"`
	ProcessedAt utils.RFC3339Date `json:"processed_at"`
}

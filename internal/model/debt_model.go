package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type DebtResponse struct {
	ID            uint                  `json:"id,omitempty"`
	ReferenceType string                `json:"reference_type,omitempty"`
	ReferenceCode string                `json:"reference_code,omitempty"`
	TotalAmount   float64               `json:"total_amount,omitempty"`
	PaidAmount    float64               `json:"paid_amount,omitempty"`
	DueDate       int64                 `json:"due_date,omitempty"`
	Status        string                `json:"status,omitempty"`
	Payments      []DebtPaymentResponse `json:"payments,omitempty"`
}

type CreateDebtRequest struct {
	DueDate      int64                      `json:"due_date" validate:"required"`
	DebtPayments []CreateDebtPaymentRequest `json:"debt_payments,omitempty"`
}

func (d *CreateDebtRequest) UnmarshalJSON(data []byte) error {
	type Alias CreateDebtRequest
	aux := &struct {
		DueDate interface{} `json:"due_date"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.DueDate.(type) {
	case string:
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
		d.DueDate = t.UnixMilli()
	case float64: // kalau user sudah kirim timestamp
		d.DueDate = int64(v)
	default:
		return fmt.Errorf("unsupported date type")
	}

	return nil
}

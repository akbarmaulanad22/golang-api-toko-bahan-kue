package repository

import (
	"tokobahankue/internal/entity"

	"github.com/sirupsen/logrus"
)

type DebtPaymentRepository struct {
	Repository[entity.DebtPayment]
	Log *logrus.Logger
}

func NewDebtPaymentRepository(log *logrus.Logger) *DebtPaymentRepository {
	return &DebtPaymentRepository{
		Log: log,
	}
}

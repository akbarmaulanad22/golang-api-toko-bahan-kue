package repository

import (
	"tokobahankue/internal/entity"

	"github.com/sirupsen/logrus"
)

type DebtRepository struct {
	Repository[entity.Debt]
	Log *logrus.Logger
}

func NewDebtRepository(log *logrus.Logger) *DebtRepository {
	return &DebtRepository{
		Log: log,
	}
}

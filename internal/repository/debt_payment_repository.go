package repository

import (
	"tokobahankue/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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

func (r *DebtPaymentRepository) CreateBulk(db *gorm.DB, payments []entity.DebtPayment) error {
	return db.CreateInBatches(&payments, 100).Error
}

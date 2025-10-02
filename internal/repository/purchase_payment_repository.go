package repository

import (
	"tokobahankue/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PurchasePaymentRepository struct {
	Repository[entity.PurchasePayment]
	Log *logrus.Logger
}

func NewPurchasePaymentRepository(log *logrus.Logger) *PurchasePaymentRepository {
	return &PurchasePaymentRepository{
		Log: log,
	}
}

func (r *PurchasePaymentRepository) CreateBulk(db *gorm.DB, payments []entity.PurchasePayment) error {
	return db.CreateInBatches(&payments, 100).Error
}

package repository

import (
	"tokobahankue/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SalePaymentRepository struct {
	Log *logrus.Logger
}

func NewSalePaymentRepository(log *logrus.Logger) *SalePaymentRepository {
	return &SalePaymentRepository{
		Log: log,
	}
}

func (r *SalePaymentRepository) CreateBulk(db *gorm.DB, payments []entity.SalePayment) error {
	return db.CreateInBatches(&payments, 100).Error
}

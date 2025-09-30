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

func (r *PurchasePaymentRepository) DeleteByCode(db *gorm.DB, code string) error {
	return db.Where("purchase_code = ?", code).Delete(&entity.PurchasePayment{}).Error
}

func (r *PurchasePaymentRepository) FindByPurchaseCode(db *gorm.DB, purchaseCode string) ([]entity.PurchasePayment, error) {
	var details = []entity.PurchasePayment{}

	if err := db.Model(&entity.PurchasePayment{}).Where("purchase_code = ?", purchaseCode).Find(&details).Error; err != nil {
		return nil, err
	}

	return details, nil
}

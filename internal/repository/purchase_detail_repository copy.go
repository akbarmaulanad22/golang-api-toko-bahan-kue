package repository

import (
	"tokobahankue/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PurchaseDetailRepository struct {
	Log *logrus.Logger
}

func NewPurchaseDetailRepository(log *logrus.Logger) *PurchaseDetailRepository {
	return &PurchaseDetailRepository{
		Log: log,
	}
}

func (r *PurchaseDetailRepository) CreateBulk(db *gorm.DB, details []entity.PurchaseDetail) error {
	return db.CreateInBatches(&details, 100).Error
}

func (r *PurchaseDetailRepository) Cancel(db *gorm.DB, purchaseCode string, sizeID uint) error {
	return db.Model(&entity.PurchaseDetail{}).
		Where("purchase_code = ? AND size_id = ? AND is_cancelled = 0", purchaseCode, sizeID).
		UpdateColumn("is_cancelled", 1).
		Error
}

func (r *PurchaseDetailRepository) FindPriceBySizeIDAndPurchaseCode(db *gorm.DB, purchaseCode string, sizeID uint, detail *entity.PurchaseDetail) error {
	return db.Model(&entity.PurchaseDetail{}).
		Where("purchase_code = ? AND size_id = ?", purchaseCode, sizeID).
		Take(&detail).
		Error
}

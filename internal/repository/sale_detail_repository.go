package repository

import (
	"tokobahankue/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SaleDetailRepository struct {
	Log *logrus.Logger
}

func NewSaleDetailRepository(log *logrus.Logger) *SaleDetailRepository {
	return &SaleDetailRepository{
		Log: log,
	}
}

func (r *SaleDetailRepository) CreateBulk(db *gorm.DB, details []entity.SaleDetail) error {
	return db.CreateInBatches(&details, 100).Error
}

func (r *SaleDetailRepository) Cancel(db *gorm.DB, saleCode string, sizeID uint) error {
	return db.Model(&entity.SaleDetail{}).
		Where("sale_code = ? AND size_id = ? AND is_cancelled = 0", saleCode, sizeID).
		UpdateColumn("is_cancelled", 1).
		Error
}

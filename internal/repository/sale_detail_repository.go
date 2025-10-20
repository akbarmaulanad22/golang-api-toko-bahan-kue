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

func (r *SaleDetailRepository) CancelByID(db *gorm.DB, id uint) error {
	return db.Model(&entity.SaleDetail{}).
		Where("id = ? AND is_cancelled = 0", id).
		Updates(map[string]interface{}{
			"is_cancelled": 1,
		}).
		Error
}

func (r *SaleDetailRepository) Cancel(db *gorm.DB, saleCode string) error {
	return db.Model(&entity.SaleDetail{}).
		Where("sale_code = ? AND is_cancelled = 0", saleCode).
		Updates(map[string]interface{}{
			"is_cancelled": 1,
		}).
		Error
}

func (r *SaleDetailRepository) FindPriceByID(db *gorm.DB, id uint, detail *entity.SaleDetail) error {
	return db.Model(&entity.SaleDetail{}).
		Where("id = ?", id).
		Take(&detail).
		Error
}

func (r *SaleDetailRepository) FindBySaleCode(db *gorm.DB, saleCode string) ([]entity.SaleDetail, error) {
	var details = []entity.SaleDetail{}

	if err := db.Model(&entity.SaleDetail{}).Where("sale_code = ? AND is_cancelled = 0", saleCode).Find(&details).Error; err != nil {
		return nil, err
	}

	return details, nil

}

func (r *SaleDetailRepository) CountActiveBySaleCode(db *gorm.DB, saleCode string) (int64, error) {
	var count int64
	err := db.Model(&entity.SaleDetail{}).
		Where("sale_code = ? AND is_cancelled = 0", saleCode).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

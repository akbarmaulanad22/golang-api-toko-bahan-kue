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

func (r *SaleDetailRepository) CancelByCodeAndID(db *gorm.DB, code string, id uint) error {
	tx := db.Model(&entity.PurchaseDetail{}).
		Where("sale_code = ? AND id = ? AND is_cancelled = 0", code, id).
		Updates(map[string]interface{}{
			"is_cancelled": 1,
		})

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
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

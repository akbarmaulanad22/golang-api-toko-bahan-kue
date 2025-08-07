package repository

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SaleRepository struct {
	Repository[entity.Sale]
	Log *logrus.Logger
}

func NewSaleRepository(log *logrus.Logger) *SaleRepository {
	return &SaleRepository{
		Log: log,
	}
}

// func (r *SaleRepository) CountBySKU(db *gorm.DB, sku any) (int64, error) {
// 	var total int64
// 	err := db.Model(&entity.Sale{}).Where("sku = ?", sku).Count(&total).Error
// 	return total, err
// }

func (r *SaleRepository) FindByCode(db *gorm.DB, sale *entity.Sale, code string) error {
	return db.Preload("Branch").Where("code = ?", code).First(sale).Error
}

func (r *SaleRepository) Search(db *gorm.DB, request *model.SearchSaleRequest) ([]entity.Sale, int64, error) {
	var sales []entity.Sale
	if err := db.Scopes(r.FilterSale(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&sales).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.Sale{}).Scopes(r.FilterSale(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return sales, total, nil
}

func (r *SaleRepository) FilterSale(request *model.SearchSaleRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if code := request.Code; code != "" {
			tx = tx.Where("code = ?", code)
		}

		if customerName := request.CustomerName; customerName != "" {
			customerName = "%" + customerName + "%"
			tx = tx.Where("customer_name LIKE ?", customerName)
		}

		if status := request.Status; status != "" {
			tx = tx.Where("status = ?", status)
		}

		startAt := request.StartAt
		endAt := request.EndAt
		if startAt != 0 && endAt != 0 {
			tx = tx.Where("paid_at BETWEEN ? AND ? OR created_at BETWEEN ? AND ?", startAt, endAt, startAt, endAt)
		}

		return tx
	}
}

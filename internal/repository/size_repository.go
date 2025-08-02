package repository

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SizeRepository struct {
	Repository[entity.Size]
	Log *logrus.Logger
}

func NewSizeRepository(log *logrus.Logger) *SizeRepository {
	return &SizeRepository{
		Log: log,
	}
}

func (r *SizeRepository) CountBySKU(db *gorm.DB, sku any) (int64, error) {
	var total int64
	err := db.Model(&entity.Size{}).Where("sku = ?", sku).Count(&total).Error
	return total, err
}

func (r *SizeRepository) FindByIdAndProductSKU(db *gorm.DB, size *entity.Size, id uint, productSKU string) error {
	return db.Preload("Product").Where("id = ? AND product_sku = ?", id, productSKU).First(size).Error
}

func (r *SizeRepository) Search(db *gorm.DB, request *model.SearchSizeRequest) ([]entity.Size, int64, error) {
	var sizes []entity.Size
	if err := db.Scopes(r.FilterSize(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&sizes).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.Size{}).Scopes(r.FilterSize(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return sizes, total, nil
}

func (r *SizeRepository) FilterSize(request *model.SearchSizeRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		tx.Where("product_sku = ?", request.ProductSKU)

		if name := request.Name; name != "" {
			name = "%" + name + "%"
			tx = tx.Where("name LIKE ?", name)
		}
		return tx
	}
}

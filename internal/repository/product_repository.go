package repository

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ProductRepository struct {
	Repository[entity.Product]
	Log *logrus.Logger
}

func NewProductRepository(log *logrus.Logger) *ProductRepository {
	return &ProductRepository{
		Log: log,
	}
}

func (r *ProductRepository) CountBySKU(db *gorm.DB, sku any) (int64, error) {
	var total int64
	err := db.Model(&entity.Product{}).Where("sku = ?", sku).Count(&total).Error
	return total, err
}

// func (r *ProductRepository) CountByNameExceptSKU(db *gorm.DB, name any, SKU any) (int64, error) {
// 	var total int64
// 	err := db.Model(&entity.Product{}).Where("name = ? AND sku != ?", name, SKU).Count(&total).Error
// 	return total, err
// }

func (r *ProductRepository) FindBySKU(db *gorm.DB, product *entity.Product, sku string) error {
	return db.Preload("Category").Where("sku = ?", sku).First(product).Error
}

func (r *ProductRepository) Search(db *gorm.DB, request *model.SearchProductRequest) ([]entity.Product, int64, error) {
	var products []entity.Product
	if err := db.Scopes(r.FilterProduct(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Preload("Category").Find(&products).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.Product{}).Scopes(r.FilterProduct(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductRepository) FilterProduct(request *model.SearchProductRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if name := request.Name; name != "" {
			name = "%" + name + "%"
			tx = tx.Where("name LIKE ?", name)
		}

		if sku := request.SKU; sku != "" {
			sku = "%" + sku + "%"
			tx = tx.Where("sku LIKE ?", sku)
		}

		return tx
	}
}

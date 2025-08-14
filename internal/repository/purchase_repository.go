package repository

import (
	"tokobahankue/internal/entity"

	"github.com/sirupsen/logrus"
)

type PurchaseRepository struct {
	Repository[entity.Purchase]
	Log *logrus.Logger
}

func NewPurchaseRepository(log *logrus.Logger) *PurchaseRepository {
	return &PurchaseRepository{
		Log: log,
	}
}

// func (r *PurchaseRepository) FindByCode(db *gorm.DB, sale *entity.Purchase, code string) error {
// 	return db.Preload("Branch").Preload("Distributor").Where("code = ?", code).First(sale).Error
// }

// func (r *PurchaseRepository) Search(db *gorm.DB, request *model.SearchPurchaseRequest) ([]entity.Purchase, int64, error) {
// 	var sales []entity.Purchase
// 	if err := db.Scopes(r.FilterPurchase(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&sales).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	var total int64 = 0
// 	if err := db.Model(&entity.Purchase{}).Scopes(r.FilterPurchase(request)).Count(&total).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	return sales, total, nil
// }

// func (r *PurchaseRepository) FilterPurchase(request *model.SearchPurchaseRequest) func(tx *gorm.DB) *gorm.DB {
// 	return func(tx *gorm.DB) *gorm.DB {
// 		if code := request.Code; code != "" {
// 			tx = tx.Where("code = ?", code)
// 		}

// 		if salesName := request.SalesName; salesName != "" {
// 			salesName = "%" + salesName + "%"
// 			tx = tx.Where("sales_name LIKE ?", salesName)
// 		}

// 		if status := request.Status; status != "" {
// 			tx = tx.Where("status = ?", status)
// 		}

// 		startAt := request.StartAt
// 		endAt := request.EndAt
// 		if startAt != 0 && endAt != 0 {
// 			tx = tx.Where("paid_at BETWEEN ? AND ? OR created_at BETWEEN ? AND ?", startAt, endAt, startAt, endAt)
// 		}

// 		return tx
// 	}
// }

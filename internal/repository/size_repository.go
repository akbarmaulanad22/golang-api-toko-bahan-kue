package repository

import (
	"fmt"
	"strings"
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

func (r *SizeRepository) FindByIdAndProductSKU(db *gorm.DB, size *entity.Size, id uint, productSKU string) error {
	return db.Where("id = ? AND product_sku = ?", id, productSKU).First(size).Error
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

func (r *SizeRepository) FindPriceMapByIDs(tx *gorm.DB, ids []uint) (map[uint]float64, error) {
	var sizes []entity.Size
	result := make(map[uint]float64)
	if err := tx.Select("id", "sell_price").Where("id IN ?", ids).Find(&sizes).Error; err != nil {
		return nil, err
	}
	for _, s := range sizes {
		result[s.ID] = s.SellPrice
	}
	return result, nil
}

func (r *SizeRepository) BulkUpdateBuyPrice(db *gorm.DB, buyPriceBySize map[uint]float64) error {
	if len(buyPriceBySize) == 0 {
		return nil
	}

	var caseStmt strings.Builder
	var ids []uint

	for sizeID, buyPrice := range buyPriceBySize {
		caseStmt.WriteString(fmt.Sprintf(" WHEN %d THEN %f", sizeID, buyPrice))
		ids = append(ids, sizeID)
	}

	sql := fmt.Sprintf(`
		UPDATE sizes
		SET buy_price = CASE id %s END
		WHERE id IN ?
	`, caseStmt.String())

	return db.Exec(sql, ids).Error
}

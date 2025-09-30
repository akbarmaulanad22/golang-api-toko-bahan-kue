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

func (r *PurchaseDetailRepository) CancelBySizeID(db *gorm.DB, purchaseCode string, sizeID uint) error {
	return db.Model(&entity.PurchaseDetail{}).
		Where("purchase_code = ? AND size_id = ? AND is_cancelled = 0", purchaseCode, sizeID).
		UpdateColumn("is_cancelled", 1).
		Error
}

func (r *PurchaseDetailRepository) Cancel(db *gorm.DB, purchaseCode string) error {
	return db.Model(&entity.PurchaseDetail{}).
		Where("purchase_code = ? AND is_cancelled = 0", purchaseCode).
		UpdateColumn("is_cancelled", 1).Error
}

func (r *PurchaseDetailRepository) FindPriceBySizeIDAndPurchaseCode(db *gorm.DB, purchaseCode string, sizeID uint, detail *entity.PurchaseDetail) error {
	return db.Model(&entity.PurchaseDetail{}).
		Where("purchase_code = ? AND size_id = ?", purchaseCode, sizeID).
		Take(&detail).
		Error
}

func (r *PurchaseDetailRepository) FindByPurchaseCode(db *gorm.DB, purchaseCode string) ([]entity.PurchaseDetail, error) {
	var details []entity.PurchaseDetail
	if err := db.Model(&entity.PurchaseDetail{}).
		Where("purchase_code = ? AND is_cancelled = 0", purchaseCode).
		Find(&details).Error; err != nil {
		return nil, err
	}
	return details, nil
}

// func (r *PurchaseDetailRepository) GetLastBuyPricePerSize(db *gorm.DB, sizeIDs []uint, excludePurchaseCode string) (map[uint]float64, error) {
// 	results := make([]struct {
// 		SizeID   uint
// 		BuyPrice float64
// 	}, 0)

// 	// Ambil record terbaru per size_id selain dari purchase yang mau dibatalkan
// 	err := db.Table("purchase_detail pd").
// 		Select("pd.size_id, pd.buy_price").
// 		Joins("JOIN purchase p ON p.code = pd.purchase_code").
// 		Where("pd.size_id IN ? AND pd.purchase_code <> ?", sizeIDs, excludePurchaseCode).
// 		Order("pd.created_at DESC").
// 		Find(&results).Error
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Map hasil ke sizeID -> buyPrice (ambil yang pertama saja karena sudah ORDER DESC)
// 	buyPriceBySize := make(map[uint]float64)
// 	for _, r := range results {
// 		if _, ok := buyPriceBySize[r.SizeID]; !ok {
// 			buyPriceBySize[r.SizeID] = r.BuyPrice
// 		}
// 	}

// 	return buyPriceBySize, nil
// }

func (r *PurchaseDetailRepository) FindLastBuyPricesBySizeIDs(db *gorm.DB, sizeIDs []uint) (map[uint]float64, error) {
	if len(sizeIDs) == 0 {
		return map[uint]float64{}, nil
	}

	rows, err := db.Raw(`
		SELECT pd.size_id, pd.buy_price
		FROM purchase_details pd
		JOIN purchases p ON p.code = pd.purchase_code
		JOIN (
			SELECT pd.size_id, MAX(p.created_at) AS last_created
			FROM purchase_details pd
			JOIN purchases p ON p.code = pd.purchase_code
			WHERE pd.size_id IN ? AND pd.is_cancelled <> 1 AND p.status <> 'CANCELLED'
			GROUP BY pd.size_id
		) latest
		ON latest.size_id = pd.size_id AND latest.last_created = p.created_at
	`, sizeIDs).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uint]float64)
	for rows.Next() {
		var sizeID uint
		var buyPrice float64
		if err := rows.Scan(&sizeID, &buyPrice); err != nil {
			return nil, err
		}
		result[sizeID] = buyPrice
	}

	return result, nil
}

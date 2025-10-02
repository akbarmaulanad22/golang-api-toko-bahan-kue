package repository

import (
	"database/sql"
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
		Updates(map[string]interface{}{
			"is_cancelled": 1,
		}).
		Error
}

func (r *PurchaseDetailRepository) Cancel(db *gorm.DB, purchaseCode string) error {
	return db.Model(&entity.PurchaseDetail{}).
		Where("purchase_code = ? AND is_cancelled = 0", purchaseCode).
		Updates(map[string]interface{}{
			"is_cancelled": 1,
		}).Error
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

func (r *PurchaseDetailRepository) CountActiveByPurchaseCode(db *gorm.DB, purchaseCode string) (int64, error) {
	var count int64
	err := db.Model(&entity.PurchaseDetail{}).
		Where("purchase_code = ? AND is_cancelled = 0", purchaseCode).
		Count(&count).Error
	return count, err
}

func (r *PurchaseDetailRepository) FindLastBuyPricesBySizeIDs(db *gorm.DB, sizeIDs []uint, excludeCode string) (map[uint]float64, error) {
	if len(sizeIDs) == 0 {
		return map[uint]float64{}, nil
	}

	var rows *sql.Rows
	var err error

	if excludeCode != "" {
		// versi dengan exclude purchase tertentu
		rows, err = db.Raw(`
            SELECT pd.size_id, pd.buy_price
            FROM purchase_details pd
            JOIN purchases p ON p.code = pd.purchase_code
            JOIN (
                SELECT pd2.size_id, MAX(p2.created_at) AS last_created
                FROM purchase_details pd2
                JOIN purchases p2 ON p2.code = pd2.purchase_code
                WHERE pd2.size_id IN ? 
                  AND pd2.is_cancelled <> 1 
                  AND p2.status <> 'CANCELLED'
                  AND p2.code <> ?                 -- exclude current purchase
                GROUP BY pd2.size_id
            ) latest
            ON latest.size_id = pd.size_id 
            AND latest.last_created = p.created_at
        `, sizeIDs, excludeCode).Rows()
	} else {
		// versi tanpa exclude
		rows, err = db.Raw(`
            SELECT pd.size_id, pd.buy_price
            FROM purchase_details pd
            JOIN purchases p ON p.code = pd.purchase_code
            JOIN (
                SELECT pd2.size_id, MAX(p2.created_at) AS last_created
                FROM purchase_details pd2
                JOIN purchases p2 ON p2.code = pd2.purchase_code
                WHERE pd2.size_id IN ? 
                  AND pd2.is_cancelled <> 1 
                  AND p2.status <> 'CANCELLED'
                GROUP BY pd2.size_id
            ) latest
            ON latest.size_id = pd.size_id 
            AND latest.last_created = p.created_at
        `, sizeIDs).Rows()
	}

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

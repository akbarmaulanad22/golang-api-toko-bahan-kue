package repository

import (
	"database/sql"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *PurchaseRepository) FindByCode(db *gorm.DB, code string) (*model.PurchaseResponse, error) {
	query := `
		SELECT 
			s.code, s.sales_name, s.status, s.created_at, s.total_price,
			b.name AS branch_name,
			sd.size_id, sd.qty, sd.buy_price AS item_buy_price, sd.is_cancelled,
			sz.name AS size_name, sz.buy_price AS size_buy_price,
			p.sku AS product_sku, p.name AS product_name,
			sp.payment_method, sp.amount, sp.note, sp.created_at AS payment_created_at
		FROM purchases s
		JOIN branches b ON s.branch_id = b.id
		JOIN purchase_details sd ON s.code = sd.purchase_code
		JOIN sizes sz ON sd.size_id = sz.id
		JOIN products p ON sz.product_sku = p.sku
		LEFT JOIN purchase_payments sp ON s.code = sp.purchase_code
		WHERE s.code = ?
	`

	rows, err := db.Raw(query, code).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var purchase *model.PurchaseResponse

	for rows.Next() {
		var (
			purchaseCode, purchasesName, status, branchName string
			createdAt                                       int64
			sizeID, qty, isCancelled                        int
			itemBuyPrice, sizeBuyPrice, totalPrice          float64
			sizeName, productSKU, productName               string

			paymentMethod, note sql.NullString
			amount              sql.NullFloat64
			paymentCreatedAt    sql.NullInt64
		)

		if err := rows.Scan(
			&purchaseCode, &purchasesName, &status, &createdAt, &totalPrice, &branchName,
			&sizeID, &qty, &itemBuyPrice, &isCancelled,
			&sizeName, &sizeBuyPrice,
			&productSKU, &productName,
			&paymentMethod, &amount, &note, &paymentCreatedAt,
		); err != nil {
			return nil, err
		}

		// inisialisasi sekali
		if purchase == nil {
			purchase = &model.PurchaseResponse{
				Code:       purchaseCode,
				SalesName:  purchasesName,
				Status:     status,
				CreatedAt:  createdAt,
				BranchName: branchName,
				TotalQty:   0,
				TotalPrice: totalPrice,
				Items:      []model.PurchaseItemResponse{},
				Payments:   []model.PurchasePaymentResponse{},
			}
		}

		// tambahkan item
		purchase.Items = append(purchase.Items, model.PurchaseItemResponse{
			Size: &model.SizeResponse{
				Name:     sizeName,
				BuyPrice: sizeBuyPrice,
			},
			Product: &model.ProductResponse{
				SKU:  productSKU,
				Name: productName,
			},
			Qty:         qty,
			Price:       itemBuyPrice,
			IsCancelled: isCancelled,
		})

		// akumulasi total qty & price
		purchase.TotalQty += qty
		// purchase.TotalPrice += float64(qty) * itemBuyPrice

		// tambahkan payment kalau ada
		if paymentMethod.Valid && paymentCreatedAt.Valid {
			purchase.Payments = append(purchase.Payments, model.PurchasePaymentResponse{
				PaymentMethod: paymentMethod.String,
				Amount:        amount.Float64,
				Note:          note.String,
				CreatedAt:     paymentCreatedAt.Int64,
			})
		}
	}

	if purchase == nil {
		return nil, gorm.ErrRecordNotFound
	}

	return purchase, nil
}

func (r *PurchaseRepository) Search(db *gorm.DB, request *model.SearchPurchaseRequest) ([]entity.Purchase, int64, error) {
	var sales []entity.Purchase
	if err := db.Order("created_at DESC").Preload("Branch").Scopes(r.FilterPurchase(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&sales).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.Purchase{}).Preload("Branch").Scopes(r.FilterPurchase(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return sales, total, nil
}

func (r *PurchaseRepository) FilterPurchase(request *model.SearchPurchaseRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if request.BranchID != nil {
			tx = tx.Where("branch_id = ?", request.BranchID)
		}

		if search := request.Search; search != "" {
			search = "%" + search + "%"
			tx = tx.Where("code LIKE ? OR sales_name LIKE ?", search, search)
		}

		if status := request.Status; status != "" {
			tx = tx.Where("status = ?", status)
		}

		startAt := request.StartAt
		endAt := request.EndAt
		if startAt != 0 && endAt != 0 {
			tx = tx.Where("created_at BETWEEN ? AND ?", startAt, endAt)
		}

		return tx
	}
}

func (r *PurchaseRepository) Cancel(db *gorm.DB, code string) error {
	return db.Model(&entity.Purchase{}).
		Where("code = ?", code).
		Updates(map[string]interface{}{
			"status":      "CANCELLED",
			"total_price": 0,
		}).Error
}

func (r *PurchaseRepository) UpdateTotalPrice(db *gorm.DB, code string, totalPrice float64) error {
	return db.Model(&entity.Purchase{}).
		Where("code = ?", code).
		UpdateColumn("total_price", totalPrice).
		Error
}

func (r *PurchaseRepository) FindLockByCode(db *gorm.DB, code string, out *entity.Purchase) error {
	return db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("code = ?", code).
		First(out).Error
}

package repository

import (
	"database/sql"
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

func (r *SaleRepository) FindByCode(db *gorm.DB, code string) (*model.SaleResponse, error) {
	query := `
		SELECT 
			s.code, s.customer_name, s.status, s.created_at,
			b.name AS branch_name,
			sd.size_id, sd.qty, sd.sell_price AS item_sell_price,
			sz.name AS size_name, sz.sell_price AS size_sell_price,
			p.sku AS product_sku, p.name AS product_name,
			sp.payment_method, sp.amount, sp.note, sp.created_at AS payment_created_at
		FROM sales s
		JOIN branches b ON s.branch_id = b.id
		JOIN sale_details sd ON s.code = sd.sale_code
		JOIN sizes sz ON sd.size_id = sz.id
		JOIN products p ON sz.product_sku = p.sku
		LEFT JOIN sale_payments sp ON s.code = sp.sale_code
		WHERE s.code = ?
	`

	rows, err := db.Raw(query, code).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sale *model.SaleResponse

	for rows.Next() {
		var (
			saleCode, customerName, status, branchName string
			createdAt                                  int64
			sizeID, qty                                int
			itemSellPrice, sizeSellPrice               float64
			sizeName, productSKU, productName          string

			paymentMethod, note sql.NullString
			amount              sql.NullFloat64
			paymentCreatedAt    sql.NullInt64
		)

		if err := rows.Scan(
			&saleCode, &customerName, &status, &createdAt, &branchName,
			&sizeID, &qty, &itemSellPrice,
			&sizeName, &sizeSellPrice,
			&productSKU, &productName,
			&paymentMethod, &amount, &note, &paymentCreatedAt,
		); err != nil {
			return nil, err
		}

		// inisialisasi sekali
		if sale == nil {
			sale = &model.SaleResponse{
				Code:         saleCode,
				CustomerName: customerName,
				Status:       status,
				CreatedAt:    createdAt,
				BranchName:   branchName,
				TotalQty:     0,
				TotalPrice:   0,
				Items:        []model.SaleItemResponse{},
				Payments:     []model.SalePaymentResponse{},
			}
		}

		// tambahkan item
		sale.Items = append(sale.Items, model.SaleItemResponse{
			Size: &model.SizeResponse{
				Name:      sizeName,
				SellPrice: sizeSellPrice,
			},
			Product: &model.ProductResponse{
				SKU:  productSKU,
				Name: productName,
			},
			Qty:   qty,
			Price: itemSellPrice,
		})

		// akumulasi total qty & price
		sale.TotalQty += qty
		sale.TotalPrice += float64(qty) * itemSellPrice

		// tambahkan payment kalau ada
		if paymentMethod.Valid && paymentCreatedAt.Valid {
			sale.Payments = append(sale.Payments, model.SalePaymentResponse{
				PaymentMethod: paymentMethod.String,
				Amount:        amount.Float64,
				Note:          note.String,
				CreatedAt:     paymentCreatedAt.Int64,
			})
		}
	}

	if sale == nil {
		return nil, gorm.ErrRecordNotFound
	}

	return sale, nil
}

func (r *SaleRepository) Search(db *gorm.DB, request *model.SearchSaleRequest) ([]entity.Sale, int64, error) {
	r.Log.Warnf("DATE BETWEEN %d AND %d", request.StartAt, request.EndAt)

	var sales []entity.Sale
	if err := db.Order("created_at DESC").Preload("Branch").Scopes(r.FilterSale(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&sales).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.Sale{}).Preload("Branch").Scopes(r.FilterSale(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return sales, total, nil
}

func (r *SaleRepository) FilterSale(request *model.SearchSaleRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if request.BranchID != nil {
			tx = tx.Where("branch_id = ?", request.BranchID)
		}

		if search := request.Search; search != "" {
			search = "%" + search + "%"
			tx = tx.Where("code LIKE ? OR customer_name LIKE ?", search, search)
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

func (r *SaleRepository) Cancel(db *gorm.DB, code string) error {
	return db.Model(&entity.Sale{}).
		Where("code = ?", code).
		UpdateColumn("status", "CANCELLED").
		Error
}

package repository

import (
	"fmt"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BranchInventoryRepository struct {
	Repository[entity.BranchInventory]
	Log *logrus.Logger
}

func NewBranchInventoryRepository(log *logrus.Logger) *BranchInventoryRepository {
	return &BranchInventoryRepository{
		Log: log,
	}
}

func (r *BranchInventoryRepository) FindByBranchIDAndSizeID(db *gorm.DB, branchInventory *entity.BranchInventory, branchID, sizeID uint) error {
	return db.Where("branch_id = ? AND size_id = ?", branchID, sizeID).Take(branchInventory).Error
}

func (r *BranchInventoryRepository) UpdateStock(db *gorm.DB, branchInventoryID uint, changeQty int) error {
	return db.Model(&entity.BranchInventory{}).
		Where("id = ?", branchInventoryID).
		UpdateColumn("stock", gorm.Expr("stock + ?", changeQty)).
		Error
}

func (r *BranchInventoryRepository) ListOwnerInventoryByBranch(db *gorm.DB) ([]model.BranchInventoryResponse, error) {
	// Flat query result
	type result struct {
		ID          uint
		BranchID    uint
		BranchName  string
		ProductSKU  uint
		ProductName string
		SizeID      uint
		Size        string
		Stock       int
		CreatedAt   int64
		UpdatedAt   int64
	}

	var rows []result
	query := `
		SELECT 
			bi.id        AS id,
			b.id         AS branch_id,
			b.name       AS branch_name,
			p.sku        AS product_sku,
			p.name       AS product_name,
			s.id         AS size_id,
			s.name       AS size,
			bi.stock,
			bi.created_at,
			bi.updated_at
		FROM branch_inventory bi
		JOIN branches b      ON bi.branch_id = b.id
		JOIN sizes 	  s      ON bi.size_id = s.id
		JOIN products p      ON s.product_sku = p.sku
		ORDER BY b.id, p.sku, s.id;
	`
	if err := db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}

	// Transform ke nested
	branchMap := make(map[uint]*model.BranchInventoryResponse)
	productMap := make(map[string]*model.BranchInventoryProductResponse) // key: request.BranchID_sku

	for _, row := range rows {
		// Branch
		if _, ok := branchMap[row.BranchID]; !ok {
			branchMap[row.BranchID] = &model.BranchInventoryResponse{
				ID:         row.ID,
				BranchID:   row.BranchID,
				BranchName: row.BranchName,
				Products:   []model.BranchInventoryProductResponse{},
				CreatedAt:  row.CreatedAt,
				UpdatedAt:  row.UpdatedAt,
			}
		}

		// Product per branch
		key := fmt.Sprintf("%d_%d", row.BranchID, row.ProductSKU)
		if _, ok := productMap[key]; !ok {
			product := model.BranchInventoryProductResponse{
				ProductSKU:  row.ProductSKU,
				ProductName: row.ProductName,
				Sizes:       []model.BranchInventorySizeResponse{},
			}
			branchMap[row.BranchID].Products = append(branchMap[row.BranchID].Products, product)
			// simpan referensi pointer ke product terakhir
			productMap[key] = &branchMap[row.BranchID].Products[len(branchMap[row.BranchID].Products)-1]
		}

		// Tambah size ke product
		productMap[key].Sizes = append(productMap[key].Sizes, model.BranchInventorySizeResponse{
			SizeID: row.SizeID,
			Size:   row.Size,
			Stock:  row.Stock,
		})
	}

	// Convert map ke slice
	branches := make([]model.BranchInventoryResponse, 0)
	for _, b := range branchMap {
		branches = append(branches, *b)
	}
	return branches, nil
}

func (r *BranchInventoryRepository) ListAdminInventory(db *gorm.DB, request *model.BranchInventoryAdminRequest) (*model.BranchInventoryResponse, error) {
	type result struct {
		ID          uint
		BranchID    uint
		BranchName  string
		ProductSKU  uint
		ProductName string
		SizeID      uint
		Size        string
		Stock       int
		CreatedAt   int64
		UpdatedAt   int64
	}

	var rows []result
	query := `
		SELECT 
			bi.id        AS id,
			b.id         AS branch_id,
			b.name       AS branch_name,
			p.sku        AS product_sku,
			p.name       AS product_name,
			s.id         AS size_id,
			s.name       AS size,
			bi.stock,
			bi.created_at,
			bi.updated_at
		FROM branch_inventory bi
		JOIN branches b ON bi.branch_id = b.id
		JOIN sizes s    ON bi.size_id = s.id
		JOIN products p ON s.product_sku = p.sku
		WHERE bi.branch_id = ?
		ORDER BY p.sku, s.id;
	`
	if err := db.Raw(query, request.BranchID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		// kalau kosong tetap return struct kosong, bukan nil
		return &model.BranchInventoryResponse{
			BranchID: request.BranchID,
			Products: []model.BranchInventoryProductResponse{},
		}, nil
	}

	// Transform ke nested
	resp := &model.BranchInventoryResponse{
		ID:         rows[0].ID,
		BranchID:   rows[0].BranchID,
		BranchName: rows[0].BranchName,
		Products:   []model.BranchInventoryProductResponse{},
		CreatedAt:  rows[0].CreatedAt,
		UpdatedAt:  rows[0].UpdatedAt,
	}

	productMap := make(map[uint]*model.BranchInventoryProductResponse, 0)
	for _, row := range rows {
		if _, ok := productMap[row.ProductSKU]; !ok {
			product := model.BranchInventoryProductResponse{
				ProductSKU:  row.ProductSKU,
				ProductName: row.ProductName,
				Sizes:       []model.BranchInventorySizeResponse{},
			}
			resp.Products = append(resp.Products, product)
			productMap[row.ProductSKU] = &resp.Products[len(resp.Products)-1]
		}

		productMap[row.ProductSKU].Sizes = append(productMap[row.ProductSKU].Sizes, model.BranchInventorySizeResponse{
			SizeID: row.SizeID,
			Size:   row.Size,
			Stock:  row.Stock,
		})
	}

	return resp, nil
}

// func (r *BranchInventoryRepository) CreateOrUp(db *gorm.DB, user *entity.BranchInventory, name string) error {
// 	return db.creat("name = ?", name).First(user).Error
// }

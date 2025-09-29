package repository

import (
	"fmt"
	"strings"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	tx := db.Model(&entity.BranchInventory{}).
		Where("id = ? AND stock + ? >= 0", branchInventoryID, changeQty).
		UpdateColumn("stock", gorm.Expr("stock + ?", changeQty))
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("stok tidak cukup / record tidak ditemukan")
	}
	return nil
}

// func (r *BranchInventoryRepository) BulkDecreaseStock(db *gorm.DB, branchID uint, details []entity.SaleDetail) error {
// 	if len(details) == 0 {
// 		return nil
// 	}

// 	// Build CASE WHEN
// 	caseStmt := "CASE size_id"
// 	sizeIDs := make([]string, len(details))
// 	for i, d := range details {
// 		caseStmt += fmt.Sprintf(" WHEN %d THEN stock - %d", d.SizeID, d.Qty)
// 		sizeIDs[i] = fmt.Sprintf("%d", d.SizeID)
// 	}
// 	caseStmt += " END"

// 	validateCase := "CASE size_id"
// 	for _, d := range details {
// 		validateCase += fmt.Sprintf(" WHEN %d THEN %d", d.SizeID, d.Qty)
// 	}
// 	validateCase += " END"

// 	query := fmt.Sprintf(`
//         UPDATE branch_inventory
//         SET stock = %s
//         WHERE branch_id = ?
//           AND size_id IN (%s)
//           AND stock >= %s
//     `, caseStmt, strings.Join(sizeIDs, ","), validateCase)

// 	tx := db.Exec(query, branchID)
// 	if tx.Error != nil {
// 		return tx.Error
// 	}

// 	if tx.RowsAffected != int64(len(details)) {
// 		return fmt.Errorf("stok tidak cukup / ada record tidak ditemukan")
// 	}
// 	return nil
// }

func (r *BranchInventoryRepository) BulkDecreaseStock(db *gorm.DB, branchID uint, qtyBySize map[uint]int) error {
	if len(qtyBySize) == 0 {
		return nil
	}

	caseStmt := "CASE size_id"
	sizeIDs := make([]string, 0, len(qtyBySize))
	for sizeID, qty := range qtyBySize {
		caseStmt += fmt.Sprintf(" WHEN %d THEN stock - %d", sizeID, qty)
		sizeIDs = append(sizeIDs, fmt.Sprintf("%d", sizeID))
	}
	caseStmt += " END"

	validateCase := "CASE size_id"
	for sizeID, qty := range qtyBySize {
		validateCase += fmt.Sprintf(" WHEN %d THEN %d", sizeID, qty)
	}
	validateCase += " END"

	query := fmt.Sprintf(`
        UPDATE branch_inventory
        SET stock = %s
        WHERE branch_id = ? 
          AND size_id IN (%s)
          AND stock >= %s
    `, caseStmt, strings.Join(sizeIDs, ","), validateCase)

	tx := db.Exec(query, branchID)
	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected != int64(len(qtyBySize)) {
		return fmt.Errorf("stok tidak cukup / ada record tidak ditemukan")
	}
	return nil
}

func (r *BranchInventoryRepository) BulkIncreaseStock(db *gorm.DB, inventories []entity.BranchInventory, qtyBySize map[uint]int) error {
	var caseStmt strings.Builder
	var ids []uint

	for _, inv := range inventories {
		addQty := qtyBySize[inv.SizeID]
		if addQty > 0 {
			caseStmt.WriteString(fmt.Sprintf(" WHEN %d THEN %d", inv.ID, addQty))
			ids = append(ids, inv.ID)
		}
	}

	if len(ids) == 0 {
		return nil
	}

	sql := fmt.Sprintf(`
		UPDATE branch_inventory 
		SET stock = stock + CASE id %s END 
		WHERE id IN ?
	`, caseStmt.String())

	return db.Exec(sql, ids).Error
}

// func (r *BranchInventoryRepository) BulkIncreaseStock(db *gorm.DB, branchID uint, details []entity.SaleDetail) error {
// 	if len(details) == 0 {
// 		return nil
// 	}

// 	// Build CASE WHEN
// 	caseStmt := "CASE size_id"
// 	sizeIDs := make([]string, len(details))
// 	for i, d := range details {
// 		caseStmt += fmt.Sprintf(" WHEN %d THEN stock + %d", d.SizeID, d.Qty)
// 		sizeIDs[i] = fmt.Sprintf("%d", d.SizeID)
// 	}
// 	caseStmt += " END"

// 	validateCase := "CASE size_id"
// 	for _, d := range details {
// 		validateCase += fmt.Sprintf(" WHEN %d THEN %d", d.SizeID, d.Qty)
// 	}
// 	validateCase += " END"

// 	query := fmt.Sprintf(`
//         UPDATE branch_inventory
//         SET stock = %s
//         WHERE branch_id = ?
//           AND size_id IN (%s)
//           AND stock >= %s
//     `, caseStmt, strings.Join(sizeIDs, ","), validateCase)

// 	tx := db.Exec(query, branchID)
// 	if tx.Error != nil {
// 		return tx.Error
// 	}
// 	if tx.RowsAffected != int64(len(details)) {
// 		return fmt.Errorf("stok tidak cukup / ada record tidak ditemukan")
// 	}
// 	return nil
// }

func (r *BranchInventoryRepository) FindByBranchAndSizeIDs(db *gorm.DB, branchID uint, sizeIDs []uint) ([]entity.BranchInventory, error) {
	var inventories []entity.BranchInventory
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("branch_id = ? AND size_id IN ?", branchID, sizeIDs).
		Find(&inventories).Error; err != nil {
		return nil, err
	}

	return inventories, nil
}

// func (r *BranchInventoryRepository) ListOwnerInventoryByBranch(db *gorm.DB) ([]model.BranchInventoryProductResponse, error) {
// 	// Flat query result
// 	type result struct {
// 		ID          uint
// 		BranchID    uint
// 		BranchName  string
// 		ProductSKU  string
// 		ProductName string
// 		SizeID      uint
// 		Size        string
// 		Stock       int
// 		CreatedAt   int64
// 		UpdatedAt   int64
// 	}

// 	var rows []result
// 	query := `
// 		SELECT
// 			bi.id        AS id,
// 			b.id         AS branch_id,
// 			b.name       AS branch_name,
// 			p.sku        AS product_sku,
// 			p.name       AS product_name,
// 			s.id         AS size_id,
// 			s.name       AS size,
// 			bi.stock,
// 			bi.created_at,
// 			bi.updated_at
// 		FROM branch_inventory bi
// 		JOIN branches b      ON bi.branch_id = b.id
// 		JOIN sizes 	  s      ON bi.size_id = s.id
// 		JOIN products p      ON s.product_sku = p.sku
// 		ORDER BY b.id, p.sku, s.id;
// 	`
// 	if err := db.Raw(query).Scan(&rows).Error; err != nil {
// 		return nil, err
// 	}

// 	// Transform ke nested
// 	branchMap := make(map[uint]*model.BranchInventoryResponse)
// 	productMap := make(map[string]*model.BranchInventoryProductResponse) // key: request.BranchID_sku

// 	for _, row := range rows {
// 		// Branch
// 		if _, ok := branchMap[row.BranchID]; !ok {
// 			branchMap[row.BranchID] = &model.BranchInventoryResponse{
// 				ID:         row.ID,
// 				BranchID:   row.BranchID,
// 				BranchName: row.BranchName,
// 				Products:   []model.BranchInventoryProductResponse{},
// 				CreatedAt:  row.CreatedAt,
// 				UpdatedAt:  row.UpdatedAt,
// 			}
// 		}

// 		// Product per branch
// 		key := fmt.Sprintf("%d_%d", row.BranchID, row.ProductSKU)
// 		if _, ok := productMap[key]; !ok {
// 			product := model.BranchInventoryProductResponse{
// 				ProductSKU:  row.ProductSKU,
// 				ProductName: row.ProductName,
// 				Sizes:       []model.BranchInventorySizeResponse{},
// 			}
// 			branchMap[row.BranchID].Products = append(branchMap[row.BranchID].Products, product)
// 			// simpan referensi pointer ke product terakhir
// 			productMap[key] = &branchMap[row.BranchID].Products[len(branchMap[row.BranchID].Products)-1]
// 		}

// 		// Tambah size ke product
// 		productMap[key].Sizes = append(productMap[key].Sizes, model.BranchInventorySizeResponse{
// 			SizeID: row.SizeID,
// 			Size:   row.Size,
// 			Stock:  row.Stock,
// 		})
// 	}

// 	// Convert map ke slice
// 	branches := make([]model.BranchInventoryResponse, 0)
// 	for _, b := range branchMap {
// 		branches = append(branches, *b)
// 	}
// 	return branches, nil
// }

func (r *BranchInventoryRepository) Search(db *gorm.DB, req *model.SearchBranchInventoryRequest) ([]model.BranchInventoryProductResponse, int64, error) {
	type row struct {
		BranchName  string
		ProductSKU  string
		ProductName string
		SizeID      uint
		SizeName    string
		Stock       int
		SellPrice   float64
	}

	var rows []row
	var total int64

	// base query
	baseQuery := `
		FROM branch_inventory bi
		JOIN sizes s ON s.id = bi.size_id
		JOIN products p ON p.sku = s.product_sku
		JOIN branches b ON b.id = bi.branch_id
		WHERE 1=1
	`

	args := []interface{}{}
	if req.Search != "" {
		search := "%" + req.Search + "%"
		baseQuery += ` AND (b.name LIKE ? OR p.sku LIKE ? OR p.name LIKE ?)`
		args = append(args, search, search, search)
	}

	if req.BranchID != nil {
		baseQuery += ` AND b.id = ?`
		args = append(args, req.BranchID)
	}

	// hitung total produk unik per branch
	countQuery := `SELECT COUNT(DISTINCT CONCAT(b.id,'-',p.sku)) ` + baseQuery
	if err := db.Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// ambil product unik dengan paging
	type productKey struct {
		BranchID    uint
		BranchName  string
		ProductSKU  string
		ProductName string
	}
	productKeys := []productKey{}

	productQuery := `
		SELECT DISTINCT b.id AS branch_id, b.name AS branch_name, p.sku AS product_sku, p.name AS product_name
	` + baseQuery + `
		ORDER BY b.id, p.sku
		LIMIT ? OFFSET ?`
	argsWithLimit := append(args, req.Size, (req.Page-1)*req.Size)

	if err := db.Raw(productQuery, argsWithLimit...).Scan(&productKeys).Error; err != nil {
		return nil, 0, err
	}

	if len(productKeys) == 0 {
		return []model.BranchInventoryProductResponse{}, total, nil
	}

	// ambil semua sizes untuk product yang sudah dipilih
	sizeArgs := []interface{}{}
	sizeConditions := []string{}
	for _, p := range productKeys {
		sizeConditions = append(sizeConditions, "(b.id = ? AND p.sku = ?)")
		sizeArgs = append(sizeArgs, p.BranchID, p.ProductSKU)
	}

	sizeQuery := `
		SELECT 
			b.name AS branch_name,
			p.sku AS product_sku,
			p.name AS product_name,
			s.id AS size_id,
			s.name AS size_name,
			s.sell_price,
			bi.stock
		FROM branch_inventory bi
		JOIN sizes s ON s.id = bi.size_id
		JOIN products p ON p.sku = s.product_sku
		JOIN branches b ON b.id = bi.branch_id
		WHERE ` + strings.Join(sizeConditions, " OR ") + `
		ORDER BY b.id, p.sku, s.id`

	if err := db.Raw(sizeQuery, sizeArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	// mapping ke response per product
	productMap := make(map[string]*model.BranchInventoryProductResponse)
	for _, row := range rows {
		key := fmt.Sprintf("%s-%s", row.BranchName, row.ProductSKU)
		if _, ok := productMap[key]; !ok {
			productMap[key] = &model.BranchInventoryProductResponse{
				BranchName:  row.BranchName,
				ProductSKU:  row.ProductSKU,
				ProductName: row.ProductName,
				Sizes:       []model.BranchInventorySizeResponse{},
			}
		}
		productMap[key].Sizes = append(productMap[key].Sizes, model.BranchInventorySizeResponse{
			SizeID:    row.SizeID,
			Size:      row.SizeName,
			Stock:     row.Stock,
			SellPrice: row.SellPrice,
		})
	}

	// final array
	results := make([]model.BranchInventoryProductResponse, 0, len(productMap))
	for _, p := range productMap {
		results = append(results, *p)
	}

	return results, total, nil
}

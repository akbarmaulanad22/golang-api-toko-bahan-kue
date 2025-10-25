package usecase

import (
	"context"
	"fmt"
	"time"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/model/converter"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StockOpnameUseCase struct {
	DB                    *gorm.DB
	Log                   *logrus.Logger
	Validate              *validator.Validate
	StockOpnameRepository *repository.StockOpnameRepository
	BranchInventoryRepo   *repository.BranchInventoryRepository
	InventoryMovementRepo *repository.InventoryMovementRepository
	CashBankRepo          *repository.CashBankTransactionRepository
}

func NewStockOpnameUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	stockOpnameRepo *repository.StockOpnameRepository,
	branchInventoryRepo *repository.BranchInventoryRepository,
	inventoryMovementRepo *repository.InventoryMovementRepository,
	cashBankRepo *repository.CashBankTransactionRepository,
) *StockOpnameUseCase {
	return &StockOpnameUseCase{
		DB:                    db,
		Log:                   log,
		Validate:              validate,
		StockOpnameRepository: stockOpnameRepo,
		BranchInventoryRepo:   branchInventoryRepo,
		InventoryMovementRepo: inventoryMovementRepo,
		CashBankRepo:          cashBankRepo,
	}
}

func (c *StockOpnameUseCase) Create(ctx context.Context, req *model.CreateStockOpnameRequest) (*model.StockOpnameResponse, error) {
	if err := c.Validate.Struct(req); err != nil {
		c.Log.WithError(err).Error("error validating request")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	opname := &entity.StockOpname{
		BranchID:  req.BranchID,
		CreatedBy: req.CreatedBy,
		Status:    "draft",
	}

	// ambil semua branch_inventory_id dari request
	invIDs := make([]uint, 0, len(req.Details))
	for _, d := range req.Details {
		invIDs = append(invIDs, d.BranchInventoryID)
	}

	// ambil semua data inventory berdasarkan ID tersebut
	inventories, err := c.BranchInventoryRepo.FindByIDs(tx, invIDs)
	if err != nil {
		c.Log.WithError(err).Error("error finding branch inventories")
		return nil, model.NewAppErr("internal server error", nil)
	}
	if len(inventories) != len(invIDs) {
		c.Log.Warnf("some branch inventories not found, expected=%d got=%d", len(invIDs), len(inventories))
		return nil, model.NewAppErr("not found", "some inventories not found for provided branch_inventory_ids")
	}

	// bikin map biar lookup cepat
	invMap := make(map[uint]entity.BranchInventory, len(inventories))
	for _, inv := range inventories {
		invMap[inv.ID] = inv
	}

	// bangun detail opname
	for _, input := range req.Details {
		inv, ok := invMap[input.BranchInventoryID]
		if !ok {
			c.Log.Warnf("branch inventory not found for id=%d", input.BranchInventoryID)
			return nil, model.NewAppErr("inventory not found", "some branch inventory not found")
		}

		opname.Details = append(opname.Details, entity.StockOpnameDetail{
			BranchInventoryID: inv.ID,
			SystemQty:         int64(inv.Stock),
			PhysicalQty:       input.PhysicalQty,
			Notes:             input.Notes,
			Difference:        input.PhysicalQty - int64(inv.Stock),
		})
	}

	// simpan opname & detail
	if err := c.StockOpnameRepository.Create(tx, opname); err != nil {
		c.Log.WithError(err).Error("error creating stock opname draft with details")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error committing stock opname")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.StockOpnameToResponse(opname), nil
}

func (c *StockOpnameUseCase) Update(ctx context.Context, req *model.UpdateStockOpnameRequest) (*model.StockOpnameResponse, error) {
	if err := c.Validate.Struct(req); err != nil {
		c.Log.WithError(err).Error("error validating request")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	opname := new(entity.StockOpname)
	if err := c.StockOpnameRepository.FindByIdWithDetails(tx, opname, req.ID); err != nil {
		c.Log.WithError(err).Error("error getting stock opname")
		return nil, helper.GetNotFoundMessage("stock opname", err)
	}

	if opname.Status != "draft" {
		c.Log.WithField("stock_opname_id", opname.ID).Error("only draft opname can be approved")
		return nil, model.NewAppErr("conflict", "only draft stock opname can be approved")
	}

	// 1. Sinkronkan physical_qty dan notes dari request ke detail opname
	detailMap := make(map[uint]model.UpdateStockOpnameDetailRequest)
	for _, d := range req.Details {
		detailMap[d.ID] = d
	}

	for i := range opname.Details {
		if dReq, ok := detailMap[opname.Details[i].ID]; ok {
			opname.Details[i].PhysicalQty = dReq.PhysicalQty
			opname.Details[i].Notes = dReq.Notes
		}
	}

	// 2. Build unique list of sizeIDs (hindari duplikat)
	sizeIDSet := make(map[uint]struct{})
	for _, d := range opname.Details {
		sizeIDSet[d.BranchInventory.SizeID] = struct{}{}
	}
	sizeIDs := make([]uint, 0, len(sizeIDSet))
	for sid := range sizeIDSet {
		sizeIDs = append(sizeIDs, sid)
	}

	// 3. Query ulang inventory sistem terkini
	inventories, err := c.BranchInventoryRepo.FindByBranchAndSizeIDs(tx, opname.BranchID, sizeIDs)
	if err != nil {
		c.Log.WithError(err).Error("error getting inventories")
		return nil, helper.GetNotFoundMessage("inventories", err)
	}

	// map sizeID -> inventory
	invMap := make(map[uint]entity.BranchInventory, len(inventories))
	for _, inv := range inventories {
		invMap[inv.SizeID] = inv
	}

	// Pastikan tiap detail memiliki inventory
	for _, d := range opname.Details {
		if _, ok := invMap[d.BranchInventory.SizeID]; !ok {
			c.Log.WithField("size_id", d.BranchInventory.SizeID).Error("inventory not found for detail size_id")
			return nil, model.NewAppErr("inventories not found", "some inventories not found for provided size_ids")
		}
	}

	// 4. Hitung perbedaan dan langsung update stok per inventory di DB,
	//    serta persiapkan inventory movement di memori
	var (
		totalLoss, totalGain float64
		movements            []entity.InventoryMovement
	)

	for i := range opname.Details {
		detail := &opname.Details[i]
		inv := invMap[detail.BranchInventory.SizeID]

		// pastikan systemQty dari DB
		systemQty := int(inv.Stock)
		diff := int(detail.PhysicalQty) - systemQty
		buyPrice := float64(inv.Size.BuyPrice)

		detail.SystemQty = int64(systemQty)
		detail.Difference = int64(diff)

		if diff == 0 {
			// nothing to do for this detail
			continue
		}

		// 4.a Update stok inventory di DB secara atomik (stock = stock + diff)
		if err := tx.Model(&entity.BranchInventory{}).
			Where("id = ?", inv.ID).
			UpdateColumn("stock", gorm.Expr("stock + ?", diff)).Error; err != nil {
			c.Log.WithError(err).WithField("branch_inventory_id", inv.ID).Error("error updating branch inventory stock")
			return nil, model.NewAppErr("internal server error", nil)
		}

		// 4.b Update inv.Stock di memori agar response konsisten (ambil nilai terbaru)
		inv.Stock = inv.Stock + diff
		// update map supaya jika ada referensi lagi, kita punya nilai yang up-to-date
		invMap[detail.BranchInventory.SizeID] = inv
		// juga update detail's BranchInventory.Stock (jika struct detail embed BranchInventory)
		detail.BranchInventory.Stock = inv.Stock

		// 4.c prepare movement (ChangeQty boleh positif/negatif)
		movements = append(movements, entity.InventoryMovement{
			BranchInventoryID: inv.ID,
			ChangeQty:         diff,
			ReferenceType:     "STOCK_OPNAME",
			ReferenceKey:      fmt.Sprintf("%d", opname.ID),
		})

		// 4.d hitung nilai loss/gain berdasarkan buyPrice
		value := float64(diff) * buyPrice
		if diff < 0 {
			// diff negatif => stock berkurang => loss
			totalLoss += -value
		} else {
			// diff positif => stock bertambah => gain (modal)
			totalGain += value
		}
	}

	// 5. Insert inventory movement log (jika ada)
	if len(movements) > 0 {
		if err := tx.Create(&movements).Error; err != nil {
			c.Log.WithError(err).Error("error inserting inventory movements")
			return nil, model.NewAppErr("internal server error", nil)
		}
	}

	// 6. Insert cashbank transactions
	if totalLoss > 0 {
		lossTx := &entity.CashBankTransaction{
			BranchID:     &opname.BranchID,
			Type:         "OUT",
			Source:       "STOCK_OPNAME",
			ReferenceKey: fmt.Sprintf("%d", opname.ID),
			Amount:       totalLoss,
			Description:  fmt.Sprintf("Stock loss adjustment from opname #%d", opname.ID),
		}
		if err := c.CashBankRepo.Create(tx, lossTx); err != nil {
			c.Log.WithError(err).Error("error creating loss transaction")
			return nil, model.NewAppErr("internal server error", nil)
		}
	}

	if totalGain > 0 {
		gainTx := &entity.CashBankTransaction{
			BranchID:     &opname.BranchID,
			Type:         "IN",
			Source:       "STOCK_OPNAME",
			ReferenceKey: fmt.Sprintf("%d", opname.ID),
			Amount:       totalGain,
			Description:  fmt.Sprintf("Stock gain adjustment from opname #%d", opname.ID),
		}
		if err := c.CashBankRepo.Create(tx, gainTx); err != nil {
			c.Log.WithError(err).Error("error creating gain transaction")
			return nil, model.NewAppErr("internal server error", nil)
		}
	}

	// 7. Update opname status + simpan semua relasi (FullSaveAssociations)
	opname.Status = "completed"
	opname.VerifiedBy = req.VerifiedBy
	opname.CompletedAt = time.Now().UnixMilli()

	if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&opname).Error; err != nil {
		c.Log.WithError(err).Error("error updating opname status to completed")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error committing stock opname approval")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.StockOpnameToResponse(opname), nil
}

func (c *StockOpnameUseCase) Delete(ctx context.Context, request *model.DeleteStockOpnameRequest) error {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request")
		return helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	opname := new(entity.StockOpname)
	if err := c.StockOpnameRepository.FindByIdWithDetails(tx, opname, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting stock opname")
		return helper.GetNotFoundMessage("stock opname", err)
	}

	if opname.Status != "draft" {
		c.Log.WithField("stock_opname_id", opname.ID).
			Error("only draft stock opname can be deleted")
		return model.NewAppErr("conflict", "only draft stock opname can be deleted")
	}

	if err := c.StockOpnameRepository.Delete(tx, opname); err != nil {
		c.Log.WithError(err).Error("error deleting stock opname")
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error committing delete stock opname")
		return model.NewAppErr("internal server error", nil)
	}

	c.Log.WithField("stock_opname_id", opname.ID).
		Info("draft stock opname deleted successfully")
	return nil
}

func (c *StockOpnameUseCase) Get(ctx context.Context, request *model.GetStockOpnameRequest) (*model.StockOpnameResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	stockOpname := new(entity.StockOpname)
	if err := c.StockOpnameRepository.FindByIdWithDetails(tx, stockOpname, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting stockOpname")
		return nil, helper.GetNotFoundMessage("stock opname", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting stock opname")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.StockOpnameToResponse(stockOpname), nil
}

func (c *StockOpnameUseCase) Search(ctx context.Context, request *model.SearchStockOpnameRequest) ([]model.StockOpnameResponse, int64, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	stockOpname, total, err := c.StockOpnameRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting stock opname")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting stock opname")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	responses := make([]model.StockOpnameResponse, len(stockOpname))
	for i, role := range stockOpname {
		responses[i] = *converter.StockOpnameToResponse(&role)
	}

	return responses, total, nil
}

package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
	"tokobahankue/internal/model/converter"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PurchaseUseCase struct {
	DB                            *gorm.DB
	Log                           *logrus.Logger
	Validate                      *validator.Validate
	PurchaseRepository            *repository.PurchaseRepository
	PurchaseDetailRepository      *repository.PurchaseDetailRepository
	PurchasePaymentRepository     *repository.PurchasePaymentRepository
	DebtRepository                *repository.DebtRepository
	DebtPaymentRepository         *repository.DebtPaymentRepository
	SizeRepository                *repository.SizeRepository
	CashBankTransactionRepository *repository.CashBankTransactionRepository
	BranchInventoryRepository     *repository.BranchInventoryRepository
	InventoryMovementRepository   *repository.InventoryMovementRepository
}

func NewPurchaseUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	purchaseRepository *repository.PurchaseRepository,
	purchaseDetailRepository *repository.PurchaseDetailRepository,
	purchasePaymentRepository *repository.PurchasePaymentRepository,
	debtRepository *repository.DebtRepository,
	debtPaymentRepository *repository.DebtPaymentRepository,
	sizeRepository *repository.SizeRepository,
	cashBankTransactionRepository *repository.CashBankTransactionRepository,
	branchInventoryRepository *repository.BranchInventoryRepository,
	inventoryMovementRepository *repository.InventoryMovementRepository,

) *PurchaseUseCase {
	return &PurchaseUseCase{
		DB:                            db,
		Log:                           logger,
		Validate:                      validate,
		PurchaseRepository:            purchaseRepository,
		DebtRepository:                debtRepository,
		SizeRepository:                sizeRepository,
		PurchaseDetailRepository:      purchaseDetailRepository,
		PurchasePaymentRepository:     purchasePaymentRepository,
		DebtPaymentRepository:         debtPaymentRepository,
		CashBankTransactionRepository: cashBankTransactionRepository,
		BranchInventoryRepository:     branchInventoryRepository,
		InventoryMovementRepository:   inventoryMovementRepository,
	}
}

func (c *PurchaseUseCase) Create(ctx context.Context, request *model.CreatePurchaseRequest) (*model.PurchaseResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		return nil, errors.New("bad request")
	}

	if request.Debt == nil && len(request.Payments) == 0 || (request.Debt != nil && len(request.Payments) > 0) {
		return nil, errors.New("bad request: either debt or payments must be provided")
	}

	qtyBySize := map[uint]int{}
	for _, d := range request.Details {
		qtyBySize[d.SizeID] += d.Qty
	}
	sizeIDs := make([]uint, 0, len(qtyBySize))
	for sid := range qtyBySize {
		sizeIDs = append(sizeIDs, sid)
	}
	slices.Sort(sizeIDs)

	// Ambil inventory
	branchInvs, err := c.BranchInventoryRepository.FindByBranchAndSizeIDs(tx, request.BranchID, sizeIDs)
	if err != nil {
		return nil, errors.New("internal server error")
	}

	branchInvMap := make(map[uint]*entity.BranchInventory, len(branchInvs))
	for i := range branchInvs {
		branchInvMap[branchInvs[i].SizeID] = &branchInvs[i]
	}

	// Buat kode sale
	purchaseCode := "PURCHASE-" + time.Now().Format("20060102150405")

	// Build sale details & movements
	n := len(request.Details)
	details := make([]entity.PurchaseDetail, n)
	movements := make([]entity.InventoryMovement, n)
	var totalPrice float64
	now := time.Now().UnixMilli()
	buyPriceBySize := map[uint]float64{}

	for i, d := range request.Details {
		inv, exists := branchInvMap[d.SizeID]
		if !exists {
			return nil, fmt.Errorf("stok untuk size_id %d tidak ditemukan di branch %d", d.SizeID, request.BranchID)
		}

		totalPrice += d.BuyPrice * float64(d.Qty)

		details[i] = entity.PurchaseDetail{
			PurchaseCode: purchaseCode,
			SizeID:       d.SizeID,
			Qty:          d.Qty,
			BuyPrice:     d.BuyPrice,
		}
		movements[i] = entity.InventoryMovement{
			BranchInventoryID: inv.ID,
			ChangeQty:         d.Qty,
			ReferenceType:     "PURCHASE",
			ReferenceKey:      purchaseCode,
			CreatedAt:         now,
		}

		// Simpan harga beli terbaru
		buyPriceBySize[d.SizeID] = d.BuyPrice
	}

	// Bulk update stok (1 query)
	if err := c.BranchInventoryRepository.BulkIncreaseStock(tx, branchInvs, qtyBySize); err != nil {
		return nil, err
	}

	// Insert sale
	sale := entity.Purchase{
		Code:          purchaseCode,
		BranchID:      request.BranchID,
		DistributorID: request.DistributorID,
		SalesName:     request.SalesName,
		TotalPrice:    totalPrice,
	}

	if err := c.PurchaseRepository.Create(tx, &sale); err != nil {
		return nil, err
	}

	if err := c.PurchaseDetailRepository.CreateBulk(tx, details); err != nil {
		return nil, err
	}

	if err := c.InventoryMovementRepository.CreateBulk(tx, movements); err != nil {
		return nil, err
	}

	if err := c.SizeRepository.BulkUpdateBuyPrice(tx, buyPriceBySize); err != nil {
		return nil, err
	}

	// Insert payments (jika ada)
	if n := len(request.Payments); n > 0 {
		payments := make([]entity.PurchasePayment, n)
		var totalPayment float64
		for i, p := range request.Payments {
			totalPayment += p.Amount
			payments[i] = entity.PurchasePayment{
				PurchaseCode:  purchaseCode,
				PaymentMethod: p.PaymentMethod,
				Amount:        p.Amount,
				Note:          p.Note,
			}
		}

		if totalPayment < totalPrice {
			return nil, errors.New("bad request: total payment is less than total price")
		}

		if err := c.PurchasePaymentRepository.CreateBulk(tx, payments); err != nil {
			return nil, err
		}

		cashBankTransaction := entity.CashBankTransaction{
			TransactionDate: now,
			Type:            "OUT",
			Source:          "PURCHASE",
			Amount:          totalPrice,
			ReferenceKey:    purchaseCode,
			BranchID:        &request.BranchID,
		}
		if err := c.CashBankTransactionRepository.Create(tx, &cashBankTransaction); err != nil {
			return nil, err
		}
	}

	// Insert debt (jika ada)
	if request.Debt != nil {
		var paidAmount float64
		for _, p := range request.Debt.DebtPayments {
			paidAmount += p.Amount
		}

		if paidAmount > totalPrice {
			return nil, errors.New("bad request: total debt payment is more than total price")
		}

		debt := entity.Debt{
			ReferenceType: "PURCHASE",
			ReferenceCode: purchaseCode,
			TotalAmount:   totalPrice,
			PaidAmount:    paidAmount,
			DueDate: func() int64 {
				if request.Debt.DueDate > 0 {
					return int64(request.Debt.DueDate)
				}
				return time.Now().Add(7 * 24 * time.Hour).UnixMilli()
			}(),
			Status: "PENDING",
		}
		if err := c.DebtRepository.Create(tx, &debt); err != nil {
			return nil, err
		}

		if n := len(request.Debt.DebtPayments); n > 0 {
			debtPayments := make([]entity.DebtPayment, n)

			for i, p := range request.Debt.DebtPayments {
				debtPayments[i] = entity.DebtPayment{
					DebtID:      debt.ID,
					PaymentDate: now,
					Amount:      p.Amount,
					Note:        p.Note,
				}
			}
			if err := c.DebtPaymentRepository.CreateBulk(tx, debtPayments); err != nil {
				return nil, err
			}
			cashBankTransaction := entity.CashBankTransaction{
				TransactionDate: now,
				Type:            "OUT",
				Source:          "DEBT",
				Amount:          paidAmount,
				Description:     "Bayar Hutang Cicilan Pertama",
				ReferenceKey:    strconv.Itoa(int(debt.ID)),
				BranchID:        &request.BranchID,
			}
			if err := c.CashBankTransactionRepository.Create(tx, &cashBankTransaction); err != nil {
				return nil, err
			}
		}
	}

	// Commit transaksi
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("internal server error")
	}
	return converter.PurchaseToResponse(&sale), nil
}

func (c *PurchaseUseCase) Search(ctx context.Context, request *model.SearchPurchaseRequest) ([]model.PurchaseResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	purchases, total, err := c.PurchaseRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting purchases")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting purchases")
		return nil, 0, errors.New("internal server error")
	}

	responses := make([]model.PurchaseResponse, len(purchases))
	for i, purchase := range purchases {
		responses[i] = *converter.PurchaseToResponse(&purchase)
	}

	return responses, total, nil
}

func (c *PurchaseUseCase) Get(ctx context.Context, request *model.GetPurchaseRequest) (*model.PurchaseResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	purchase, err := c.PurchaseRepository.FindByCode(tx, request.Code)
	if err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return nil, errors.New("internal server error")
	}

	return purchase, nil
}

// func (c *PurchaseUseCase) Cancel(ctx context.Context, request *model.CancelPurchaseRequest) (*model.PurchaseResponse, error) {
// 	tx := c.DB.WithContext(ctx).Begin()
// 	defer tx.Rollback()

// 	if err := c.Validate.Struct(request); err != nil {
// 		c.Log.WithError(err).Error("error validating request body")
// 		return nil, errors.New("bad request")
// 	}

// 	purchase, err := c.PurchaseRepository.FindByCode(tx, request.Code)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error getting purchase")
// 		return nil, errors.New("not found")
// 	}

// 	createdTime := time.UnixMilli(purchase.CreatedAt)
// 	now := time.Now()

// 	// Hitung durasi sejak dibuat
// 	duration := now.Sub(createdTime)

// 	// Jika status BUKAN PENDING dan sudah lewat 24 jam => tolak
// 	if duration.Hours() >= 24 {
// 		c.Log.WithField("purchase_code", purchase.Code).Error("error updating purchase: exceeded 24-hour window")
// 		return nil, errors.New("forbidden")
// 	}

// 	// Lanjut update status
// 	if err := c.PurchaseRepository.Cancel(tx, purchase.Code); err != nil {
// 		c.Log.WithError(err).Error("error updating purchase")
// 		return nil, errors.New("internal server error")
// 	}

// 	debt := new(entity.Debt)
// 	if err := c.DebtRepository.FindByPurchaseCodeOrInit(tx, debt, purchase.Code); err != nil {
// 		c.Log.WithError(err).Error("error getting debt")
// 		return nil, errors.New("not found")
// 	}

// 	if debt.ID != 0 {
// 		if err := c.DebtRepository.UpdateStatus(tx, debt.ID); err != nil {
// 			c.Log.WithError(err).Error("error update debt")
// 			return nil, errors.New("internal server error")
// 		}
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		c.Log.WithError(err).Error("error updating purchase")
// 		return nil, errors.New("internal server error")
// 	}

// 	return purchase, nil
// }

func (c *PurchaseUseCase) Cancel(ctx context.Context, request *model.CancelPurchaseRequest) (*model.PurchaseResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.Rollback()

	// validate input
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	// 1) Load purchase entity (lock) - must return entity.Purchase with BranchID, CreatedAt, Code, Status
	purchaseEntity := new(entity.Purchase)
	if err := c.PurchaseRepository.FindLockByCode(tx, request.Code, purchaseEntity); err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return nil, errors.New("not found")
	}

	if purchaseEntity.Status == "CANCELLED" {
		return nil, errors.New("purchase already cancelled")
	}

	// 2) check 24-hour window (use CreatedAt stored in millis)
	createdTime := time.UnixMilli(purchaseEntity.CreatedAt)
	if time.Since(createdTime).Hours() >= 24 {
		c.Log.WithField("purchase_code", purchaseEntity.Code).Error("error cancelling purchase: exceeded 24-hour window")
		return nil, errors.New("forbidden")
	}

	// 3) load purchase details (only not-cancelled)
	details, err := c.PurchaseDetailRepository.FindByPurchaseCode(tx, request.Code)
	if err != nil {
		c.Log.WithError(err).Error("error getting purchase details")
		return nil, errors.New("internal server error")
	}

	// 4) aggregate qty per size_id (avoid duplicate sizes)
	qtyBySize := make(map[uint]int, len(details))
	for _, d := range details {
		// only consider not-cancelled rows; repo should return only is_cancelled=0
		qtyBySize[d.SizeID] += d.Qty
	}

	// build sorted sizeIDs for deterministic lock order (reduce deadlock risk)
	sizeIDs := make([]uint, 0, len(qtyBySize))
	for sid := range qtyBySize {
		sizeIDs = append(sizeIDs, sid)
	}
	slices.Sort(sizeIDs)

	// 5) fetch & lock branch_inventory rows for this branch and sizes
	inventories, err := c.BranchInventoryRepository.FindByBranchAndSizeIDs(tx, purchaseEntity.BranchID, sizeIDs)
	if err != nil {
		c.Log.WithError(err).Error("error getting branch inventory")
		return nil, errors.New("internal server error")
	}
	if len(inventories) == 0 {
		// data integrity issue
		return nil, errors.New("branch inventory rows not found for purchase sizes")
	}

	// map size->inventory and also ensure stock enough
	invBySize := make(map[uint]entity.BranchInventory, len(inventories))
	for _, inv := range inventories {
		invBySize[inv.SizeID] = inv
	}

	// 6) optional: ensure no negative stock after decreasing
	for sid, qty := range qtyBySize {
		inv, ok := invBySize[sid]
		if !ok {
			return nil, fmt.Errorf("branch_inventory not found for size_id %d", sid)
		}
		if inv.Stock < qty {
			return nil, fmt.Errorf("insufficient stock for size_id %d: have %d need %d", sid, inv.Stock, qty)
		}
	}

	// 7) Bulk decrease stock with single UPDATE CASE (stock = stock - qty)
	if err := c.BranchInventoryRepository.BulkDecreaseStock(tx, purchaseEntity.BranchID, qtyBySize); err != nil {
		c.Log.WithError(err).Error("error decreasing branch inventory")
		return nil, errors.New("internal server error")
	}

	// 8) Insert inventory_movements (bulk) with negative change_qty (prealloc without append)
	// count valid movements
	mvCount := 0
	for sid, qty := range qtyBySize {
		if qty > 0 {
			if _, ok := invBySize[sid]; ok {
				mvCount++
			}
		}
	}
	movements := make([]entity.InventoryMovement, mvCount)
	idx := 0
	now := time.Now().UnixMilli()
	for sid, qty := range qtyBySize {
		if qty == 0 {
			continue
		}
		inv := invBySize[sid]
		movements[idx] = entity.InventoryMovement{
			BranchInventoryID: inv.ID,
			ChangeQty:         -qty, // negative because returning to distributor
			ReferenceType:     "PURCHASE_CANCEL",
			ReferenceKey:      purchaseEntity.Code,
			CreatedAt:         now,
		}
		idx++
	}
	if mvCount > 0 {
		if err := c.InventoryMovementRepository.CreateBulk(tx, movements); err != nil {
			c.Log.WithError(err).Error("error inserting inventory movements")
			return nil, errors.New("internal server error")
		}
	}

	// 9) Mark purchase_details as cancelled (single UPDATE)
	if err := c.PurchaseDetailRepository.Cancel(tx, purchaseEntity.Code); err != nil {
		c.Log.WithError(err).Error("error updating purchase details")
		return nil, errors.New("internal server error")
	}

	// 10) Delete purchase_payments (single query)
	if err := c.PurchasePaymentRepository.DeleteByCode(tx, purchaseEntity.Code); err != nil {
		c.Log.WithError(err).Error("error deleting purchase payments")
		return nil, errors.New("internal server error")
	}

	// 11) Handle debts: delete debt_payments (by debt ids) & mark debts VOID
	// 11a) pluck debt ids for this purchase
	debtIDs, err := c.DebtRepository.FindPluckByPurchaseCode(tx, purchaseEntity.Code)
	if err != nil {
		c.Log.WithError(err).Error("error fetching debts")
		return nil, errors.New("internal server error")
	}
	if len(debtIDs) > 0 {
		if err := c.DebtPaymentRepository.DeleteINDebtID(tx, debtIDs); err != nil {
			c.Log.WithError(err).Error("error deleting debt payments")
			return nil, errors.New("internal server error")
		}
		if err := c.DebtRepository.VoidByPurchaseCode(tx, purchaseEntity.Code); err != nil {
			c.Log.WithError(err).Error("error updating debts to VOID")
			return nil, errors.New("internal server error")
		}
	}

	// 12) Update purchase status to CANCELLED
	if err := c.PurchaseRepository.Cancel(tx, purchaseEntity.Code); err != nil {
		c.Log.WithError(err).Error("error updating purchase status")
		return nil, errors.New("internal server error")
	}

	// commit
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error committing cancel purchase")
		return nil, errors.New("internal server error")
	}

	// finally return the purchase response (fresh)
	resp, err := c.PurchaseRepository.FindByCode(c.DB, purchaseEntity.Code)
	if err != nil {
		// commit succeeded but fetch failed — still return success-ish but with minimal info
		c.Log.WithError(err).Warn("cancel succeeded but failed to fetch response")
		return &model.PurchaseResponse{Code: purchaseEntity.Code, Status: "CANCELLED"}, nil
	}
	return resp, nil
}

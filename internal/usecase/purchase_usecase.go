package usecase

import (
	"context"
	"fmt"
	"slices"
	"strconv"
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
	if err := c.Validate.Struct(request); err != nil {
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if request.Debt == nil && len(request.Payments) == 0 || (request.Debt != nil && len(request.Payments) > 0) {
		c.Log.Error("debt or payments must be provided")
		return nil, model.NewAppErr("bad request", "either debt or payments must be provided")
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
		c.Log.WithError(err).Error("error getting inventories")
		return nil, helper.GetNotFoundMessage("inventories", err)
	}

	branchInvMap := make(map[uint]*entity.BranchInventory, len(branchInvs))
	for i := range branchInvs {
		branchInvMap[branchInvs[i].SizeID] = &branchInvs[i]
	}

	// Buat kode purchase
	purchaseCode := "PURCHASE-" + time.Now().Format("20060102150405")

	// Build purchase details & movements
	n := len(request.Details)
	details := make([]entity.PurchaseDetail, n)
	movements := make([]entity.InventoryMovement, n)
	var totalPrice float64
	now := time.Now().UnixMilli()
	buyPriceBySize := map[uint]float64{}

	for i, d := range request.Details {
		inv, exists := branchInvMap[d.SizeID]
		if !exists {
			c.Log.WithField("size_id", d.SizeID).Errorf("size not found")
			return nil, model.NewAppErr("size not found", "stok tidak tersedia")
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
		c.Log.WithError(err).Error("error creating inventories")
		return nil, model.NewAppErr("internal server error", nil)
	}

	// Insert purchase
	purchase := entity.Purchase{
		Code:          purchaseCode,
		BranchID:      request.BranchID,
		DistributorID: request.DistributorID,
		SalesName:     request.SalesName,
		TotalPrice:    totalPrice,
	}

	if err := c.PurchaseRepository.Create(tx, &purchase); err != nil {
		c.Log.WithError(err).Error("error creating purchase")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := c.PurchaseDetailRepository.CreateBulk(tx, details); err != nil {
		c.Log.WithError(err).Error("error creating purchase details")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := c.InventoryMovementRepository.CreateBulk(tx, movements); err != nil {
		c.Log.WithError(err).Error("error creating inventory movements")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := c.SizeRepository.BulkUpdateBuyPrice(tx, buyPriceBySize); err != nil {
		c.Log.WithError(err).Error("error updating sizes")
		return nil, model.NewAppErr("internal server error", nil)
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
			c.Log.Error("total payment is less than total price")
			return nil, model.NewAppErr("validation error", "total payment is less than total price")
		}

		if err := c.PurchasePaymentRepository.CreateBulk(tx, payments); err != nil {
			c.Log.WithError(err).Error("error creating purchase payments")
			return nil, model.NewAppErr("internal server error", nil)
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
			c.Log.WithError(err).Error("error creating cash bank transaction")
			return nil, model.NewAppErr("internal server error", nil)
		}
	}

	// Insert debt (jika ada)
	if request.Debt != nil {
		var paidAmount float64
		for _, p := range request.Debt.DebtPayments {
			paidAmount += p.Amount
		}

		if paidAmount > totalPrice {
			c.Log.Error("total debt payment is more than total price")
			return nil, model.NewAppErr("validation error", "total debt payment is more than total price")
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
			c.Log.WithError(err).Error("error creating debt")
			return nil, model.NewAppErr("internal server error", nil)

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
				c.Log.WithError(err).Error("error creating debt payments")
				return nil, model.NewAppErr("internal server error", nil)

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
				c.Log.WithError(err).Error("error cash bank transaction branch")
				return nil, model.NewAppErr("internal server error", nil)

			}
		}
	}

	// Commit transaksi
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating purchase")
		return nil, model.NewAppErr("internal server error", nil)
	}
	return converter.PurchaseToResponse(&purchase), nil
}

func (c *PurchaseUseCase) Search(ctx context.Context, request *model.SearchPurchaseRequest) ([]model.PurchaseResponse, int64, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	purchases, total, err := c.PurchaseRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting purchases")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting purchases")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	responses := make([]model.PurchaseResponse, len(purchases))
	for i, purchase := range purchases {
		responses[i] = *converter.PurchaseToResponse(&purchase)
	}

	return responses, total, nil
}

func (c *PurchaseUseCase) Get(ctx context.Context, request *model.GetPurchaseRequest) (*model.PurchaseResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	purchase, err := c.PurchaseRepository.FindByCode(tx, request.Code)
	if err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return nil, helper.GetNotFoundMessage("purchase", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return purchase, nil
}

func (c *PurchaseUseCase) Cancel(ctx context.Context, request *model.CancelPurchaseRequest) error {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	purchaseEntity := new(entity.Purchase)
	if err := c.PurchaseRepository.FindLockByCode(tx, request.Code, purchaseEntity); err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return helper.GetNotFoundMessage("purchase", err)
	}

	createdTime := time.UnixMilli(purchaseEntity.CreatedAt)
	if time.Since(createdTime).Hours() >= 24 {
		c.Log.WithField("purchase_code", purchaseEntity.Code).Error("purchase cannot be deleted after 24 hours")
		return model.NewAppErr("forbidden", "purchase cannot be deleted after 24 hours")
	}

	if purchaseEntity.Status == "CANCELLED" {
		c.Log.WithField("purchase_code", purchaseEntity.Code).Error("purchase already cancelled")
		return model.NewAppErr("conflict", "purchase already cancelled")
	}

	details, err := c.PurchaseDetailRepository.FindByPurchaseCode(tx, request.Code)
	if err != nil {
		c.Log.WithError(err).Error("error getting purchase details")
		return helper.GetNotFoundMessage("purchase details", err)
	}

	qtyBySize := make(map[uint]int, len(details))
	for _, d := range details {
		qtyBySize[d.SizeID] += d.Qty
	}

	if len(qtyBySize) == 0 {
		c.Log.WithField("purchase_code", purchaseEntity.Code).Error("all details already cancelled")
		return model.NewAppErr("conflict", "all purchase detail already cancelled")
	}

	sizeIDs := make([]uint, 0, len(qtyBySize))
	for sid := range qtyBySize {
		sizeIDs = append(sizeIDs, sid)
	}
	slices.Sort(sizeIDs)

	inventories, err := c.BranchInventoryRepository.FindByBranchAndSizeIDs(tx, purchaseEntity.BranchID, sizeIDs)
	if err != nil {
		c.Log.WithError(err).Error("error getting inventories")
		return helper.GetNotFoundMessage("inventory", err)
	}

	lastPrices, err := c.PurchaseDetailRepository.FindLastBuyPricesBySizeIDs(tx, sizeIDs, purchaseEntity.Code)
	if err != nil {
		c.Log.WithError(err).Error("error get last buy prices")
		return helper.GetNotFoundMessage("purchase detail", err)
	}

	if len(lastPrices) > 0 {
		if err := c.SizeRepository.BulkUpdateBuyPrice(tx, lastPrices); err != nil {
			c.Log.WithError(err).Error("error bulk update buy price")
			return model.NewAppErr("internal server error", nil)
		}
	}

	invBySize := make(map[uint]entity.BranchInventory, len(inventories))
	for _, inv := range inventories {
		invBySize[inv.SizeID] = inv
	}
	for sid, qty := range qtyBySize {
		inv, ok := invBySize[sid]
		if !ok {
			c.Log.WithField("size_id", sid).Errorf("size not found")
			return model.NewAppErr("size not found", fmt.Sprintf("inventories not found for size_id %d", sid))
		}
		if inv.Stock < qty {
			c.Log.WithField("size_id", sid).Errorf("size insufficient stock")
			return model.NewAppErr("validation not match", fmt.Sprintf("insufficient stock for size_id %d: have %d need %d", sid, inv.Stock, qty))
		}
	}

	if err := c.BranchInventoryRepository.BulkDecreaseStock(tx, purchaseEntity.BranchID, qtyBySize); err != nil {
		c.Log.WithError(err).Error("error update inventories")
		return model.NewAppErr("internal server error", nil)
	}

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
			ChangeQty:         -qty,
			ReferenceType:     "PURCHASE_CANCELLED",
			ReferenceKey:      purchaseEntity.Code,
			CreatedAt:         now,
		}
		idx++
	}
	if mvCount > 0 {
		if err := c.InventoryMovementRepository.CreateBulk(tx, movements); err != nil {
			c.Log.WithError(err).Error("error creating inventory movements")
			return model.NewAppErr("internal server error", nil)
		}
	}

	if err := c.PurchaseDetailRepository.Cancel(tx, purchaseEntity.Code); err != nil {
		c.Log.WithError(err).Error("error updating purchase details")
		return model.NewAppErr("internal server error", nil)
	}

	debt := new(entity.Debt)
	if err := c.DebtRepository.FindByPurchaseCodeOrInit(tx, debt, purchaseEntity.Code); err != nil {
		c.Log.WithError(err).Error("error getting debt")
		return helper.GetNotFoundMessage("debt", err)
	}

	if debt.ID != 0 {
		if debt.PaidAmount > 0 {
			refundTx := entity.CashBankTransaction{
				TransactionDate: time.Now().UnixMilli(),
				Type:            "IN",
				Source:          "PURCHASE_DEBT_CANCELLED",
				Amount:          debt.PaidAmount,
				Description:     "Refund karena pembatalan pembelian " + purchaseEntity.Code,
				ReferenceKey:    purchaseEntity.Code,
				BranchID:        &purchaseEntity.BranchID,
			}
			if err := c.CashBankTransactionRepository.Create(tx, &refundTx); err != nil {
				c.Log.WithError(err).Error("error creating cash bank transaction")
				return model.NewAppErr("internal server error", nil)
			}
		}

		debt.TotalAmount = 0
		debt.PaidAmount = 0
		debt.Status = "VOID"
		if err := c.DebtRepository.Update(tx, debt); err != nil {
			c.Log.WithError(err).Error("error update debt")
			return model.NewAppErr("internal server error", nil)
		}

	} else {
		refundTx := entity.CashBankTransaction{
			TransactionDate: time.Now().UnixMilli(),
			Type:            "IN",
			Source:          "PURCHASE_CANCELLED",
			Amount:          purchaseEntity.TotalPrice,
			Description:     "Refund karena pembatalan pembelian " + purchaseEntity.Code,
			ReferenceKey:    purchaseEntity.Code,
			BranchID:        &purchaseEntity.BranchID,
		}
		if err := c.CashBankTransactionRepository.Create(tx, &refundTx); err != nil {
			c.Log.WithError(err).Error("error create cash bank transaction")
			return model.NewAppErr("internal server error", nil)
		}
	}

	if err := c.PurchaseRepository.Cancel(tx, purchaseEntity.Code); err != nil {
		c.Log.WithError(err).Error("error updating purchase")
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error cancel purchase")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

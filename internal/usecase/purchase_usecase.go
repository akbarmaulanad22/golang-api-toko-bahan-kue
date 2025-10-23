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

	// Kumpulkan qty per branch_inventory_id
	qtyByInv := map[uint]int{}
	for _, d := range request.Details {
		qtyByInv[d.BranchInventoryID] += d.Qty
	}

	// Ambil inventory
	invIDs := make([]uint, 0, len(qtyByInv))
	for id := range qtyByInv {
		invIDs = append(invIDs, id)
	}
	slices.Sort(invIDs)

	branchInvs, err := c.BranchInventoryRepository.FindByIDs(tx, invIDs)
	if err != nil {
		c.Log.WithError(err).Error("error getting inventories")
		return nil, helper.GetNotFoundMessage("inventories", err)
	}

	branchInvMap := make(map[uint]*entity.BranchInventory, len(branchInvs))
	for i := range branchInvs {
		branchInvMap[branchInvs[i].ID] = &branchInvs[i]
	}

	// Buat kode purchase
	purchaseCode := "PURCHASE-" + time.Now().Format("20060102150405")

	// Build purchase details & movements
	n := len(request.Details)
	details := make([]entity.PurchaseDetail, n)
	movements := make([]entity.InventoryMovement, n)
	var totalPrice float64
	now := time.Now().UnixMilli()
	buyPriceBySize := map[uint]float64{} // pakai SizeID sebagai key

	for i, d := range request.Details {
		inv, exists := branchInvMap[d.BranchInventoryID]
		if !exists {
			c.Log.WithField("branch_inventory_id", d.BranchInventoryID).Errorf("branch inventory not found")
			return nil, model.NewAppErr("inventory not found", "stok tidak tersedia")
		}

		totalPrice += d.BuyPrice * float64(d.Qty)

		details[i] = entity.PurchaseDetail{
			PurchaseCode:      purchaseCode,
			BranchInventoryID: d.BranchInventoryID,
			Qty:               d.Qty,
			BuyPrice:          d.BuyPrice,
		}

		movements[i] = entity.InventoryMovement{
			BranchInventoryID: inv.ID,
			ChangeQty:         d.Qty,
			ReferenceType:     "PURCHASE",
			ReferenceKey:      purchaseCode,
			CreatedAt:         now,
		}

		// Simpan harga beli terbaru berdasarkan SizeID
		buyPriceBySize[inv.SizeID] = d.BuyPrice
	}

	// Bulk update stok (1 query)
	if err := c.BranchInventoryRepository.BulkIncreaseStock(tx, branchInvs, qtyByInv); err != nil {
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
		return helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	purchase := new(entity.Purchase)
	if err := c.PurchaseRepository.FindLockByCode(tx, request.Code, purchase); err != nil {
		return helper.GetNotFoundMessage("purchase", err)
	}

	if time.Since(time.UnixMilli(purchase.CreatedAt)).Hours() >= 24 {
		return model.NewAppErr("forbidden", "purchase cannot be cancelled after 24 hours")
	}

	if purchase.Status == "CANCELLED" {
		return model.NewAppErr("conflict", "purchase already cancelled")
	}

	details, err := c.PurchaseDetailRepository.FindByPurchaseCode(tx, purchase.Code)
	if err != nil {
		return helper.GetNotFoundMessage("purchase details", err)
	}

	qtyByInv := map[uint]int{}
	invIDs := []uint{}
	for _, d := range details {
		if !d.IsCancelled {
			qtyByInv[d.BranchInventoryID] += d.Qty
			invIDs = append(invIDs, d.BranchInventoryID)
		}
	}

	if len(qtyByInv) == 0 {
		return model.NewAppErr("conflict", "all purchase details already cancelled")
	}

	inventories, err := c.BranchInventoryRepository.FindByIDs(tx, invIDs)
	if err != nil {
		return helper.GetNotFoundMessage("inventory", err)
	}

	invByID := map[uint]entity.BranchInventory{}
	sizeIDs := []uint{}
	for _, inv := range inventories {
		invByID[inv.ID] = inv
		sizeIDs = append(sizeIDs, inv.SizeID)
	}

	for invID, qty := range qtyByInv {
		if invByID[invID].Stock < qty {
			return model.NewAppErr("validation not match", fmt.Sprintf(
				"insufficient stock for branch_inventory_id %d: have %d need %d",
				invID, invByID[invID].Stock, qty,
			))
		}
	}

	lastPrices, err := c.PurchaseDetailRepository.FindLastBuyPricesBySizeIDs(tx, sizeIDs, purchase.Code)
	if err != nil {
		return helper.GetNotFoundMessage("purchase detail", err)
	}

	if len(lastPrices) > 0 {
		if err := c.SizeRepository.BulkUpdateBuyPrice(tx, lastPrices); err != nil {
			return model.NewAppErr("internal server error", nil)
		}
	}

	if err := c.BranchInventoryRepository.BulkDecreaseStock(tx, purchase.BranchID, qtyByInv); err != nil {
		return model.NewAppErr("internal server error", nil)
	}

	movements := []entity.InventoryMovement{}
	now := time.Now().UnixMilli()
	for invID, qty := range qtyByInv {
		if qty > 0 {
			movements = append(movements, entity.InventoryMovement{
				BranchInventoryID: invID,
				ChangeQty:         -qty,
				ReferenceType:     "PURCHASE_CANCELLED",
				ReferenceKey:      purchase.Code,
				CreatedAt:         now,
			})
		}
	}

	if len(movements) > 0 {
		if err := c.InventoryMovementRepository.CreateBulk(tx, movements); err != nil {
			return model.NewAppErr("internal server error", nil)
		}
	}

	if err := c.PurchaseDetailRepository.Cancel(tx, purchase.Code); err != nil {
		return model.NewAppErr("internal server error", nil)
	}

	debt := new(entity.Debt)
	if err := c.DebtRepository.FindByPurchaseCodeOrInit(tx, debt, purchase.Code); err != nil {
		return helper.GetNotFoundMessage("debt", err)
	}

	refundAmount := purchase.TotalPrice
	if debt.ID != 0 {
		refundAmount = debt.PaidAmount
		debt.TotalAmount = 0
		debt.PaidAmount = 0
		debt.Status = "VOID"
		if err := c.DebtRepository.Update(tx, debt); err != nil {
			return model.NewAppErr("internal server error", nil)
		}
	}

	refundTx := entity.CashBankTransaction{
		TransactionDate: now,
		Type:            "IN",
		Source:          "PURCHASE_CANCELLED",
		Amount:          refundAmount,
		Description:     fmt.Sprintf("Refund pembatalan purchase %s", purchase.Code),
		ReferenceKey:    purchase.Code,
		BranchID:        &purchase.BranchID,
	}
	if err := c.CashBankTransactionRepository.Create(tx, &refundTx); err != nil {
		return model.NewAppErr("internal server error", nil)
	}

	if err := c.PurchaseRepository.Cancel(tx, purchase.Code); err != nil {
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

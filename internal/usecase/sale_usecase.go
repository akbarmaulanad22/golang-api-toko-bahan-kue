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

type SaleUseCase struct {
	DB                            *gorm.DB
	Log                           *logrus.Logger
	Validate                      *validator.Validate
	SaleRepository                *repository.SaleRepository
	SaleDetailRepository          *repository.SaleDetailRepository
	SalePaymentRepository         *repository.SalePaymentRepository
	DebtRepository                *repository.DebtRepository
	DebtPaymentRepository         *repository.DebtPaymentRepository
	SizeRepository                *repository.SizeRepository
	CashBankTransactionRepository *repository.CashBankTransactionRepository
	BranchInventoryRepository     *repository.BranchInventoryRepository
	InventoryMovementRepository   *repository.InventoryMovementRepository
}

func NewSaleUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	saleRepository *repository.SaleRepository,
	saleDetailRepository *repository.SaleDetailRepository,
	salePaymentRepository *repository.SalePaymentRepository,
	debtRepository *repository.DebtRepository,
	debtPaymentRepository *repository.DebtPaymentRepository,
	sizeRepository *repository.SizeRepository,
	cashBankTransactionRepository *repository.CashBankTransactionRepository,
	branchInventoryRepository *repository.BranchInventoryRepository,
	inventoryMovementRepository *repository.InventoryMovementRepository,

) *SaleUseCase {
	return &SaleUseCase{
		DB:                            db,
		Log:                           logger,
		Validate:                      validate,
		SaleRepository:                saleRepository,
		DebtRepository:                debtRepository,
		SizeRepository:                sizeRepository,
		SaleDetailRepository:          saleDetailRepository,
		SalePaymentRepository:         salePaymentRepository,
		DebtPaymentRepository:         debtPaymentRepository,
		CashBankTransactionRepository: cashBankTransactionRepository,
		BranchInventoryRepository:     branchInventoryRepository,
		InventoryMovementRepository:   inventoryMovementRepository,
	}
}

func (c *SaleUseCase) Create(ctx context.Context, request *model.CreateSaleRequest) (*model.SaleResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if (request.Debt == nil && len(request.Payments) == 0) || (request.Debt != nil && len(request.Payments) > 0) {
		c.Log.Error("debt or payments must be provided exclusively")
		return nil, model.NewAppErr("bad request", "either debt or payments must be provided")
	}

	// Ambil semua branch_inventory sesuai request
	invIDs := make([]uint, len(request.Details))
	for i, d := range request.Details {
		invIDs[i] = d.BranchInventoryID
	}

	inventories, err := c.BranchInventoryRepository.FindByIDsWithSize(tx, invIDs)
	if err != nil {
		c.Log.WithError(err).Error("error getting inventories")
		return nil, model.NewAppErr("internal server error", nil)
	}

	invMap := make(map[uint]*entity.BranchInventory, len(inventories))
	for i := range inventories {
		invMap[inventories[i].ID] = &inventories[i]
	}

	// Buat kode sale
	saleCode := "SALE-" + time.Now().Format("20060102150405")

	n := len(request.Details)
	details := make([]entity.SaleDetail, n)
	movements := make([]entity.InventoryMovement, n)
	qtyByBranchInventory := make(map[uint]int, n)
	var totalPrice float64
	now := time.Now().UnixMilli()

	for i, d := range request.Details {
		inv, ok := invMap[d.BranchInventoryID]
		if !ok {
			c.Log.WithField("branch_inventory_id", d.BranchInventoryID).Error("inventory not found")
			return nil, model.NewAppErr("validation error", "stok tidak ditemukan")
		}

		if inv.Stock < d.Qty {
			c.Log.WithField("branch_inventory_id", d.BranchInventoryID).Error("insufficient stock")
			return nil, model.NewAppErr("validation error", "stok tidak cukup")
		}

		totalPrice += inv.Size.SellPrice * float64(d.Qty)
		qtyByBranchInventory[d.BranchInventoryID] += d.Qty

		details[i] = entity.SaleDetail{
			SaleCode:          saleCode,
			BranchInventoryID: d.BranchInventoryID,
			Qty:               d.Qty,
			SellPrice:         inv.Size.SellPrice,
		}

		movements[i] = entity.InventoryMovement{
			BranchInventoryID: inv.ID,
			ChangeQty:         -d.Qty,
			ReferenceType:     "SALE",
			ReferenceKey:      saleCode,
			CreatedAt:         now,
		}
	}

	if err := c.BranchInventoryRepository.BulkDecreaseStockNew(tx, qtyByBranchInventory); err != nil {
		c.Log.WithError(err).Error("error updating stock")
		return nil, model.NewAppErr("internal server error", nil)
	}

	sale := entity.Sale{
		Code:         saleCode,
		BranchID:     request.BranchID,
		CustomerName: request.CustomerName,
		TotalPrice:   totalPrice,
		CreatedAt:    now,
	}

	if err := c.SaleRepository.Create(tx, &sale); err != nil {
		c.Log.WithError(err).Error("error creating sale")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := c.SaleDetailRepository.CreateBulk(tx, details); err != nil {
		c.Log.WithError(err).Error("error creating sale details")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := c.InventoryMovementRepository.CreateBulk(tx, movements); err != nil {
		c.Log.WithError(err).Error("error creating inventory movements")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if n := len(request.Payments); n > 0 {
		payments := make([]entity.SalePayment, n)
		var totalPayment float64
		for i, p := range request.Payments {
			totalPayment += p.Amount
			payments[i] = entity.SalePayment{
				SaleCode:      saleCode,
				PaymentMethod: p.PaymentMethod,
				Amount:        p.Amount,
				Note:          p.Note,
				CreatedAt:     now,
			}
		}

		if totalPayment < totalPrice {
			return nil, model.NewAppErr("validation error", "total payment is less than total price")
		}

		if err := c.SalePaymentRepository.CreateBulk(tx, payments); err != nil {
			return nil, model.NewAppErr("internal server error", nil)
		}

		cashBankTransaction := entity.CashBankTransaction{
			TransactionDate: now,
			Type:            "IN",
			Source:          "SALE",
			Amount:          totalPrice,
			ReferenceKey:    saleCode,
			BranchID:        &request.BranchID,
		}
		if err := c.CashBankTransactionRepository.Create(tx, &cashBankTransaction); err != nil {
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
			return nil, model.NewAppErr("validation error", "total debt payment is more than total price")
		}

		debt := entity.Debt{
			ReferenceType: "SALE",
			ReferenceCode: saleCode,
			TotalAmount:   totalPrice,
			PaidAmount:    paidAmount,
			DueDate: func() int64 {
				if request.Debt.DueDate > 0 {
					return int64(request.Debt.DueDate)
				}
				return time.Now().Add(7 * 24 * time.Hour).UnixMilli()
			}(),
			Status:    "PENDING",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := c.DebtRepository.Create(tx, &debt); err != nil {
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
				return nil, model.NewAppErr("internal server error", nil)
			}
			cashBankTransaction := entity.CashBankTransaction{
				TransactionDate: now,
				Type:            "IN",
				Source:          "DEBT",
				Amount:          paidAmount,
				Description:     "Bayar Hutang Cicilan Pertama",
				ReferenceKey:    strconv.Itoa(int(debt.ID)),
				BranchID:        &request.BranchID,
			}
			if err := c.CashBankTransactionRepository.Create(tx, &cashBankTransaction); err != nil {
				return nil, model.NewAppErr("internal server error", nil)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.SaleToResponse(&sale), nil
}

func (c *SaleUseCase) Search(ctx context.Context, request *model.SearchSaleRequest) ([]model.SaleResponse, int64, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	sales, total, err := c.SaleRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting sales")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting sales")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	responses := make([]model.SaleResponse, len(sales))
	for i, sale := range sales {
		responses[i] = *converter.SaleToResponse(&sale)
	}

	return responses, total, nil
}

func (c *SaleUseCase) Get(ctx context.Context, request *model.GetSaleRequest) (*model.SaleResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	sale, err := c.SaleRepository.FindByCode(tx, request.Code)
	if err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return nil, helper.GetNotFoundMessage("sale", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return sale, nil
}

func (c *SaleUseCase) Cancel(ctx context.Context, request *model.CancelSaleRequest) error {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// === 1. Lock sale ===
	sale := new(entity.Sale)
	if err := c.SaleRepository.FindLockByCode(tx, request.Code, sale); err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return helper.GetNotFoundMessage("sale", err)
	}

	createdTime := time.UnixMilli(sale.CreatedAt)
	if time.Since(createdTime).Hours() >= 24 {
		c.Log.WithField("sale_code", sale.Code).Error("sale cannot be deleted after 24 hours")
		return model.NewAppErr("forbidden", "sale cannot be deleted after 24 hours")
	}

	if sale.Status == "CANCELLED" {
		c.Log.WithField("sale_code", sale.Code).Error("sale already cancelled")
		return model.NewAppErr("conflict", "sale already cancelled")
	}

	// === 2. Ambil sale detail (dengan branch_inventory) ===
	details, err := c.SaleDetailRepository.FindWithBranchInventory(tx, request.Code)
	if err != nil {
		c.Log.WithError(err).Error("error getting sale details")
		return helper.GetNotFoundMessage("sale details", err)
	}

	// === 3. Hitung qty per branch_inventory ===
	qtyByBranchInv := map[uint]int{}
	for _, d := range details {
		if !d.IsCancelled {
			qtyByBranchInv[d.BranchInventoryID] += d.Qty
		}
	}

	if len(qtyByBranchInv) == 0 {
		c.Log.WithField("sale_code", sale.Code).Error("all details already cancelled")
		return model.NewAppErr("conflict", "all sale detail already cancelled")
	}

	// === 4. Ambil inventory dengan lock ===
	branchInvIDs := make([]uint, 0, len(qtyByBranchInv))
	for id := range qtyByBranchInv {
		branchInvIDs = append(branchInvIDs, id)
	}
	slices.Sort(branchInvIDs)

	inventories, err := c.BranchInventoryRepository.FindByIDsWithSize(tx, branchInvIDs)
	if err != nil {
		c.Log.WithError(err).Error("error getting inventories")
		return helper.GetNotFoundMessage("inventories", err)
	}

	// === 5. Kembalikan stok (bulk increase) ===
	if err := c.BranchInventoryRepository.BulkIncreaseStockNew(tx, inventories, qtyByBranchInv); err != nil {
		c.Log.WithError(err).Error("error increasing inventories")
		return model.NewAppErr("internal server error", nil)
	}

	// === 6. Buat inventory movement ===
	movements := make([]entity.InventoryMovement, 0, len(inventories))
	for _, inv := range inventories {
		addQty := qtyByBranchInv[inv.ID]
		if addQty == 0 {
			continue
		}
		movements = append(movements, entity.InventoryMovement{
			BranchInventoryID: inv.ID,
			ChangeQty:         addQty,
			ReferenceType:     "SALE_CANCELLED",
			ReferenceKey:      request.Code,
			CreatedAt:         time.Now().UnixMilli(),
		})
	}
	if err := c.InventoryMovementRepository.CreateBulk(tx, movements); err != nil {
		c.Log.WithError(err).Error("error creating inventory movements")
		return model.NewAppErr("internal server error", nil)
	}

	// === 7. Update sale detail jadi cancelled ===
	if err := c.SaleDetailRepository.Cancel(tx, request.Code); err != nil {
		c.Log.WithError(err).Error("error updating sale detail")
		return model.NewAppErr("internal server error", nil)
	}

	// === 8. Handle debt dan cashbank transaction ===
	debt := new(entity.Debt)
	if err := c.DebtRepository.FindBySaleCodeOrInit(tx, debt, request.Code); err != nil {
		c.Log.WithError(err).Error("error getting debt payment")
		return helper.GetNotFoundMessage("debt", err)
	}

	if debt.ID != 0 {
		// kalau ada hutang, void + keluarkan uang kembali
		if debt.PaidAmount > 0 {
			outTx := entity.CashBankTransaction{
				TransactionDate: time.Now().UnixMilli(),
				Type:            "OUT",
				Source:          "SALE_DEBT_CANCELLED",
				Amount:          debt.PaidAmount,
				Description:     fmt.Sprintf("Pembatalan penjualan berutang %s", sale.Code),
				ReferenceKey:    sale.Code,
				BranchID:        &sale.BranchID,
				CreatedAt:       time.Now().UnixMilli(),
				UpdatedAt:       time.Now().UnixMilli(),
			}
			if err := c.CashBankTransactionRepository.Create(tx, &outTx); err != nil {
				c.Log.WithError(err).Error("error creating cash bank transaction")
				return model.NewAppErr("internal server error", nil)
			}
		}

		debt.Status = "VOID"
		debt.TotalAmount = 0
		debt.PaidAmount = 0
		if err := c.DebtRepository.Update(tx, debt); err != nil {
			c.Log.WithError(err).Error("error updating debt")
			return model.NewAppErr("internal server error", nil)
		}
	} else {
		// kalau non-debt
		outTx := entity.CashBankTransaction{
			TransactionDate: time.Now().UnixMilli(),
			Type:            "OUT",
			Source:          "SALE_CANCELLED",
			Amount:          sale.TotalPrice,
			Description:     fmt.Sprintf("Pembatalan penjualan %s", sale.Code),
			ReferenceKey:    sale.Code,
			BranchID:        &sale.BranchID,
			CreatedAt:       time.Now().UnixMilli(),
			UpdatedAt:       time.Now().UnixMilli(),
		}
		if err := c.CashBankTransactionRepository.Create(tx, &outTx); err != nil {
			c.Log.WithError(err).Error("error creating cash bank transaction")
			return model.NewAppErr("internal server error", nil)
		}
	}

	// === 9. Update status sale jadi CANCELLED ===
	if err := c.SaleRepository.Cancel(tx, sale.Code); err != nil {
		c.Log.WithError(err).Error("error updating sale")
		return model.NewAppErr("internal server error", nil)
	}

	// === 10. Commit ===
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error cancel sale")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

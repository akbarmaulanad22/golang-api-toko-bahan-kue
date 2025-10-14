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
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		return nil, helper.GetValidationMessage(err)
	}

	if request.Debt == nil && len(request.Payments) == 0 || (request.Debt != nil && len(request.Payments) > 0) {
		return nil, model.NewAppErr("bad request", "either debt or payments must be provided")
	}

	// Ambil harga per size
	sizeIDs := make([]uint, len(request.Details))
	for i, d := range request.Details {
		sizeIDs[i] = d.SizeID
	}

	sizesWithProduct, err := c.SizeRepository.FindSizesWithProductByIDs(tx, sizeIDs)
	if err != nil {
		return nil, model.NewAppErr("internal server error", nil)
	}
	sizeMap := make(map[uint]model.SizeWithProduct, len(sizesWithProduct))
	for _, s := range sizesWithProduct {
		sizeMap[s.ID] = s
	}

	// Ambil inventory
	branchInvs, err := c.BranchInventoryRepository.FindByBranchAndSizeIDs(tx, request.BranchID, sizeIDs)
	if err != nil {
		return nil, model.NewAppErr("internal server error", nil)
	}

	branchInvMap := make(map[uint]*entity.BranchInventory, len(branchInvs))
	for i := range branchInvs {
		branchInvMap[branchInvs[i].SizeID] = &branchInvs[i]
	}

	// Buat kode sale
	saleCode := "SALE-" + time.Now().Format("20060102150405")

	// Build sale details & movements
	n := len(request.Details)
	details := make([]entity.SaleDetail, n)
	movements := make([]entity.InventoryMovement, n)
	var totalPrice float64
	now := time.Now().UnixMilli()

	// Juga siapkan qtyBySize untuk BulkDecreaseStock
	qtyBySize := make(map[uint]int, len(request.Details))

	for i, d := range request.Details {
		size, ok := sizeMap[d.SizeID]
		if !ok {
			return nil, model.NewAppErr("size not found", nil)
		}
		inv, exists := branchInvMap[d.SizeID]
		if !exists {
			return nil, model.NewAppErr("size not found", fmt.Sprintf("size %s tidak tersedia pada product %s", size.Size, size.ProductName))
		}
		if inv.Stock < d.Qty {
			return nil, model.NewAppErr("validation error", "stok tidak cukup")
		}

		totalPrice += size.SellPrice * float64(d.Qty)
		qtyBySize[d.SizeID] += d.Qty

		details[i] = entity.SaleDetail{
			SaleCode:  saleCode,
			SizeID:    d.SizeID,
			Qty:       d.Qty,
			SellPrice: size.SellPrice,
		}
		movements[i] = entity.InventoryMovement{
			BranchInventoryID: inv.ID,
			ChangeQty:         -d.Qty,
			ReferenceType:     "SALE",
			ReferenceKey:      saleCode,
			CreatedAt:         now,
		}
	}

	// Bulk update stok (pakai qtyBySize)
	if err := c.BranchInventoryRepository.BulkDecreaseStock(tx, request.BranchID, qtyBySize); err != nil {
		return nil, model.NewAppErr("internal server error", nil)

	}

	// Insert sale
	sale := entity.Sale{
		Code:         saleCode,
		BranchID:     request.BranchID,
		CustomerName: request.CustomerName,
		TotalPrice:   totalPrice,
		CreatedAt:    now,
	}
	if err := c.SaleRepository.Create(tx, &sale); err != nil {
		return nil, model.NewAppErr("internal server error", nil)

	}
	if err := c.SaleDetailRepository.CreateBulk(tx, details); err != nil {
		return nil, model.NewAppErr("internal server error", nil)

	}
	if err := c.InventoryMovementRepository.CreateBulk(tx, movements); err != nil {
		return nil, model.NewAppErr("internal server error", nil)

	}

	// Insert payments (jika ada)
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
			return nil, model.NewAppErr("bad request", "total payment is less than total price")
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
			return nil, model.NewAppErr("bad request", "total debt payment is more than total price")
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

	// Commit transaksi
	if err := tx.Commit().Error; err != nil {
		return nil, model.NewAppErr("internal server error", nil)
	}
	return converter.SaleToResponse(&sale), nil
}

func (c *SaleUseCase) Search(ctx context.Context, request *model.SearchSaleRequest) ([]model.SaleResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

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
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

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
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	sale := new(entity.Sale)
	if err := c.SaleRepository.FindLockByCode(tx, request.Code, sale); err != nil {
		c.Log.WithError(err).Errorf("SALE_CANCEL: sale %s not found", request.Code)
		return helper.GetNotFoundMessage("sale", err)
	}

	createdTime := time.UnixMilli(sale.CreatedAt)
	if time.Since(createdTime).Hours() >= 24 {
		c.Log.WithField("sale", sale.Code).Error("SALE_CANCEL: exceeded 24-hour window")
		return model.NewAppErr("forbidden", "sale cannot be deleted after 24 hours")
	}

	if sale.Status == "CANCELLED" {
		c.Log.Warnf("SALE_CANCEL: sale %s already cancelled", request.Code)
		return model.NewAppErr("conflict", "sale already cancelled")
	}

	details, err := c.SaleDetailRepository.FindBySaleCode(tx, request.Code)
	if err != nil {
		c.Log.WithError(err).Errorf("SALE_CANCEL: failed to find sale details %s", request.Code)
		return helper.GetNotFoundMessage("sale details", err)
	}

	qtyBySize := map[uint]int{}
	for _, d := range details {
		if !d.IsCancelled {
			qtyBySize[d.SizeID] += d.Qty
		}
	}

	if len(qtyBySize) == 0 {
		c.Log.Warnf("SALE_CANCEL: all details already cancelled %s", request.Code)
		return model.NewAppErr("conflict", "all sale detail already cancelled")
	}

	sizeIDs := make([]uint, 0, len(qtyBySize))
	for sid := range qtyBySize {
		sizeIDs = append(sizeIDs, sid)
	}
	slices.Sort(sizeIDs)

	inventories, err := c.BranchInventoryRepository.FindByBranchAndSizeIDs(tx, sale.BranchID, sizeIDs)
	if err != nil {
		c.Log.WithError(err).Errorf("SALE_CANCEL: failed get branch_inventory branch=%d, sizes=%+v", sale.BranchID, sizeIDs)
		return helper.GetNotFoundMessage("inventory", err)
	}

	if err := c.BranchInventoryRepository.BulkIncreaseStock(tx, inventories, qtyBySize); err != nil {
		c.Log.WithError(err).Errorf("SALE_CANCEL: bulk increase stock failed sale=%s", sale.Code)
		return model.NewAppErr("internal server error", nil)
	}

	movements := make([]entity.InventoryMovement, 0, len(inventories))
	for _, inv := range inventories {
		addQty := qtyBySize[inv.SizeID]
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
		c.Log.WithError(err).Errorf("SALE_CANCEL: failed create inventory_movements sale=%s", sale.Code)
		return model.NewAppErr("internal server error", nil)
	}

	if err := c.SaleDetailRepository.Cancel(tx, request.Code); err != nil {
		c.Log.WithError(err).Errorf("SALE_CANCEL: failed cancel sale_details %s", sale.Code)
		return model.NewAppErr("internal server error", nil)
	}

	debt := new(entity.Debt)
	if err := c.DebtRepository.FindBySaleCodeOrInit(tx, debt, request.Code); err != nil {
		c.Log.WithError(err).Errorf("SALE_CANCEL: failed get debt sale=%s", sale.Code)
		return helper.GetNotFoundMessage("debt", err)
	}

	if debt.ID != 0 {
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
				c.Log.WithError(err).Errorf("SALE_CANCEL: failed create cashbank transaction debt sale=%s", sale.Code)
				return model.NewAppErr("internal server error", nil)
			}

		}

		debt.Status = "VOID"
		debt.TotalAmount = 0
		debt.PaidAmount = 0
		if err := c.DebtRepository.Update(tx, debt); err != nil {
			c.Log.WithError(err).Errorf("SALE_CANCEL: failed update debt sale=%s", sale.Code)
			return model.NewAppErr("internal server error", nil)
		}
	} else {
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
			c.Log.WithError(err).Errorf("SALE_CANCEL: failed create cashbank transaction sale=%s", sale.Code)
			return model.NewAppErr("internal server error", nil)
		}
	}

	if err := c.SaleRepository.Cancel(tx, sale.Code); err != nil {
		c.Log.WithError(err).Errorf("SALE_CANCEL: failed update sale status %s", sale.Code)
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Errorf("SALE_CANCEL: failed commit transaction sale=%s", sale.Code)
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

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

	if request.Debt == nil && len(request.Payments) == 0 {
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
			Amount:          totalPayment,
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

func (c *PurchaseUseCase) Cancel(ctx context.Context, request *model.CancelPurchaseRequest) (*model.PurchaseResponse, error) {
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

	createdTime := time.UnixMilli(purchase.CreatedAt)
	now := time.Now()

	// Hitung durasi sejak dibuat
	duration := now.Sub(createdTime)

	// Jika status BUKAN PENDING dan sudah lewat 24 jam => tolak
	if duration.Hours() >= 24 {
		c.Log.WithField("purchase_code", purchase.Code).Error("error updating purchase: exceeded 24-hour window")
		return nil, errors.New("forbidden")
	}

	// Lanjut update status
	if err := c.PurchaseRepository.Cancel(tx, purchase.Code); err != nil {
		c.Log.WithError(err).Error("error updating purchase")
		return nil, errors.New("internal server error")
	}

	debt := new(entity.Debt)
	if err := c.DebtRepository.FindByPurchaseCodeOrInit(tx, debt, purchase.Code); err != nil {
		c.Log.WithError(err).Error("error getting debt")
		return nil, errors.New("not found")
	}

	if debt.ID != 0 {
		if err := c.DebtRepository.UpdateStatus(tx, debt.ID); err != nil {
			c.Log.WithError(err).Error("error update debt")
			return nil, errors.New("internal server error")
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating purchase")
		return nil, errors.New("internal server error")
	}

	return purchase, nil
}

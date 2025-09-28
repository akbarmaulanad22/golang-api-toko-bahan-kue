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
		return nil, errors.New("bad request")
	}

	if request.Debt == nil && len(request.Payments) == 0 {
		return nil, errors.New("bad request: either debt or payments must be provided")
	}

	// Ambil harga per size
	sizeIDs := make([]uint, len(request.Details))
	for i, d := range request.Details {
		sizeIDs[i] = d.SizeID
	}

	sizeMap, err := c.SizeRepository.FindPriceMapByIDs(tx, sizeIDs)
	if err != nil {
		return nil, errors.New("internal server error")
	}

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
	saleCode := "SALE-" + time.Now().Format("20060102150405")

	// Build sale details & movements
	n := len(request.Details)
	details := make([]entity.SaleDetail, n)
	movements := make([]entity.InventoryMovement, n)
	var totalPrice float64
	now := time.Now().UnixMilli()

	for i, d := range request.Details {
		price, ok := sizeMap[d.SizeID]
		if !ok {
			return nil, errors.New("invalid size id")
		}
		inv, exists := branchInvMap[d.SizeID]
		if !exists {
			return nil, fmt.Errorf("stok untuk size_id %d tidak ditemukan di branch %d", d.SizeID, request.BranchID)
		}
		if inv.Stock < d.Qty {
			return nil, fmt.Errorf("stok tidak cukup untuk size_id %d (stok: %d, diminta: %d)", d.SizeID, inv.Stock, d.Qty)
		}

		totalPrice += price * float64(d.Qty)

		details[i] = entity.SaleDetail{
			SaleCode:  saleCode,
			SizeID:    d.SizeID,
			Qty:       d.Qty,
			SellPrice: price,
		}
		movements[i] = entity.InventoryMovement{
			BranchInventoryID: inv.ID,
			ChangeQty:         -d.Qty,
			ReferenceType:     "SALE",
			ReferenceKey:      saleCode,
			CreatedAt:         now,
		}
	}

	// Bulk update stok (1 query)
	if err := c.BranchInventoryRepository.BulkDecreaseStock(tx, request.BranchID, details); err != nil {
		return nil, err
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
		return nil, err
	}
	if err := c.SaleDetailRepository.CreateBulk(tx, details); err != nil {
		return nil, err
	}
	if err := c.InventoryMovementRepository.CreateBulk(tx, movements); err != nil {
		return nil, err
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
			return nil, errors.New("bad request: total payment is less than total price")
		}

		if err := c.SalePaymentRepository.CreateBulk(tx, payments); err != nil {
			return nil, err
		}

		cashBankTransaction := entity.CashBankTransaction{
			TransactionDate: now,
			Type:            "IN",
			Source:          "SALE",
			Amount:          totalPayment,
			ReferenceKey:    saleCode,
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
				Type:            "IN",
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
	return converter.SaleToResponse(&sale), nil
}

func (c *SaleUseCase) Search(ctx context.Context, request *model.SearchSaleRequest) ([]model.SaleResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	sales, total, err := c.SaleRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting sales")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting sales")
		return nil, 0, errors.New("internal server error")
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
		return nil, errors.New("bad request")
	}

	sale, err := c.SaleRepository.FindByCode(tx, request.Code)
	if err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return nil, errors.New("internal server error")
	}

	return sale, nil
}

func (c *SaleUseCase) Cancel(ctx context.Context, request *model.CancelSaleRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// 1) Lock sale
	sale := new(entity.Sale)
	if err := c.SaleRepository.FindLockByCode(tx, request.Code, sale); err != nil {
		return fmt.Errorf("sale not found: %w", err)
	}
	if sale.Status == "CANCELLED" {
		return errors.New("sale already cancelled")
	}

	// 2) Get sale_details (not cancelled yet)
	details, err := c.SaleDetailRepository.FindBySaleCode(tx, request.Code)
	if err != nil {
		return fmt.Errorf("sale details not found: %w", err)
	}

	// 3) Hitung qty per size_id
	qtyBySize := map[uint]int{}
	for _, d := range details {
		qtyBySize[d.SizeID] += d.Qty
	}
	sizeIDs := make([]uint, 0, len(qtyBySize))
	for sid := range qtyBySize {
		sizeIDs = append(sizeIDs, sid)
	}
	slices.Sort(sizeIDs)

	// 4) Ambil branch_inventory (lock)
	inventories, err := c.BranchInventoryRepository.FindByBranchAndSizeIDs(tx, sale.BranchID, sizeIDs)
	if err != nil {
		return err
	}

	// 5) Bulk update branch_inventory.stock (CASE WHEN)
	if err := c.BranchInventoryRepository.BulkIncreaseStock(tx, inventories, qtyBySize); err != nil {
		return err
	}

	// 6) Insert inventory_movements (bulk insert)
	movements := make([]entity.InventoryMovement, 0, len(inventories))
	for _, inv := range inventories {
		addQty := qtyBySize[inv.SizeID]
		if addQty == 0 {
			continue
		}
		movements = append(movements, entity.InventoryMovement{
			BranchInventoryID: inv.ID,
			ChangeQty:         addQty,
			ReferenceType:     "SALE_CANCEL",
			ReferenceKey:      request.Code,
		})
	}

	if err := c.InventoryMovementRepository.CreateBulk(tx, movements); err != nil {
		return err
	}

	// 7) Update sale_details jadi cancelled (batch GORM)
	if err := c.SaleDetailRepository.Cancel(tx, request.Code); err != nil {
		return err
	}

	// 8) Hapus sale_payments (batch GORM)
	if err := c.SalePaymentRepository.DeleteBySaleCode(tx, request.Code); err != nil {
		return err
	}

	// 9) Hapus debt_payments + void debt (GORM batch)
	// 9a) Ambil debts yang berkaitan (pluck id)
	debtIDs, err := c.DebtRepository.FindBySaleCode(tx, request.Code)
	if err != nil {
		return err
	}

	if len(debtIDs) > 0 {
		// Hapus debt_payments untuk debt_ids ini (single query)
		if err := c.DebtPaymentRepository.DeleteINDebtID(tx, debtIDs); err != nil {
			return err
		}

		if err := c.CashBankTransactionRepository.DeleteByDebtID(tx, debtIDs); err != nil {
			return err
		}

	}

	// 9b) Update debts menjadi VOID + paid_amount = 0
	if err := c.DebtRepository.FindBySaleCodeAndVoid(tx, request.Code); err != nil {
		return err
	}

	// 10) create cash bank transaction
	cashBankTransaction := entity.CashBankTransaction{
		TransactionDate: time.Now().UnixMilli(),
		Type:            "OUT",
		Source:          "SALE",
		Amount:          sale.TotalPrice,
		Description:     "PEMNJUALAN DIBATALKAN",
		ReferenceKey:    sale.Code,
		BranchID:        &sale.BranchID,
	}

	if err := c.CashBankTransactionRepository.Create(tx, &cashBankTransaction); err != nil {
		return err
	}

	// 10) delete cash bank transaction
	// if err := c.CashBankTransactionRepository.DeleteBySaleCode(tx, request.Code); err != nil {
	// 	return err
	// }

	// 11) Update sale status jadi CANCELLED
	if err := c.SaleRepository.Cancel(tx, request.Code); err != nil {
		return err
	}

	return tx.Commit().Error
}

// func (c *SaleUseCase) Cancel(ctx context.Context, request *model.CancelSaleRequest) (*model.SaleResponse, error) {
// 	tx := c.DB.WithContext(ctx).Begin()
// 	defer tx.Rollback()

// 	// Validasi request
// 	if err := c.Validate.Struct(request); err != nil {
// 		c.Log.WithError(err).Error("error validating request body")
// 		return nil, errors.New("bad request")
// 	}

// 	// Cari sale
// 	sale, err := c.SaleRepository.FindByCode(tx, request.Code)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error getting sale")
// 		return nil, errors.New("not found")
// 	}

// 	// Cek batas waktu cancel
// 	createdTime := time.UnixMilli(sale.CreatedAt)
// 	if time.Since(createdTime).Hours() >= 24 {
// 		c.Log.WithField("sale_code", sale.Code).Error("error updating sale: exceeded 24-hour window")
// 		return nil, errors.New("forbidden")
// 	}

// 	// Update status sale
// 	if err := c.SaleRepository.Cancel(tx, sale.Code); err != nil {
// 		c.Log.WithError(err).Error("error updating sale")
// 		return nil, errors.New("internal server error")
// 	}

// 	// Update status debt kalau ada
// 	debt := new(entity.Debt)
// 	if err := c.DebtRepository.FindBySaleCodeOrInit(tx, debt, sale.Code); err != nil {
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
// 		c.Log.WithError(err).Error("error commit transaction")
// 		return nil, errors.New("internal server error")
// 	}

// 	return sale, nil
// }

// func (c *SaleUseCase) Cancel(ctx context.Context, request *model.CancelSaleRequest) (*model.SaleResponse, error) {
// 	tx := c.DB.WithContext(ctx).Begin()
// 	defer tx.Rollback()

// 	if err := c.Validate.Struct(request); err != nil {
// 		c.Log.WithError(err).Error("error validating request body")
// 		return nil, errors.New("bad request")
// 	}

// 	sale, err := c.SaleRepository.FindByCode(tx, request.Code)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error getting sale")
// 		return nil, errors.New("not found")
// 	}

// 	createdTime := time.UnixMilli(sale.CreatedAt)
// 	now := time.Now()

// 	// Hitung durasi sejak dibuat
// 	duration := now.Sub(createdTime)

// 	// Jika status BUKAN PENDING dan sudah lewat 24 jam => tolak
// 	if duration.Hours() >= 24 {
// 		c.Log.WithField("sale_code", sale.Code).Error("error updating sale: exceeded 24-hour window")
// 		return nil, errors.New("forbidden")
// 	}

// 	// Lanjut update status
// 	if err := c.SaleRepository.Cancel(tx, sale.Code); err != nil {
// 		c.Log.WithError(err).Error("error updating sale")
// 		return nil, errors.New("internal server error")
// 	}

// 	debt := new(entity.Debt)
// 	if err := c.DebtRepository.FindBySaleCodeOrInit(tx, debt, sale.Code); err != nil {
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
// 		c.Log.WithError(err).Error("error updating sale")
// 		return nil, errors.New("internal server error")
// 	}

// 	return sale, nil
// }

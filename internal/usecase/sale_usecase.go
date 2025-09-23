package usecase

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
	"tokobahankue/internal/model/converter"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
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
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	// validasi minimal salah satu (Debt / Payments) harus ada
	if request.Debt == nil && len(request.Payments) == 0 {
		return nil, errors.New("bad request: either debt or payments must be provided")
	}

	// ambil size harga dalam bentuk map
	sizeIDs := make([]uint, 0, len(request.Details))
	for _, d := range request.Details {
		sizeIDs = append(sizeIDs, d.SizeID)
	}

	sizeMap, err := c.SizeRepository.FindPriceMapByIDs(tx, sizeIDs)
	if err != nil {
		return nil, errors.New("internal server error")
	}

	// buat code
	saleCode := "SALE-" + time.Now().Format("20060102150405")

	// buat sale details + total harga
	details := make([]entity.SaleDetail, 0, len(request.Details))
	var totalPrice float64
	for _, d := range request.Details {
		price, ok := sizeMap[d.SizeID]
		if !ok {
			return nil, errors.New("invalid size id")
		}
		totalPrice += price * float64(d.Qty)
		details = append(details, entity.SaleDetail{
			SaleCode:  saleCode,
			SizeID:    d.SizeID,
			Qty:       d.Qty,
			SellPrice: price,
		})

		branchInv := &entity.BranchInventory{}
		err := c.BranchInventoryRepository.FindByBranchIDAndSizeID(tx, branchInv, request.BranchID, d.SizeID)

		if err != nil {
			c.Log.WithError(err).Error("error querying branch inventory")
			return nil, errors.New("internal server error")
		}

		if err := c.BranchInventoryRepository.UpdateStock(tx, branchInv.ID, -d.Qty); err != nil {
			return nil, errors.New("error updating stock")
		}

		// Catat movement
		movement := &entity.InventoryMovement{
			BranchInventoryID: branchInv.ID,
			ChangeQty:         -d.Qty,
			ReferenceType:     "SALE",
			ReferenceKey:      saleCode,
		}

		if err := c.InventoryMovementRepository.Create(tx, movement); err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); ok {
				switch mysqlErr.Number {
				case 1452:
					if strings.Contains(mysqlErr.Message, "FOREIGN KEY (`branch_inventory_id`)") {
						c.Log.Warn("branch inventory doesnt exists")
						return nil, errors.New("invalid branch inventory id")
					}
					return nil, errors.New("foreign key constraint failed")
				}
			}
			return nil, errors.New("error creating inventory movement")
		}
	}

	// buat sale
	sale := entity.Sale{
		Code:         saleCode,
		BranchID:     request.BranchID,
		CustomerName: request.CustomerName,
		TotalPrice:   totalPrice,
	}

	if err := c.SaleRepository.Create(tx, &sale); err != nil {

		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case 1452:
				if strings.Contains(mysqlErr.Message, "FOREIGN KEY (`branch_id`)") {
					c.Log.Warn("branch doesnt exists")
					return nil, errors.New("invalid branch id")
				}
				return nil, errors.New("foreign key constraint failed")
			}
		}

		return nil, err
	}

	if err := c.SaleDetailRepository.CreateBulk(tx, details); err != nil {
		return nil, err
	}

	// insert payments kalau ada
	if len(request.Payments) > 0 {
		// build sale payments + total bayar
		payments := make([]entity.SalePayment, 0, len(request.Payments))
		var totalPayment float64
		for _, p := range request.Payments {
			totalPayment += p.Amount
			payments = append(payments, entity.SalePayment{
				SaleCode:      saleCode,
				PaymentMethod: p.PaymentMethod,
				Amount:        p.Amount,
				Note:          p.Note,
			})
		}

		// kalau ada payments → cek jumlah payment >= totalPrice
		if totalPayment < totalPrice {
			return nil, errors.New("bad request: total payment is less than total price")
		}

		if err := c.SalePaymentRepository.CreateBulk(tx, payments); err != nil {
			return nil, err
		}

		// masukin ke catatan keuangan
		cashBankTransaction := entity.CashBankTransaction{
			TransactionDate: sale.CreatedAt,
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

	// insert debt kalau ada
	if request.Debt != nil {
		debt := entity.Debt{
			ReferenceType: "SALE",
			ReferenceCode: saleCode,
			TotalAmount:   totalPrice,
			PaidAmount:    0,
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

		// insert debt payments kalau ada
		if len(request.Debt.DebtPayments) > 0 {
			debtPayments := make([]entity.DebtPayment, 0, len(request.Debt.DebtPayments))
			for _, p := range request.Debt.DebtPayments {

				// totalin biaya yang dibayar
				debt.PaidAmount += p.Amount

				debtPayments = append(debtPayments, entity.DebtPayment{
					DebtID:      debt.ID,
					PaymentDate: sale.CreatedAt,
					// PaymentDate: time.Now().UnixMilli(),
					Amount: p.Amount,
					Note:   p.Note,
				})
			}

			if err := c.DebtPaymentRepository.CreateBulk(tx, debtPayments); err != nil {
				return nil, err
			}

			// masukin ke catatan keuangan
			cashBankTransaction := entity.CashBankTransaction{
				TransactionDate: sale.CreatedAt,
				Type:            "IN",
				Source:          "DEBT",
				Amount:          debt.PaidAmount,
				Description:     "Bayar Hutang Cicilan Pertama",
				ReferenceKey:    strconv.Itoa(int(debt.ID)),
				BranchID:        &request.BranchID,
			}

			if err := c.CashBankTransactionRepository.Create(tx, &cashBankTransaction); err != nil {
				return nil, err
			}

		}

		if err := c.DebtRepository.Update(tx, &debt); err != nil {
			return nil, err
		}
	}

	// commit transaksi
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error committing transaction")
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

func (c *SaleUseCase) Cancel(ctx context.Context, request *model.CancelSaleRequest) (*model.SaleResponse, error) {
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

	createdTime := time.UnixMilli(sale.CreatedAt)
	now := time.Now()

	// Hitung durasi sejak dibuat
	duration := now.Sub(createdTime)

	// Jika status BUKAN PENDING dan sudah lewat 24 jam => tolak
	if duration.Hours() >= 24 {
		c.Log.WithField("sale_code", sale.Code).Error("error updating sale: exceeded 24-hour window")
		return nil, errors.New("forbidden")
	}

	// Lanjut update status
	if err := c.SaleRepository.Cancel(tx, sale.Code); err != nil {
		c.Log.WithError(err).Error("error updating sale")
		return nil, errors.New("internal server error")
	}

	debt := new(entity.Debt)
	if err := c.DebtRepository.FindBySaleCodeOrInit(tx, debt, sale.Code); err != nil {
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
		c.Log.WithError(err).Error("error updating sale")
		return nil, errors.New("internal server error")
	}

	return sale, nil
}

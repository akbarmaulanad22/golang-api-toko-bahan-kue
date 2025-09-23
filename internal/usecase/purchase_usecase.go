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
	purchaseCode := "PURCHASE-" + time.Now().Format("20060102150405")

	// buat purchase details + total harga
	details := make([]entity.PurchaseDetail, 0, len(request.Details))
	var totalPrice float64
	for _, d := range request.Details {
		price, ok := sizeMap[d.SizeID]
		if !ok {
			return nil, errors.New("invalid size id")
		}
		totalPrice += price * float64(d.Qty)
		details = append(details, entity.PurchaseDetail{
			PurchaseCode: purchaseCode,
			SizeID:       d.SizeID,
			Qty:          d.Qty,
			BuyPrice:     price,
		})

		// ====================================================================

		branchInv := &entity.BranchInventory{}
		err := c.BranchInventoryRepository.FindByBranchIDAndSizeID(tx, branchInv, request.BranchID, d.SizeID)

		if err != nil {
			c.Log.WithError(err).Error("error querying branch inventory")
			return nil, errors.New("internal server error")
		}

		if err := c.BranchInventoryRepository.UpdateStock(tx, branchInv.ID, d.Qty); err != nil {
			return nil, errors.New("error updating stock")
		}

		// Catat movement
		movement := &entity.InventoryMovement{
			BranchInventoryID: branchInv.ID,
			ChangeQty:         d.Qty,
			ReferenceType:     "PURCHASE",
			ReferenceKey:      purchaseCode,
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

	// buat purchase
	purchase := entity.Purchase{
		Code:          purchaseCode,
		BranchID:      request.BranchID,
		DistributorID: request.DistributorID,
		SalesName:     request.SalesName,
		TotalPrice:    totalPrice,
	}

	if err := c.PurchaseRepository.Create(tx, &purchase); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case 1452:
				if strings.Contains(mysqlErr.Message, "FOREIGN KEY (`distributor_id`)") {
					c.Log.Warn("distributor doesnt exists")
					return nil, errors.New("invalid distributor id")
				}
				if strings.Contains(mysqlErr.Message, "FOREIGN KEY (`branch_id`)") {
					c.Log.Warn("branch doesnt exists")
					return nil, errors.New("invalid branch id")
				}
				return nil, errors.New("foreign key constraint failed")
			}
		}
		return nil, errors.New("error creating purchases")
	}

	if err := c.PurchaseDetailRepository.CreateBulk(tx, details); err != nil {
		return nil, err
	}

	// insert payments kalau ada
	if len(request.Payments) > 0 {
		// build purchase payments + total bayar
		payments := make([]entity.PurchasePayment, 0, len(request.Payments))
		var totalPayment float64
		for _, p := range request.Payments {
			totalPayment += p.Amount
			payments = append(payments, entity.PurchasePayment{
				PurchaseCode:  purchaseCode,
				PaymentMethod: p.PaymentMethod,
				Amount:        p.Amount,
				Note:          p.Note,
			})
		}

		// kalau ada payments → cek jumlah payment >= totalPrice
		if totalPayment < totalPrice {
			return nil, errors.New("bad request: total payment is less than total price")
		}

		if err := c.PurchasePaymentRepository.CreateBulk(tx, payments); err != nil {
			return nil, err
		}

		// masukin ke catatan keuangan
		cashBankTransaction := entity.CashBankTransaction{
			TransactionDate: purchase.CreatedAt,
			Type:            "IN",
			Source:          "PURCHASE",
			Amount:          totalPayment,
			ReferenceKey:    purchaseCode,
			BranchID:        &request.BranchID,
		}

		if err := c.CashBankTransactionRepository.Create(tx, &cashBankTransaction); err != nil {
			return nil, err
		}
	}

	// insert debt kalau ada
	if request.Debt != nil {
		debt := entity.Debt{
			ReferenceType: "PURCHASE",
			ReferenceCode: purchaseCode,
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
					PaymentDate: purchase.CreatedAt,
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
				TransactionDate: purchase.CreatedAt,
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

	return converter.PurchaseToResponse(&purchase), nil
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

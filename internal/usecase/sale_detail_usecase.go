package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SaleDetailUseCase struct {
	DB                            *gorm.DB
	Log                           *logrus.Logger
	Validate                      *validator.Validate
	SaleDetailRepository          *repository.SaleDetailRepository
	SaleRepository                *repository.SaleRepository
	InventoryMovementRepository   *repository.InventoryMovementRepository
	BranchInventoryRepository     *repository.BranchInventoryRepository
	CashBankTransactionRepository *repository.CashBankTransactionRepository
	DebtRepository                *repository.DebtRepository
}

func NewSaleDetailUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	saleDetailRepository *repository.SaleDetailRepository,
	saleRepository *repository.SaleRepository,
	inventoryMovementRepository *repository.InventoryMovementRepository,
	branchInventoryRepository *repository.BranchInventoryRepository,
	cashBankTransactionRepository *repository.CashBankTransactionRepository,
	debtRepository *repository.DebtRepository,
) *SaleDetailUseCase {
	return &SaleDetailUseCase{
		DB:                            db,
		Log:                           logger,
		Validate:                      validate,
		SaleDetailRepository:          saleDetailRepository,
		SaleRepository:                saleRepository,
		InventoryMovementRepository:   inventoryMovementRepository,
		BranchInventoryRepository:     branchInventoryRepository,
		CashBankTransactionRepository: cashBankTransactionRepository,
		DebtRepository:                debtRepository,
	}
}

func (c *SaleDetailUseCase) Cancel(ctx context.Context, request *model.CancelSaleDetailRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// 1) validate
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("SALE_DETAIL_CANCEL: invalid request")
		return errors.New("bad request")
	}
	c.Log.WithField("sale_code", request.SaleCode).WithField("size_id", request.SizeID).
		Info("SALE_DETAIL_CANCEL: start cancel sale detail")

	// 2) lock sale
	sale := new(entity.Sale)
	if err := c.SaleRepository.FindLockByCode(tx, request.SaleCode, sale); err != nil {
		c.Log.WithError(err).WithField("sale_code", request.SaleCode).Error("SALE_DETAIL_CANCEL: sale not found")
		return errors.New("not found")
	}
	c.Log.WithField("sale_code", sale.Code).WithField("status", sale.Status).Info("SALE_DETAIL_CANCEL: sale locked")

	// 3) check 24-hour window
	createdTime := time.UnixMilli(sale.CreatedAt)
	if time.Since(createdTime).Hours() >= 24 {
		c.Log.WithField("sale_code", sale.Code).Error("SALE_DETAIL_CANCEL: exceeded 24-hour window")
		return errors.New("forbidden")
	}

	// 4) get sale detail (price & qty)
	detail := new(entity.SaleDetail)
	if err := c.SaleDetailRepository.FindPriceBySizeIDAndSaleCode(tx, sale.Code, request.SizeID, detail); err != nil {
		c.Log.WithError(err).WithFields(logrus.Fields{"sale_code": sale.Code, "size_id": request.SizeID}).Error("SALE_DETAIL_CANCEL: sale detail not found")
		return errors.New("not found")
	}
	if detail.IsCancelled {
		c.Log.WithFields(logrus.Fields{"sale_code": sale.Code, "size_id": request.SizeID}).Warn("SALE_DETAIL_CANCEL: sale detail already cancelled")
		return errors.New("forbidden")
	}
	c.Log.WithFields(logrus.Fields{"sale_code": sale.Code, "size_id": detail.SizeID, "qty": detail.Qty, "sell_price": detail.SellPrice}).Info("SALE_DETAIL_CANCEL: detail fetched")

	// 5) mark detail cancelled
	if err := c.SaleDetailRepository.CancelBySizeID(tx, sale.Code, request.SizeID); err != nil {
		c.Log.WithError(err).WithFields(logrus.Fields{"sale_code": sale.Code, "size_id": request.SizeID}).Error("SALE_DETAIL_CANCEL: failed cancel detail")
		return errors.New("internal server error")
	}
	c.Log.WithFields(logrus.Fields{"sale_code": sale.Code, "size_id": request.SizeID}).Info("SALE_DETAIL_CANCEL: detail marked cancelled")

	// 6) restore stock (use existing FindByBranchIDAndSizeID + BulkIncreaseStock)
	branchInv := new(entity.BranchInventory)
	if err := c.BranchInventoryRepository.FindByBranchIDAndSizeID(tx, branchInv, sale.BranchID, detail.SizeID); err != nil {
		c.Log.WithError(err).WithFields(logrus.Fields{"branch_id": sale.BranchID, "size_id": detail.SizeID}).Error("SALE_DETAIL_CANCEL: branch inventory not found")
		return errors.New("internal server error")
	}

	// BulkIncreaseStock expects inventories slice + qty map (we pass single entry)
	inventories := []entity.BranchInventory{*branchInv}
	qtyMap := map[uint]int{detail.SizeID: detail.Qty}
	if err := c.BranchInventoryRepository.BulkIncreaseStock(tx, inventories, qtyMap); err != nil {
		c.Log.WithError(err).WithFields(logrus.Fields{"branch_id": sale.BranchID, "size_id": detail.SizeID, "qty": detail.Qty}).Error("SALE_DETAIL_CANCEL: failed increase stock")
		return errors.New("internal server error")
	}
	c.Log.WithFields(logrus.Fields{"branch_id": sale.BranchID, "size_id": detail.SizeID, "qty": detail.Qty}).Info("SALE_DETAIL_CANCEL: stock restored")

	// 7) create inventory movement
	movement := entity.InventoryMovement{
		BranchInventoryID: branchInv.ID,
		ChangeQty:         detail.Qty,
		ReferenceType:     "SALE_DETAIL_CANCELLED",
		ReferenceKey:      sale.Code,
		CreatedAt:         time.Now().UnixMilli(),
	}
	if err := c.InventoryMovementRepository.Create(tx, &movement); err != nil {
		c.Log.WithError(err).WithFields(logrus.Fields{"branch_inventory_id": branchInv.ID}).Error("SALE_DETAIL_CANCEL: failed create inventory movement")
		return errors.New("internal server error")
	}
	c.Log.WithField("movement_id", movement.ID).Info("SALE_DETAIL_CANCEL: inventory movement created")

	// 8) debt handling / refund logic
	cancelAmount := detail.SellPrice * float64(detail.Qty) // money to refund / reduce from sale
	debt := new(entity.Debt)
	if err := c.DebtRepository.FindBySaleCodeOrInit(tx, debt, sale.Code); err != nil {
		c.Log.WithError(err).WithField("sale_code", sale.Code).Error("SALE_DETAIL_CANCEL: failed to find debt")
		return errors.New("internal server error")
	}

	if debt.ID != 0 {
		c.Log.WithField("debt_id", debt.ID).Info("SALE_DETAIL_CANCEL: debt exists, adjust amounts")

		// Reduce total_amount
		debt.TotalAmount -= cancelAmount
		if debt.TotalAmount < 0 {
			debt.TotalAmount = 0
		}

		// If customer already paid more than new total -> refund the difference (OUT)
		if debt.PaidAmount > debt.TotalAmount {
			refund := debt.PaidAmount - debt.TotalAmount
			outTx := entity.CashBankTransaction{
				TransactionDate: time.Now().UnixMilli(),
				Type:            "OUT", // money going out to customer
				Source:          "SALE_DEBT_CANCELLED",
				Amount:          refund,
				Description:     fmt.Sprintf("Refund pembatalan penjualan per item %s", sale.Code),
				ReferenceKey:    sale.Code,
				BranchID:        &sale.BranchID,
				CreatedAt:       time.Now().UnixMilli(),
				UpdatedAt:       time.Now().UnixMilli(),
			}
			if err := c.CashBankTransactionRepository.Create(tx, &outTx); err != nil {
				c.Log.WithError(err).Error("SALE_DETAIL_CANCEL: failed create cashbank refund transaction")
				return errors.New("internal server error")
			}
			c.Log.WithFields(logrus.Fields{"sale_code": sale.Code, "refund": refund}).Info("SALE_DETAIL_CANCEL: refund created because paid > new total")
			// adjust paid amount
			debt.PaidAmount = debt.TotalAmount
		}

		// Update debt status
		if debt.TotalAmount == 0 && debt.PaidAmount == 0 {
			debt.Status = "VOID"
		} else if debt.PaidAmount >= debt.TotalAmount {
			debt.Status = "PAID"
		} else {
			debt.Status = "PENDING"
		}

		if err := c.DebtRepository.Update(tx, debt); err != nil {
			c.Log.WithError(err).WithField("debt_id", debt.ID).Error("SALE_DETAIL_CANCEL: failed update debt")
			return errors.New("internal server error")
		}
		c.Log.WithFields(logrus.Fields{"debt_id": debt.ID, "status": debt.Status}).Info("SALE_DETAIL_CANCEL: debt updated")
	} else {
		// no debt, immediately refund the cancelAmount to cash bank (OUT)
		outTx := entity.CashBankTransaction{
			TransactionDate: time.Now().UnixMilli(),
			Type:            "OUT", // refund to customer
			Source:          "SALE_CANCELLED",
			Amount:          cancelAmount,
			Description:     fmt.Sprintf("Refund pembatalan penjualan per item %s", sale.Code),
			ReferenceKey:    sale.Code,
			BranchID:        &sale.BranchID,
			CreatedAt:       time.Now().UnixMilli(),
			UpdatedAt:       time.Now().UnixMilli(),
		}
		if err := c.CashBankTransactionRepository.Create(tx, &outTx); err != nil {
			c.Log.WithError(err).Error("SALE_DETAIL_CANCEL: failed create cashbank refund transaction (no debt)")
			return errors.New("internal server error")
		}
		c.Log.WithFields(logrus.Fields{"sale_code": sale.Code, "refund": cancelAmount}).Info("SALE_DETAIL_CANCEL: refund created (no debt)")
	}

	// 9) update sale total_price (decrease)
	newTotal := sale.TotalPrice - cancelAmount
	if newTotal < 0 {
		newTotal = 0
	}
	if err := c.SaleRepository.UpdateTotalPrice(tx, sale.Code, newTotal); err != nil {
		c.Log.WithError(err).WithField("sale_code", sale.Code).Error("SALE_DETAIL_CANCEL: failed update sale total_price")
		return errors.New("internal server error")
	}
	c.Log.WithFields(logrus.Fields{"sale_code": sale.Code, "old_total": sale.TotalPrice, "new_total": newTotal}).Info("SALE_DETAIL_CANCEL: sale total_price updated")

	// 10) if all sale_details cancelled -> mark sale cancelled & debt void
	remaining, err := c.SaleDetailRepository.CountActiveBySaleCode(tx, sale.Code)
	if err != nil {
		c.Log.WithError(err).WithField("sale_code", sale.Code).Error("SALE_DETAIL_CANCEL: failed count active sale_details")
		return errors.New("internal server error")
	}
	if remaining == 0 {
		// update sale jadi CANCELLED
		if err := c.SaleRepository.Cancel(tx, sale.Code); err != nil {
			c.Log.WithError(err).WithField("sale_code", sale.Code).Error("SALE_DETAIL_CANCEL: failed update sale status to CANCELLED")
			return errors.New("internal server error")
		}
		c.Log.WithField("sale_code", sale.Code).Info("SALE_DETAIL_CANCEL: sale status updated to CANCELLED")

		// update debt jadi VOID
		if debt.ID != 0 {
			debt.Status = "VOID"
			debt.TotalAmount = 0
			debt.PaidAmount = 0

			if err := c.DebtRepository.Update(tx, debt); err != nil {
				c.Log.WithError(err).WithField("debt_id", debt.ID).Error("SALE_DETAIL_CANCEL: failed void debt")
				return errors.New("internal server error")
			}
			c.Log.WithFields(logrus.Fields{"debt_id": debt.ID, "status": debt.Status}).Info("SALE_DETAIL_CANCEL: debt set to VOID (all details cancelled)")
		}
	}

	// 11) commit
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).WithField("sale_code", sale.Code).Error("SALE_DETAIL_CANCEL: failed commit transaction")
		return errors.New("internal server error")
	}
	c.Log.WithField("sale_code", sale.Code).Info("SALE_DETAIL_CANCEL: completed successfully")

	return nil
}

package usecase

import (
	"context"
	"fmt"
	"time"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/helper"
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

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("SALE_DETAIL_CANCEL: invalid request")
		return helper.GetValidationMessage(err)
	}

	sale := new(entity.Sale)
	if err := c.SaleRepository.FindLockByCode(tx, request.SaleCode, sale); err != nil {
		c.Log.WithError(err).WithField("sale_code", request.SaleCode).Error("SALE_DETAIL_CANCEL: sale not found")
		return helper.GetNotFoundMessage("sale", err)
	}

	createdTime := time.UnixMilli(sale.CreatedAt)
	if time.Since(createdTime).Hours() >= 24 {
		c.Log.WithField("sale_code", sale.Code).Error("SALE_DETAIL_CANCEL: exceeded 24-hour window")
		return model.NewAppErr("forbidden", "sale cannot be deleted after 24 hours")
	}

	detail := new(entity.SaleDetail)
	if err := c.SaleDetailRepository.FindPriceBySizeIDAndSaleCode(tx, sale.Code, request.SizeID, detail); err != nil {
		c.Log.WithError(err).WithFields(logrus.Fields{"sale_code": sale.Code, "size_id": request.SizeID}).Error("SALE_DETAIL_CANCEL: sale detail not found")
		return helper.GetNotFoundMessage("sale details", err)
	}

	if detail.IsCancelled {
		c.Log.WithFields(logrus.Fields{"sale_code": sale.Code, "size_id": request.SizeID}).Warn("SALE_DETAIL_CANCEL: sale detail already cancelled")
		return model.NewAppErr("conflict", "sale detail already cancelled")
	}

	if err := c.SaleDetailRepository.CancelBySizeID(tx, sale.Code, request.SizeID); err != nil {
		c.Log.WithError(err).WithFields(logrus.Fields{"sale_code": sale.Code, "size_id": request.SizeID}).Error("SALE_DETAIL_CANCEL: failed cancel detail")
		return model.NewAppErr("internal server error", nil)
	}

	branchInv := new(entity.BranchInventory)
	if err := c.BranchInventoryRepository.FindByBranchIDAndSizeID(tx, branchInv, sale.BranchID, detail.SizeID); err != nil {
		c.Log.WithError(err).WithFields(logrus.Fields{"branch_id": sale.BranchID, "size_id": detail.SizeID}).Error("SALE_DETAIL_CANCEL: branch inventory not found")
		return helper.GetNotFoundMessage("inventory", err)
	}

	inventories := []entity.BranchInventory{*branchInv}
	qtyMap := map[uint]int{detail.SizeID: detail.Qty}
	if err := c.BranchInventoryRepository.BulkIncreaseStock(tx, inventories, qtyMap); err != nil {
		c.Log.WithError(err).WithFields(logrus.Fields{"branch_id": sale.BranchID, "size_id": detail.SizeID, "qty": detail.Qty}).Error("SALE_DETAIL_CANCEL: failed increase stock")
		return model.NewAppErr("internal server error", nil)
	}

	movement := entity.InventoryMovement{
		BranchInventoryID: branchInv.ID,
		ChangeQty:         detail.Qty,
		ReferenceType:     "SALE_DETAIL_CANCELLED",
		ReferenceKey:      sale.Code,
		CreatedAt:         time.Now().UnixMilli(),
	}
	if err := c.InventoryMovementRepository.Create(tx, &movement); err != nil {
		c.Log.WithError(err).WithFields(logrus.Fields{"branch_inventory_id": branchInv.ID}).Error("SALE_DETAIL_CANCEL: failed create inventory movement")
		return model.NewAppErr("internal server error", nil)
	}

	cancelAmount := detail.SellPrice * float64(detail.Qty) // money to refund / reduce from sale
	debt := new(entity.Debt)
	if err := c.DebtRepository.FindBySaleCodeOrInit(tx, debt, sale.Code); err != nil {
		c.Log.WithError(err).WithField("sale_code", sale.Code).Error("SALE_DETAIL_CANCEL: failed to find debt")
		return model.NewAppErr("internal server error", nil)
	}

	if debt.ID != 0 {
		debt.TotalAmount -= cancelAmount
		if debt.TotalAmount < 0 {
			debt.TotalAmount = 0
		}

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
				return model.NewAppErr("internal server error", nil)
			}

			debt.PaidAmount = debt.TotalAmount
		}

		if debt.TotalAmount == 0 && debt.PaidAmount == 0 {
			debt.Status = "VOID"
		} else if debt.PaidAmount >= debt.TotalAmount {
			debt.Status = "PAID"
		} else {
			debt.Status = "PENDING"
		}

		if err := c.DebtRepository.Update(tx, debt); err != nil {
			c.Log.WithError(err).WithField("debt_id", debt.ID).Error("SALE_DETAIL_CANCEL: failed update debt")
			return model.NewAppErr("internal server error", nil)
		}
	} else {
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
			return model.NewAppErr("internal server error", nil)
		}
	}

	newTotal := sale.TotalPrice - cancelAmount
	if newTotal < 0 {
		newTotal = 0
	}
	if err := c.SaleRepository.UpdateTotalPrice(tx, sale.Code, newTotal); err != nil {
		c.Log.WithError(err).WithField("sale_code", sale.Code).Error("SALE_DETAIL_CANCEL: failed update sale total_price")
		return model.NewAppErr("internal server error", nil)
	}

	remaining, err := c.SaleDetailRepository.CountActiveBySaleCode(tx, sale.Code)
	if err != nil {
		c.Log.WithError(err).WithField("sale_code", sale.Code).Error("SALE_DETAIL_CANCEL: failed count active sale_details")
		return model.NewAppErr("internal server error", nil)
	}
	if remaining == 0 {
		if err := c.SaleRepository.Cancel(tx, sale.Code); err != nil {
			c.Log.WithError(err).WithField("sale_code", sale.Code).Error("SALE_DETAIL_CANCEL: failed update sale status to CANCELLED")
			return model.NewAppErr("internal server error", nil)
		}

		if debt.ID != 0 {
			debt.Status = "VOID"
			debt.TotalAmount = 0
			debt.PaidAmount = 0

			if err := c.DebtRepository.Update(tx, debt); err != nil {
				c.Log.WithError(err).WithField("debt_id", debt.ID).Error("SALE_DETAIL_CANCEL: failed void debt")
				return model.NewAppErr("internal server error", nil)
			}
		}
	}

	// 11) commit
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).WithField("sale_code", sale.Code).Error("SALE_DETAIL_CANCEL: failed commit transaction")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

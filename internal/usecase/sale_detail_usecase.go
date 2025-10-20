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
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	sale := new(entity.Sale)
	if err := c.SaleRepository.FindLockByCode(tx, request.SaleCode, sale); err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return helper.GetNotFoundMessage("sale", err)
	}

	createdTime := time.UnixMilli(sale.CreatedAt)
	if time.Since(createdTime).Hours() >= 24 {
		c.Log.WithField("sale_code", sale.Code).Error("sale cannot be deleted after 24 hours")
		return model.NewAppErr("forbidden", "sale cannot be deleted after 24 hours")
	}

	detail := new(entity.SaleDetail)
	if err := c.SaleDetailRepository.FindPriceByID(tx, request.ID, detail); err != nil {
		c.Log.WithError(err).Error("error getting sale detail")
		return helper.GetNotFoundMessage("sale detail", err)
	}

	if detail.IsCancelled {
		c.Log.WithField("sale_detail_id", detail.ID).Error("sale detail already cancelled")
		return model.NewAppErr("conflict", "sale detail already cancelled")
	}

	if err := c.SaleDetailRepository.CancelByCodeAndID(tx, request.SaleCode, request.ID); err != nil {
		c.Log.WithError(err).Error("error updating sale detail")
		return helper.GetNotFoundMessage("sale detail", err)
	}

	branchInv := new(entity.BranchInventory)
	if err := c.BranchInventoryRepository.FindByBranchIDAndSizeID(tx, branchInv, sale.BranchID, detail.SizeID); err != nil {
		c.Log.WithError(err).Error("error getting inventories")
		return helper.GetNotFoundMessage("inventory", err)
	}

	inventories := []entity.BranchInventory{*branchInv}
	qtyMap := map[uint]int{detail.SizeID: detail.Qty}
	if err := c.BranchInventoryRepository.BulkIncreaseStock(tx, inventories, qtyMap); err != nil {
		c.Log.WithError(err).Error("error updating inventories")
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
		c.Log.WithError(err).Error("error creating inventory movement")
		return model.NewAppErr("internal server error", nil)
	}

	cancelAmount := detail.SellPrice * float64(detail.Qty) // money to refund / reduce from sale
	debt := new(entity.Debt)
	if err := c.DebtRepository.FindBySaleCodeOrInit(tx, debt, sale.Code); err != nil {
		c.Log.WithError(err).Error("error getting debt")
		return helper.GetNotFoundMessage("debt", err)
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
				c.Log.WithError(err).Error("error creating cash bank transaction")
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
			c.Log.WithError(err).Error("error updating debt")
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
			c.Log.WithError(err).Error("error creating cash bank transaction")
			return model.NewAppErr("internal server error", nil)
		}
	}

	newTotal := sale.TotalPrice - cancelAmount
	if newTotal < 0 {
		newTotal = 0
	}
	if err := c.SaleRepository.UpdateTotalPrice(tx, sale.Code, newTotal); err != nil {
		c.Log.WithError(err).Error("error updating cash bank transaction")
		return model.NewAppErr("internal server error", nil)
	}

	remaining, err := c.SaleDetailRepository.CountActiveBySaleCode(tx, sale.Code)
	if err != nil {
		c.Log.WithError(err).Error("error counting sale detail")
		return model.NewAppErr("internal server error", nil)
	}

	if remaining == 0 {
		if err := c.SaleRepository.Cancel(tx, sale.Code); err != nil {
			c.Log.WithError(err).Error("error updating sale")
			return model.NewAppErr("internal server error", nil)
		}

		if debt.ID != 0 {
			debt.Status = "VOID"
			debt.TotalAmount = 0
			debt.PaidAmount = 0

			if err := c.DebtRepository.Update(tx, debt); err != nil {
				c.Log.WithError(err).Error("error updating debt")
				return model.NewAppErr("internal server error", nil)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error cancel sale detail")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

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

type PurchaseDetailUseCase struct {
	DB                            *gorm.DB
	Log                           *logrus.Logger
	Validate                      *validator.Validate
	PurchaseDetailRepository      *repository.PurchaseDetailRepository
	PurchaseRepository            *repository.PurchaseRepository
	BranchInventoryRepository     *repository.BranchInventoryRepository
	InventoryMovementRepository   *repository.InventoryMovementRepository
	SizeRepository                *repository.SizeRepository
	DebtRepository                *repository.DebtRepository
	CashBankTransactionRepository *repository.CashBankTransactionRepository
	PurchasePaymentRepository     *repository.PurchasePaymentRepository
}

func NewPurchaseDetailUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	purchaseDetailRepository *repository.PurchaseDetailRepository,
	purchaseRepository *repository.PurchaseRepository,
	branchInventoryRepository *repository.BranchInventoryRepository,
	inventoryMovementRepository *repository.InventoryMovementRepository,
	sizeRepository *repository.SizeRepository,
	debtRepository *repository.DebtRepository,
	cashBankTransactionRepository *repository.CashBankTransactionRepository,
	purchasePaymentRepository *repository.PurchasePaymentRepository,
) *PurchaseDetailUseCase {
	return &PurchaseDetailUseCase{
		DB:                            db,
		Log:                           logger,
		Validate:                      validate,
		PurchaseDetailRepository:      purchaseDetailRepository,
		PurchaseRepository:            purchaseRepository,
		BranchInventoryRepository:     branchInventoryRepository,
		InventoryMovementRepository:   inventoryMovementRepository,
		SizeRepository:                sizeRepository,
		DebtRepository:                debtRepository,
		CashBankTransactionRepository: cashBankTransactionRepository,
		PurchasePaymentRepository:     purchasePaymentRepository,
	}
}

func (c *PurchaseDetailUseCase) Cancel(ctx context.Context, request *model.CancelPurchaseDetailRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	purchase := new(entity.Purchase)
	if err := c.PurchaseRepository.FindLockByCode(tx, request.PurchaseCode, purchase); err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return helper.GetNotFoundMessage("purchase", err)
	}

	createdTime := time.UnixMilli(purchase.CreatedAt)
	if time.Since(createdTime).Hours() >= 24 {
		c.Log.WithField("purchase_code", purchase.Code).Error("purchase cannot be deleted after 24 hours")
		return model.NewAppErr("forbidden", "purchase cannot be deleted after 24 hours")
	}

	detail := new(entity.PurchaseDetail)
	if err := c.PurchaseDetailRepository.FindPriceBySizeIDAndPurchaseCode(tx, purchase.Code, request.SizeID, detail); err != nil {
		c.Log.WithError(err).Error("error getting purchase detail")
		return helper.GetNotFoundMessage("purchase details", err)
	}
	if detail.IsCancelled {
		c.Log.WithField("purchase_code", purchase.Code).Error("purchase detail already cancelled")
		return model.NewAppErr("conflict", "purchase detail already cancelled")
	}

	if err := c.PurchaseDetailRepository.CancelBySizeID(tx, purchase.Code, request.SizeID); err != nil {
		c.Log.WithError(err).Error("error updating purchase detail")
		return model.NewAppErr("internal server error", nil)
	}

	inv := new(entity.BranchInventory)
	if err := c.BranchInventoryRepository.FindByBranchIDAndSizeID(tx, inv, purchase.BranchID, detail.SizeID); err != nil {
		c.Log.WithError(err).Error("error getting branch inventory")
		return helper.GetNotFoundMessage("inventory", err)
	}
	if inv.Stock < detail.Qty {
		c.Log.WithField("size_id", detail.SizeID).Error("insufficient stock")
		return model.NewAppErr("validation not match", fmt.Sprintf("insufficient stock for size_id %d: have %d need %d", detail.SizeID, inv.Stock, detail.Qty))
	}
	if err := c.BranchInventoryRepository.BulkDecreaseStock(tx, purchase.BranchID, map[uint]int{detail.SizeID: detail.Qty}); err != nil {
		c.Log.WithError(err).Error("error updating inventory stock")
		return model.NewAppErr("internal server error", nil)
	}

	mv := entity.InventoryMovement{
		BranchInventoryID: inv.ID,
		ChangeQty:         -detail.Qty,
		ReferenceType:     "PURCHASE_DETAIL_CANCELLED",
		ReferenceKey:      purchase.Code,
	}

	if err := c.InventoryMovementRepository.Create(tx, &mv); err != nil {
		c.Log.WithError(err).Error("error creating inventory movement")
		return model.NewAppErr("internal server error", nil)
	}

	lastPrices, err := c.PurchaseDetailRepository.FindLastBuyPricesBySizeIDs(tx, []uint{detail.SizeID}, purchase.Code)
	if err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return helper.GetNotFoundMessage("inventory", err)
	}

	if lastPrice, ok := lastPrices[detail.SizeID]; ok {
		if err := c.SizeRepository.BulkUpdateBuyPrice(tx, map[uint]float64{detail.SizeID: lastPrice}); err != nil {
			c.Log.WithError(err).Error("error updating size")
			return model.NewAppErr("internal server error", nil)
		}
	}

	debt := new(entity.Debt)
	if err := c.DebtRepository.FindByPurchaseCodeOrInit(tx, debt, purchase.Code); err != nil {
		c.Log.WithError(err).Error("error getting debt")
		return model.NewAppErr("internal server error", nil)
	}

	cancelAmount := detail.BuyPrice * float64(detail.Qty)

	if debt.ID != 0 {
		debt.TotalAmount -= cancelAmount
		if debt.TotalAmount < 0 {
			debt.TotalAmount = 0
		}

		if debt.PaidAmount > debt.TotalAmount {
			refund := debt.PaidAmount - debt.TotalAmount
			refundTx := entity.CashBankTransaction{
				TransactionDate: time.Now().UnixMilli(),
				Type:            "IN",
				Source:          "PURCHASE_DEBT_CANCELLED",
				Amount:          refund,
				Description:     fmt.Sprintf("Refund pembatalan sebagian %s", purchase.Code),
				ReferenceKey:    purchase.Code,
				BranchID:        &purchase.BranchID,
			}
			if err := c.CashBankTransactionRepository.Create(tx, &refundTx); err != nil {
				c.Log.WithError(err).Error("error creating cash bank transaction")
				return model.NewAppErr("internal server error", nil)
			}
			debt.PaidAmount = debt.TotalAmount
		}

		if debt.TotalAmount == 0 && debt.PaidAmount == 0 {
			debt.Status = "VOID"
		} else if debt.PaidAmount >= debt.TotalAmount {
			debt.Status = "PAID"

			payment := &entity.PurchasePayment{
				PurchaseCode:  purchase.Code,
				PaymentMethod: "CASH",
				Amount:        debt.TotalAmount,
			}
			if err := c.PurchasePaymentRepository.Create(tx, payment); err != nil {
				c.Log.WithError(err).Error("error creating purchase payment")
				return model.NewAppErr("internal server error", nil)
			}
		} else {
			debt.Status = "PENDING"
		}

		if err := c.DebtRepository.Update(tx, debt); err != nil {
			c.Log.WithError(err).Error("error updating debt")
			return model.NewAppErr("internal server error", nil)
		}
	} else {
		purchase.TotalPrice -= cancelAmount
		if err := c.PurchaseRepository.UpdateTotalPrice(tx, purchase.Code, purchase.TotalPrice); err != nil {
			c.Log.WithError(err).Error("error updating purchase")
			return model.NewAppErr("internal server error", nil)
		}

		refundTx := entity.CashBankTransaction{
			TransactionDate: time.Now().UnixMilli(),
			Type:            "IN",
			Source:          "PURCHASE_CANCELLED",
			Amount:          cancelAmount,
			Description:     fmt.Sprintf("Refund pembatalan sebagian %s", purchase.Code),
			ReferenceKey:    purchase.Code,
			BranchID:        &purchase.BranchID,
		}
		if err := c.CashBankTransactionRepository.Create(tx, &refundTx); err != nil {
			c.Log.WithError(err).Error("error creating cash bank transaction")
			return model.NewAppErr("internal server error", nil)
		}
	}

	newTotal := purchase.TotalPrice - (detail.BuyPrice * float64(detail.Qty))
	if newTotal < 0 {
		newTotal = 0
	}
	if err := c.PurchaseRepository.UpdateTotalPrice(tx, purchase.Code, newTotal); err != nil {
		c.Log.WithError(err).Error("error updating purchase")
		return model.NewAppErr("internal server error", nil)
	}

	remaining, err := c.PurchaseDetailRepository.CountActiveByPurchaseCode(tx, purchase.Code)
	if err != nil {
		c.Log.WithError(err).Error("error getting purchase detail")
		return model.NewAppErr("internal server error", nil)
	}
	if remaining == 0 {
		if err := c.PurchaseRepository.Cancel(tx, purchase.Code); err != nil {
			c.Log.WithError(err).Error("error updating purchase")
			return model.NewAppErr("internal server error", nil)
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error cancel purchase detail")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

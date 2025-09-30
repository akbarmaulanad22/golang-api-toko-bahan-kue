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

// func (c *PurchaseDetailUseCase) Cancel(ctx context.Context, request *model.CancelPurchaseDetailRequest) error {
// 	tx := c.DB.WithContext(ctx).Begin()
// 	defer tx.Rollback()

// 	// 1. Validasi input
// 	if err := c.Validate.Struct(request); err != nil {
// 		c.Log.WithError(err).Error("error validating request body")
// 		return errors.New("bad request")
// 	}

// 	// 2. Ambil purchase (lock)
// 	purchase := new(entity.Purchase)
// 	if err := c.PurchaseRepository.FindLockByCode(tx, request.PurchaseCode, purchase); err != nil {
// 		c.Log.WithError(err).Error("error getting purchase")
// 		return errors.New("not found")
// 	}

// 	// cek 24 jam
// 	createdTime := time.UnixMilli(purchase.CreatedAt)
// 	if time.Since(createdTime).Hours() >= 24 {
// 		c.Log.WithField("purchase_code", purchase.Code).
// 			Error("error cancelling purchase detail: exceeded 24-hour window")
// 		return errors.New("forbidden")
// 	}

// 	// 3. Ambil detail
// 	detail := new(entity.PurchaseDetail)
// 	if err := c.PurchaseDetailRepository.FindPriceBySizeIDAndPurchaseCode(tx, purchase.Code, request.SizeID, detail); err != nil {
// 		c.Log.WithError(err).Error("error getting purchase detail")
// 		return errors.New("not found")
// 	}
// 	if detail.IsCancelled {
// 		return errors.New("purchase detail already cancelled")
// 	}

// 	// 4. Update status cancelled
// 	if err := c.PurchaseDetailRepository.CancelBySizeID(tx, purchase.Code, request.SizeID); err != nil {
// 		c.Log.WithError(err).Error("error cancelling purchase detail")
// 		return errors.New("internal server error")
// 	}

// 	// 5. Update stock (branch_inventory)
// 	inv := new(entity.BranchInventory)
// 	if err := c.BranchInventoryRepository.FindByBranchIDAndSizeID(tx, inv, purchase.BranchID, detail.SizeID); err != nil {
// 		c.Log.WithError(err).Error("branch inventory not found")
// 		return errors.New("internal server error")
// 	}
// 	if inv.Stock < detail.Qty {
// 		return fmt.Errorf("insufficient stock for size_id %d", detail.SizeID)
// 	}
// 	if err := c.BranchInventoryRepository.BulkDecreaseStock(tx, purchase.BranchID, map[uint]int{detail.SizeID: detail.Qty}); err != nil {
// 		c.Log.WithError(err).Error("error decreasing stock")
// 		return errors.New("internal server error")
// 	}

// 	// 6. Insert inventory_movement
// 	mv := entity.InventoryMovement{
// 		BranchInventoryID: inv.ID,
// 		ChangeQty:         -detail.Qty,
// 		ReferenceType:     "PURCHASE_DETAIL_CANCELLED",
// 		ReferenceKey:      purchase.Code,
// 		CreatedAt:         time.Now().UnixMilli(),
// 	}
// 	if err := c.InventoryMovementRepository.Create(tx, &mv); err != nil {
// 		c.Log.WithError(err).Error("error inserting inventory movement")
// 		return errors.New("internal server error")
// 	}

// 	// 7. Update buy price kalau perlu
// 	lastPrices, err := c.PurchaseDetailRepository.FindLastBuyPricesBySizeIDs(tx, []uint{detail.SizeID}, purchase.Code)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error finding last buy price")
// 		return errors.New("internal server error")
// 	}
// 	if lastPrice, ok := lastPrices[detail.SizeID]; ok {
// 		if err := c.SizeRepository.BulkUpdateBuyPrice(tx, map[uint]float64{detail.SizeID: lastPrice}); err != nil {
// 			c.Log.WithError(err).Error("error updating buy price")
// 			return errors.New("internal server error")
// 		}
// 	}

// 	// 8. Debt handling
// 	debt := new(entity.Debt)
// 	if err := c.DebtRepository.FindByPurchaseCodeOrInit(tx, debt, purchase.Code); err != nil {
// 		c.Log.WithError(err).Error("error getting debt")
// 		return errors.New("internal server error")
// 	}

// 	cancelAmount := detail.BuyPrice * float64(detail.Qty)

// 	if debt.ID != 0 {
// 		// Kurangi total_amount
// 		debt.TotalAmount -= cancelAmount
// 		if debt.TotalAmount < 0 {
// 			debt.TotalAmount = 0
// 		}

// 		// Refund kalau overpaid
// 		if debt.PaidAmount > debt.TotalAmount {
// 			refund := debt.PaidAmount - debt.TotalAmount
// 			refundTx := entity.CashBankTransaction{
// 				TransactionDate: time.Now().UnixMilli(),
// 				Type:            "IN",
// 				Source:          "PURCHASE_DEBT_CANCELLED",
// 				Amount:          refund,
// 				Description:     fmt.Sprintf("Refund pembatalan sebagian %s", purchase.Code),
// 				ReferenceKey:    purchase.Code,
// 				BranchID:        &purchase.BranchID,
// 			}
// 			if err := c.CashBankTransactionRepository.Create(tx, &refundTx); err != nil {
// 				c.Log.WithError(err).Error("error insert refund transaction")
// 				return errors.New("internal server error")
// 			}
// 			debt.PaidAmount = debt.TotalAmount
// 		}

// 		if err := c.DebtRepository.Update(tx, debt); err != nil {
// 			c.Log.WithError(err).Error("error update debt")
// 			return errors.New("internal server error")
// 		}
// 	} else {
// 		// Ga ada debt → kurangi total_price + refund langsung
// 		purchase.TotalPrice -= cancelAmount
// 		if err := c.PurchaseRepository.UpdateTotalPrice(tx, purchase.Code, purchase.TotalPrice); err != nil {
// 			c.Log.WithError(err).Error("error update purchase total_price")
// 			return errors.New("internal server error")
// 		}

// 		refundTx := entity.CashBankTransaction{
// 			TransactionDate: time.Now().UnixMilli(),
// 			Type:            "IN",
// 			Source:          "PURCHASE_CANCELLED",
// 			Amount:          cancelAmount,
// 			Description:     fmt.Sprintf("Refund pembatalan sebagian %s", purchase.Code),
// 			ReferenceKey:    purchase.Code,
// 			BranchID:        &purchase.BranchID,
// 		}
// 		if err := c.CashBankTransactionRepository.Create(tx, &refundTx); err != nil {
// 			c.Log.WithError(err).Error("error insert refund transaction")
// 			return errors.New("internal server error")
// 		}
// 	}

// 	// 8. Update purchase total_price
// 	newTotal := purchase.TotalPrice - (detail.BuyPrice * float64(detail.Qty))
// 	if newTotal < 0 {
// 		newTotal = 0
// 	}
// 	if err := c.PurchaseRepository.UpdateTotalPrice(tx, purchase.Code, newTotal); err != nil {
// 		c.Log.WithError(err).Error("error updating purchase total_price")
// 		return errors.New("internal server error")
// 	}

// 	// 9. Jika semua detail sudah cancelled, update status purchase jadi CANCELLED
// 	remaining, err := c.PurchaseDetailRepository.CountActiveByPurchaseCode(tx, purchase.Code)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error counting remaining details")
// 		return errors.New("internal server error")
// 	}
// 	if remaining == 0 {
// 		if err := c.PurchaseRepository.Cancel(tx, purchase.Code); err != nil {
// 			c.Log.WithError(err).Error("error updating purchase status")
// 			return errors.New("internal server error")
// 		}
// 	}

// 	// 9. Commit
// 	if err := tx.Commit().Error; err != nil {
// 		c.Log.WithError(err).Error("error committing cancel purchase detail")
// 		return errors.New("internal server error")
// 	}

// 	return nil
// }
// ==============================

// func (c *PurchaseDetailUseCase) Cancel(ctx context.Context, request *model.CancelPurchaseDetailRequest) error {
// 	tx := c.DB.WithContext(ctx).Begin()
// 	defer tx.Rollback()

// 	// 1. Validasi input
// 	if err := c.Validate.Struct(request); err != nil {
// 		c.Log.WithError(err).Error("error validating request body")
// 		return errors.New("bad request")
// 	}

// 	// 2. Ambil purchase (lock)
// 	purchase := new(entity.Purchase)
// 	if err := c.PurchaseRepository.FindLockByCode(tx, request.PurchaseCode, purchase); err != nil {
// 		c.Log.WithError(err).Error("error getting purchase")
// 		return errors.New("not found")
// 	}

// 	// cek 24 jam
// 	createdTime := time.UnixMilli(purchase.CreatedAt)
// 	if time.Since(createdTime).Hours() >= 24 {
// 		c.Log.WithField("purchase_code", purchase.Code).
// 			Error("error cancelling purchase detail: exceeded 24-hour window")
// 		return errors.New("forbidden")
// 	}

// 	// 3. Ambil detail
// 	detail := new(entity.PurchaseDetail)
// 	if err := c.PurchaseDetailRepository.FindPriceBySizeIDAndPurchaseCode(tx, purchase.Code, request.SizeID, detail); err != nil {
// 		c.Log.WithError(err).Error("error getting purchase detail")
// 		return errors.New("not found")
// 	}
// 	if detail.IsCancelled {
// 		return errors.New("purchase detail already cancelled")
// 	}

// 	// 4. Update status cancelled
// 	if err := c.PurchaseDetailRepository.CancelBySizeID(tx, purchase.Code, request.SizeID); err != nil {
// 		c.Log.WithError(err).Error("error cancelling purchase detail")
// 		return errors.New("internal server error")
// 	}

// 	// 5. Update stock (branch_inventory)
// 	inv := new(entity.BranchInventory)
// 	if err := c.BranchInventoryRepository.FindByBranchIDAndSizeID(tx, inv, purchase.BranchID, detail.SizeID); err != nil {
// 		c.Log.WithError(err).Error("branch inventory not found")
// 		return errors.New("internal server error")
// 	}
// 	if inv.Stock < detail.Qty {
// 		return fmt.Errorf("insufficient stock for size_id %d", detail.SizeID)
// 	}
// 	if err := c.BranchInventoryRepository.BulkDecreaseStock(tx, purchase.BranchID, map[uint]int{detail.SizeID: detail.Qty}); err != nil {
// 		c.Log.WithError(err).Error("error decreasing stock")
// 		return errors.New("internal server error")
// 	}

// 	// 6. Insert inventory_movement
// 	mv := entity.InventoryMovement{
// 		BranchInventoryID: inv.ID,
// 		ChangeQty:         -detail.Qty,
// 		ReferenceType:     "PURCHASE_DETAIL_CANCELLED",
// 		ReferenceKey:      purchase.Code,
// 		CreatedAt:         time.Now().UnixMilli(),
// 	}
// 	if err := c.InventoryMovementRepository.Create(tx, &mv); err != nil {
// 		c.Log.WithError(err).Error("error inserting inventory movement")
// 		return errors.New("internal server error")
// 	}

// 	// 7. Update buy price kalau perlu
// 	lastPrices, err := c.PurchaseDetailRepository.FindLastBuyPricesBySizeIDs(tx, []uint{detail.SizeID}, purchase.Code)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error finding last buy price")
// 		return errors.New("internal server error")
// 	}
// 	if lastPrice, ok := lastPrices[detail.SizeID]; ok {
// 		if err := c.SizeRepository.BulkUpdateBuyPrice(tx, map[uint]float64{detail.SizeID: lastPrice}); err != nil {
// 			c.Log.WithError(err).Error("error updating buy price")
// 			return errors.New("internal server error")
// 		}
// 	}

// 	// 8. Debt handling
// 	cancelAmount := detail.BuyPrice * float64(detail.Qty)
// 	debt := new(entity.Debt)
// 	if err := c.DebtRepository.FindByPurchaseCodeOrInit(tx, debt, purchase.Code); err != nil {
// 		c.Log.WithError(err).Error("error getting debt")
// 		return errors.New("internal server error")
// 	}

// 	if debt.ID != 0 {
// 		// Kurangi total_amount
// 		debt.TotalAmount -= cancelAmount
// 		if debt.TotalAmount < 0 {
// 			debt.TotalAmount = 0
// 		}

// 		// Refund kalau overpaid
// 		if debt.PaidAmount > debt.TotalAmount {
// 			refund := debt.PaidAmount - debt.TotalAmount
// 			refundTx := entity.CashBankTransaction{
// 				TransactionDate: time.Now().UnixMilli(),
// 				Type:            "IN",
// 				Source:          "PURCHASE_DEBT_CANCELLED",
// 				Amount:          refund,
// 				Description:     fmt.Sprintf("Refund pembatalan sebagian %s", purchase.Code),
// 				ReferenceKey:    purchase.Code,
// 				BranchID:        &purchase.BranchID,
// 			}
// 			if err := c.CashBankTransactionRepository.Create(tx, &refundTx); err != nil {
// 				c.Log.WithError(err).Error("error insert refund transaction")
// 				return errors.New("internal server error")
// 			}
// 			debt.PaidAmount = debt.TotalAmount
// 		}

// 		if err := c.DebtRepository.Update(tx, debt); err != nil {
// 			c.Log.WithError(err).Error("error update debt")
// 			return errors.New("internal server error")
// 		}

// 		// kalau sudah 0 semua → VOID
// 		if debt.TotalAmount == 0 && debt.PaidAmount == 0 {
// 			if err := c.DebtRepository.UpdateStatus(tx, debt.ID); err != nil {
// 				c.Log.WithError(err).Error("error update debt status")
// 				return errors.New("internal server error")
// 			}
// 		}
// 	}

// 	// 9. Update purchase total_price
// 	newTotal := purchase.TotalPrice - cancelAmount
// 	if newTotal < 0 {
// 		newTotal = 0
// 	}
// 	if err := c.PurchaseRepository.UpdateTotalPrice(tx, purchase.Code, newTotal); err != nil {
// 		c.Log.WithError(err).Error("error updating purchase total_price")
// 		return errors.New("internal server error")
// 	}

// 	// 10. Jika semua detail sudah cancelled, update status purchase jadi CANCELLED
// 	remaining, err := c.PurchaseDetailRepository.CountActiveByPurchaseCode(tx, purchase.Code)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error counting remaining details")
// 		return errors.New("internal server error")
// 	}
// 	if remaining == 0 {
// 		if err := c.PurchaseRepository.Cancel(tx, purchase.Code); err != nil {
// 			c.Log.WithError(err).Error("error updating purchase status")
// 			return errors.New("internal server error")
// 		}
// 	}

// 	// 11. Commit
// 	if err := tx.Commit().Error; err != nil {
// 		c.Log.WithError(err).Error("error committing cancel purchase detail")
// 		return errors.New("internal server error")
// 	}

// 	return nil
// }

func (c *PurchaseDetailUseCase) Cancel(ctx context.Context, request *model.CancelPurchaseDetailRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// 1. Validasi input
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	// 2. Ambil purchase (lock)
	purchase := new(entity.Purchase)
	if err := c.PurchaseRepository.FindLockByCode(tx, request.PurchaseCode, purchase); err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return errors.New("not found")
	}

	// cek 24 jam
	createdTime := time.UnixMilli(purchase.CreatedAt)
	if time.Since(createdTime).Hours() >= 24 {
		c.Log.WithField("purchase_code", purchase.Code).
			Error("error cancelling purchase detail: exceeded 24-hour window")
		return errors.New("forbidden")
	}

	// 3. Ambil detail
	detail := new(entity.PurchaseDetail)
	if err := c.PurchaseDetailRepository.FindPriceBySizeIDAndPurchaseCode(tx, purchase.Code, request.SizeID, detail); err != nil {
		c.Log.WithError(err).Error("error getting purchase detail")
		return errors.New("not found")
	}
	if detail.IsCancelled {
		return errors.New("purchase detail already cancelled")
	}

	// 4. Update status cancelled
	if err := c.PurchaseDetailRepository.CancelBySizeID(tx, purchase.Code, request.SizeID); err != nil {
		c.Log.WithError(err).Error("error cancelling purchase detail")
		return errors.New("internal server error")
	}

	// 5. Update stock (branch_inventory)
	inv := new(entity.BranchInventory)
	if err := c.BranchInventoryRepository.FindByBranchIDAndSizeID(tx, inv, purchase.BranchID, detail.SizeID); err != nil {
		c.Log.WithError(err).Error("branch inventory not found")
		return errors.New("internal server error")
	}
	if inv.Stock < detail.Qty {
		return fmt.Errorf("insufficient stock for size_id %d", detail.SizeID)
	}
	if err := c.BranchInventoryRepository.BulkDecreaseStock(tx, purchase.BranchID, map[uint]int{detail.SizeID: detail.Qty}); err != nil {
		c.Log.WithError(err).Error("error decreasing stock")
		return errors.New("internal server error")
	}

	// 6. Insert inventory_movement
	mv := entity.InventoryMovement{
		BranchInventoryID: inv.ID,
		ChangeQty:         -detail.Qty,
		ReferenceType:     "PURCHASE_DETAIL_CANCELLED",
		ReferenceKey:      purchase.Code,
		CreatedAt:         time.Now().UnixMilli(),
	}
	if err := c.InventoryMovementRepository.Create(tx, &mv); err != nil {
		c.Log.WithError(err).Error("error inserting inventory movement")
		return errors.New("internal server error")
	}

	// 7. Update buy price kalau perlu
	lastPrices, err := c.PurchaseDetailRepository.FindLastBuyPricesBySizeIDs(tx, []uint{detail.SizeID}, purchase.Code)
	if err != nil {
		c.Log.WithError(err).Error("error finding last buy price")
		return errors.New("internal server error")
	}
	if lastPrice, ok := lastPrices[detail.SizeID]; ok {
		if err := c.SizeRepository.BulkUpdateBuyPrice(tx, map[uint]float64{detail.SizeID: lastPrice}); err != nil {
			c.Log.WithError(err).Error("error updating buy price")
			return errors.New("internal server error")
		}
	}

	// 8. Debt handling
	debt := new(entity.Debt)
	if err := c.DebtRepository.FindByPurchaseCodeOrInit(tx, debt, purchase.Code); err != nil {
		c.Log.WithError(err).Error("error getting debt")
		return errors.New("internal server error")
	}

	cancelAmount := detail.BuyPrice * float64(detail.Qty)

	if debt.ID != 0 {
		// Kurangi total_amount
		debt.TotalAmount -= cancelAmount
		if debt.TotalAmount < 0 {
			debt.TotalAmount = 0
		}

		// Refund kalau overpaid
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
				c.Log.WithError(err).Error("error insert refund transaction")
				return errors.New("internal server error")
			}
			debt.PaidAmount = debt.TotalAmount
		}

		// Update status debt
		if debt.TotalAmount == 0 && debt.PaidAmount == 0 {
			debt.Status = "VOID"
		} else if debt.PaidAmount >= debt.TotalAmount {
			debt.Status = "PAID"

			// Buat purchase_payment
			payment := &entity.PurchasePayment{
				PurchaseCode:  purchase.Code,
				PaymentMethod: "CASH",
				Amount:        debt.TotalAmount,
				CreatedAt:     time.Now().UnixMilli(),
			}
			if err := c.PurchasePaymentRepository.Create(tx, payment); err != nil {
				c.Log.WithError(err).Error("error creating purchase payment")
				return errors.New("internal server error")
			}
		} else {
			debt.Status = "PENDING"
		}

		if err := c.DebtRepository.Update(tx, debt); err != nil {
			c.Log.WithError(err).Error("error update debt")
			return errors.New("internal server error")
		}
	} else {
		// Ga ada debt → kurangi total_price + refund langsung
		purchase.TotalPrice -= cancelAmount
		if err := c.PurchaseRepository.UpdateTotalPrice(tx, purchase.Code, purchase.TotalPrice); err != nil {
			c.Log.WithError(err).Error("error update purchase total_price")
			return errors.New("internal server error")
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
			c.Log.WithError(err).Error("error insert refund transaction")
			return errors.New("internal server error")
		}
	}

	// 9. Update purchase total_price
	newTotal := purchase.TotalPrice - (detail.BuyPrice * float64(detail.Qty))
	if newTotal < 0 {
		newTotal = 0
	}
	if err := c.PurchaseRepository.UpdateTotalPrice(tx, purchase.Code, newTotal); err != nil {
		c.Log.WithError(err).Error("error updating purchase total_price")
		return errors.New("internal server error")
	}

	// 10. Jika semua detail sudah cancelled, update status purchase jadi CANCELLED
	remaining, err := c.PurchaseDetailRepository.CountActiveByPurchaseCode(tx, purchase.Code)
	if err != nil {
		c.Log.WithError(err).Error("error counting remaining details")
		return errors.New("internal server error")
	}
	if remaining == 0 {
		if err := c.PurchaseRepository.Cancel(tx, purchase.Code); err != nil {
			c.Log.WithError(err).Error("error updating purchase status")
			return errors.New("internal server error")
		}
	}

	// 11. Commit
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error committing cancel purchase detail")
		return errors.New("internal server error")
	}

	return nil
}

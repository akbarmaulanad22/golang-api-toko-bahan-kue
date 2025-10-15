package usecase

import (
	"context"
	"strconv"
	"time"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/model/converter"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DebtPaymentUseCase struct {
	DB                            *gorm.DB
	Log                           *logrus.Logger
	Validate                      *validator.Validate
	DebtPaymentRepository         *repository.DebtPaymentRepository
	CashBankTransactionRepository *repository.CashBankTransactionRepository
	DebtRepository                *repository.DebtRepository
	SaleRepository                *repository.SaleRepository
	PurchaseRepository            *repository.PurchaseRepository
}

func NewDebtPaymentUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	debtPaymentRepository *repository.DebtPaymentRepository,
	cashBankTransactionRepository *repository.CashBankTransactionRepository,
	debtRepository *repository.DebtRepository,
	saleRepository *repository.SaleRepository,
	purchaseRepository *repository.PurchaseRepository,
) *DebtPaymentUseCase {
	return &DebtPaymentUseCase{
		DB:                            db,
		Log:                           logger,
		Validate:                      validate,
		DebtPaymentRepository:         debtPaymentRepository,
		CashBankTransactionRepository: cashBankTransactionRepository,
		DebtRepository:                debtRepository,
		SaleRepository:                saleRepository,
		PurchaseRepository:            purchaseRepository,
	}
}

func (c *DebtPaymentUseCase) Create(ctx context.Context, request *model.CreateDebtPaymentRequest) (*model.DebtPaymentResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	debt := new(entity.Debt)
	if err := c.DebtRepository.FindById(tx, debt, request.DebtID); err != nil {
		c.Log.WithError(err).Error("error getting debt")
		return nil, helper.GetNotFoundMessage("debt payment", err)
	}

	if debt.Status == "VOID" {
		c.Log.WithField("debt_id", request.DebtID).Error("debt payment already cancelled")
		return nil, model.NewAppErr("conflict", "debt payment already cancelled")
	}

	// Tambah paid amount
	debt.PaidAmount += request.Amount

	// Jika sudah lunas, update status jadi PAID
	if debt.PaidAmount >= debt.TotalAmount {
		debt.Status = "PAID"
	}

	if err := c.DebtRepository.Update(tx, debt); err != nil {
		c.Log.WithError(err).Error("error update paid amount debt")
		return nil, model.NewAppErr("internal server error", nil)
	}

	debtPayment := &entity.DebtPayment{
		Note:        request.Note,
		Amount:      request.Amount,
		PaymentDate: request.PaymentDate,
		DebtID:      request.DebtID,
	}

	if err := c.DebtPaymentRepository.Create(tx, debtPayment); err != nil {

		c.Log.WithError(err).Error("error creating debt payment")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1452 {
			return nil, model.NewAppErr("referenced resource not found", "the specified debt does not exist.")
		}

		return nil, model.NewAppErr("internal server error", nil)
	}

	var txType string
	if debt.ReferenceType == "SALE" {
		txType = "IN"
	} else {
		txType = "OUT"
	}

	cashBankTransaction := &entity.CashBankTransaction{
		TransactionDate: request.PaymentDate,
		Type:            txType,
		Source:          "DEBT",
		Amount:          request.Amount,
		Description:     request.Note,
		ReferenceKey:    strconv.Itoa(int(request.DebtID)),
		BranchID:        request.BranchID,
	}

	if err := c.CashBankTransactionRepository.Create(tx, cashBankTransaction); err != nil {

		c.Log.WithError(err).Error("error creating cash bank transaction")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1452 {
			return nil, model.NewAppErr("referenced resource not found", "the specified branch does not exist.")

		}
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating debt payment")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.DebtPaymentToResponse(debtPayment), nil
}

func (c *DebtPaymentUseCase) Delete(ctx context.Context, request *model.DeleteDebtPaymentRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	debtPayment := new(entity.DebtPayment)
	if err := c.DebtPaymentRepository.FindById(tx, debtPayment, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting debt payment")
		return helper.GetNotFoundMessage("debt payments", err)
	}

	debt := new(entity.Debt)
	if err := c.DebtRepository.FindById(tx, debt, debtPayment.DebtID); err != nil {
		c.Log.WithError(err).Error("error getting debt")
		return helper.GetNotFoundMessage("debt", err)
	}

	var branchID *uint
	var txType string

	switch debt.ReferenceType {
	case "SALE":
		sale, err := c.SaleRepository.FindByCode(tx, debt.ReferenceCode)
		if err != nil {
			c.Log.WithError(err).Error("error getting sale for branch id")
			return helper.GetNotFoundMessage("sale", err)
		}
		branchID = &sale.BranchID
		txType = "OUT"

	case "PURCHASE":
		purchase, err := c.PurchaseRepository.FindByCode(tx, debt.ReferenceCode)
		if err != nil {
			c.Log.WithError(err).Error("error getting purchase for branch id")
			return helper.GetNotFoundMessage("purchase", err)
		}
		branchID = &purchase.BranchID
		txType = "IN"
	}

	createdAtTime := time.UnixMilli(debtPayment.CreatedAt)
	if time.Since(createdAtTime) > time.Hour {
		c.Log.WithField("debt_id", debt.ID).Error("debt cannot be deleted after 1 hours")
		return model.NewAppErr("forbidden", "debt cannot be deleted after 1 hours")
	}

	newPaidAmount := debt.PaidAmount - debtPayment.Amount
	if newPaidAmount < 0 {
		newPaidAmount = 0
	}

	debt.PaidAmount = newPaidAmount

	if debt.PaidAmount == 0 {
		debt.Status = "PENDING"
	}

	if err := c.DebtRepository.Update(tx, debt); err != nil {
		c.Log.WithError(err).Error("error updating debt")
		return model.NewAppErr("internal server error", nil)
	}

	cashBankTransaction := &entity.CashBankTransaction{
		TransactionDate: time.Now().UnixMilli(),
		Type:            txType,
		Source:          "DEBT",
		Amount:          debtPayment.Amount,
		Description:     "PEMBATALAN BIAYA UTANG",
		ReferenceKey:    strconv.Itoa(int(debt.ID)),
		BranchID:        branchID,
		CreatedAt:       time.Now().UnixMilli(),
		UpdatedAt:       time.Now().UnixMilli(),
	}

	if err := c.CashBankTransactionRepository.Create(tx, cashBankTransaction); err != nil {
		c.Log.WithError(err).Error("error creating cash bank transaction")
		return model.NewAppErr("internal server error", nil)
	}

	if err := c.DebtPaymentRepository.Delete(tx, debtPayment); err != nil {
		c.Log.WithError(err).Error("error deleting debt payment")
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error delete debt payment")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

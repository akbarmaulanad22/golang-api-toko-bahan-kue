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

type ExpenseUseCase struct {
	DB                            *gorm.DB
	Log                           *logrus.Logger
	Validate                      *validator.Validate
	ExpenseRepository             *repository.ExpenseRepository
	CashBankTransactionRepository *repository.CashBankTransactionRepository
}

func NewExpenseUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	expenseRepository *repository.ExpenseRepository, cashBankTransactionRepository *repository.CashBankTransactionRepository) *ExpenseUseCase {
	return &ExpenseUseCase{
		DB:                            db,
		Log:                           logger,
		Validate:                      validate,
		ExpenseRepository:             expenseRepository,
		CashBankTransactionRepository: cashBankTransactionRepository,
	}
}

func (c *ExpenseUseCase) Create(ctx context.Context, request *model.CreateExpenseRequest) (*model.ExpenseResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	expense := &entity.Expense{
		Description: request.Description,
		Amount:      request.Amount,
		BranchID:    request.BranchID,
	}

	if err := c.ExpenseRepository.Create(tx, expense); err != nil {
		c.Log.WithError(err).Error("error creating expense")

		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1452 {
			return nil, model.NewAppErr("referenced resource not found", "the specified branch does not exist.")
		}

		return nil, model.NewAppErr("internal server error", nil)
	}

	cashBankTransaction := entity.CashBankTransaction{
		TransactionDate: time.Now().UnixMilli(),
		Type:            "IN",
		Source:          "EXPENSE",
		Amount:          request.Amount,
		ReferenceKey:    strconv.Itoa(int(expense.ID)),
		BranchID:        request.BranchID,
	}

	if err := c.CashBankTransactionRepository.Create(tx, &cashBankTransaction); err != nil {
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating expense")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.ExpenseToResponse(expense), nil
}

func (c *ExpenseUseCase) Update(ctx context.Context, request *model.UpdateExpenseRequest) (*model.ExpenseResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	expense := new(entity.Expense)
	if err := c.ExpenseRepository.FindById(tx, expense, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting expense")
		return nil, helper.GetNotFoundMessage("expense", err)
	}

	if expense.Description == request.Description && expense.Amount == request.Amount {
		return converter.ExpenseToResponse(expense), nil
	}

	expense.Description = request.Description
	expense.Amount = request.Amount

	if err := c.ExpenseRepository.Update(tx, expense); err != nil {
		c.Log.WithError(err).Error("error updating expense")
		return nil, model.NewAppErr("internal server error", nil)
	}

	cashBankTransaction := new(entity.CashBankTransaction)
	if err := c.CashBankTransactionRepository.FindByKeyAndSource(tx, cashBankTransaction, request.ID, "EXPENSE"); err != nil {
		c.Log.WithError(err).Error("error getting cash bank transaction")
		return nil, helper.GetNotFoundMessage("cash bank transaction", err)
	}

	cashBankTransaction.Amount = expense.Amount
	cashBankTransaction.Description = expense.Description

	if err := c.CashBankTransactionRepository.Update(tx, cashBankTransaction); err != nil {
		c.Log.WithError(err).Error("error updating cash bank transaction")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating expense")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.ExpenseToResponse(expense), nil
}

func (c *ExpenseUseCase) Delete(ctx context.Context, request *model.DeleteExpenseRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	expense := new(entity.Expense)
	if err := c.ExpenseRepository.FindById(tx, expense, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting expense")
		return helper.GetNotFoundMessage("expense", err)
	}

	cashBankTransaction := new(entity.CashBankTransaction)
	if err := c.CashBankTransactionRepository.FindByKeyAndSource(tx, cashBankTransaction, request.ID, "EXPENSE"); err != nil {
		c.Log.WithError(err).Error("error getting cash bank transaction")
		return helper.GetNotFoundMessage("cash bank transaction", err)
	}

	if err := c.CashBankTransactionRepository.Delete(tx, cashBankTransaction); err != nil {
		c.Log.WithError(err).Error("error deleting cash bank transaction")
		return model.NewAppErr("internal server error", nil)
	}

	if err := c.ExpenseRepository.Delete(tx, expense); err != nil {
		c.Log.WithError(err).Error("error deleting expense")
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting expense")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

func (c *ExpenseUseCase) Search(ctx context.Context, request *model.SearchExpenseRequest) ([]model.ExpenseResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	expenses, total, err := c.ExpenseRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting expenses")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting expenses")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	responses := make([]model.ExpenseResponse, len(expenses))
	for i, expense := range expenses {
		responses[i] = *converter.ExpenseToResponse(&expense)
	}

	return responses, total, nil
}

func (c *ExpenseUseCase) ConsolidateReport(ctx context.Context, request *model.SearchConsolidateExpenseRequest) (*model.ConsolidatedExpenseResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	expenses, err := c.ExpenseRepository.ConsolidateReport(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting expenses consolidated report")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting expenses consolidated report")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return expenses, nil
}

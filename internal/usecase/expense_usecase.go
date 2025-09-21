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
		return nil, errors.New("bad request")
	}

	expense := &entity.Expense{
		Description: request.Description,
		Amount:      request.Amount,
		BranchID:    request.BranchID,
	}

	if err := c.ExpenseRepository.Create(tx, expense); err != nil {
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

		c.Log.WithError(err).Error("error creating expense")
		return nil, errors.New("internal server error")
	}

	// masukin ke catatan keuangan
	cashBankTransaction := entity.CashBankTransaction{
		TransactionDate: time.Now().UnixMilli(),
		Type:            "IN",
		Source:          "EXPENSE",
		Amount:          request.Amount,
		ReferenceKey:    strconv.Itoa(int(expense.ID)),
		BranchID:        request.BranchID,
	}

	if err := c.CashBankTransactionRepository.Create(tx, &cashBankTransaction); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating expense")
		return nil, errors.New("internal server error")
	}

	return converter.ExpenseToResponse(expense), nil
}

func (c *ExpenseUseCase) Update(ctx context.Context, request *model.UpdateExpenseRequest) (*model.ExpenseResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	expense := new(entity.Expense)
	if err := c.ExpenseRepository.FindById(tx, expense, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting expense")
		return nil, errors.New("not found")
	}

	if expense.Description == request.Description && expense.Amount == request.Amount {
		return converter.ExpenseToResponse(expense), nil
	}

	expense.Description = request.Description
	expense.Amount = request.Amount

	if err := c.ExpenseRepository.Update(tx, expense); err != nil {
		c.Log.WithError(err).Error("error updating expense")
		return nil, errors.New("internal server error")
	}

	cashBankTransaction := new(entity.CashBankTransaction)
	if err := c.CashBankTransactionRepository.FindByKeyAndSource(tx, cashBankTransaction, request.ID, "EXPENSE"); err != nil {
		c.Log.WithError(err).Error("error getting cash bank transaction")
		return nil, errors.New("not found")
	}

	cashBankTransaction.Amount = expense.Amount

	if err := c.CashBankTransactionRepository.Update(tx, cashBankTransaction); err != nil {
		c.Log.WithError(err).Error("error updating cash bank transaction")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating expense")
		return nil, errors.New("internal server error")
	}

	return converter.ExpenseToResponse(expense), nil
}

func (c *ExpenseUseCase) Delete(ctx context.Context, request *model.DeleteExpenseRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	expense := new(entity.Expense)
	if err := c.ExpenseRepository.FindById(tx, expense, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting expense")
		return errors.New("not found")
	}

	cashBankTransaction := new(entity.CashBankTransaction)
	if err := c.CashBankTransactionRepository.FindByKeyAndSource(tx, cashBankTransaction, request.ID, "EXPENSE"); err != nil {
		c.Log.WithError(err).Error("error getting cash bank transaction")
		return errors.New("not found")
	}

	if err := c.CashBankTransactionRepository.Delete(tx, cashBankTransaction); err != nil {
		c.Log.WithError(err).Error("error deleting cash bank transaction")
		return errors.New("internal server error")
	}

	if err := c.ExpenseRepository.Delete(tx, expense); err != nil {
		c.Log.WithError(err).Error("error deleting expense")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting expense")
		return errors.New("internal server error")
	}

	return nil
}

func (c *ExpenseUseCase) Search(ctx context.Context, request *model.SearchExpenseRequest) ([]model.ExpenseResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	expenses, total, err := c.ExpenseRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting expenses")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting expenses")
		return nil, 0, errors.New("internal server error")
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
		return nil, errors.New("bad request")
	}

	expenses, err := c.ExpenseRepository.ConsolidateReport(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting expenses consolidated report")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting expenses consolidated report")
		return nil, errors.New("internal server error")
	}

	return expenses, nil
}

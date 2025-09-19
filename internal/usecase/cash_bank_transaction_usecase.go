package usecase

import (
	"context"
	"errors"
	"strings"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
	"tokobahankue/internal/model/converter"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CashBankTransactionUseCase struct {
	DB                            *gorm.DB
	Log                           *logrus.Logger
	Validate                      *validator.Validate
	CashBankTransactionRepository *repository.CashBankTransactionRepository
}

func NewCashBankTransactionUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	cashBankTransactionRepository *repository.CashBankTransactionRepository) *CashBankTransactionUseCase {
	return &CashBankTransactionUseCase{
		DB:                            db,
		Log:                           logger,
		Validate:                      validate,
		CashBankTransactionRepository: cashBankTransactionRepository,
	}
}

func (c *CashBankTransactionUseCase) Create(ctx context.Context, request *model.CreateCashBankTransactionRequest) (*model.CashBankTransactionResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	cashBankTransaction := &entity.CashBankTransaction{
		TransactionDate: request.TransactionDate,
		Type:            request.Type,
		Source:          request.Source,
		Amount:          request.Amount,
		Description:     request.Description,
		ReferenceKey:    request.ReferenceKey,
		BranchID:        request.BranchID,
	}

	if err := c.CashBankTransactionRepository.Create(tx, cashBankTransaction); err != nil {
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

		c.Log.WithError(err).Error("error creating cash bank transaction")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating cash bank transaction")
		return nil, errors.New("internal server error")
	}

	return converter.CashBankTransactionToResponse(cashBankTransaction), nil
}

func (c *CashBankTransactionUseCase) Update(ctx context.Context, request *model.UpdateCashBankTransactionRequest) (*model.CashBankTransactionResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	cashBankTransaction := new(entity.CashBankTransaction)
	if err := c.CashBankTransactionRepository.FindById(tx, cashBankTransaction, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting cash bank transaction")
		return nil, errors.New("not found")
	}

	cashBankTransaction.TransactionDate = request.TransactionDate
	cashBankTransaction.Type = request.Type
	cashBankTransaction.Source = request.Source
	cashBankTransaction.Amount = request.Amount
	cashBankTransaction.Description = request.Description
	cashBankTransaction.ReferenceKey = request.ReferenceKey
	cashBankTransaction.Description = request.Description

	if err := c.CashBankTransactionRepository.Update(tx, cashBankTransaction); err != nil {
		c.Log.WithError(err).Error("error updating cash bank transaction")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating cash bank transaction")
		return nil, errors.New("internal server error")
	}

	return converter.CashBankTransactionToResponse(cashBankTransaction), nil
}

func (c *CashBankTransactionUseCase) Get(ctx context.Context, request *model.GetCashBankTransactionRequest) (*model.CashBankTransactionResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	cashBankTransaction := new(entity.CashBankTransaction)
	if err := c.CashBankTransactionRepository.FindById(tx, cashBankTransaction, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting cash bank transaction")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting cash bank transaction")
		return nil, errors.New("internal server error")
	}

	return converter.CashBankTransactionToResponse(cashBankTransaction), nil
}

func (c *CashBankTransactionUseCase) Delete(ctx context.Context, request *model.DeleteCashBankTransactionRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	cashBankTransaction := new(entity.CashBankTransaction)
	if err := c.CashBankTransactionRepository.FindById(tx, cashBankTransaction, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting cash bank transaction")
		return errors.New("not found")
	}

	if err := c.CashBankTransactionRepository.Delete(tx, cashBankTransaction); err != nil {
		c.Log.WithError(err).Error("error deleting cash bank transaction")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting cash bank transaction")
		return errors.New("internal server error")
	}

	return nil
}

func (c *CashBankTransactionUseCase) Search(ctx context.Context, request *model.SearchCashBankTransactionRequest) ([]model.CashBankTransactionResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	cashBankTransactions, total, err := c.CashBankTransactionRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting cash bank transactions")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting cash bank transactions")
		return nil, 0, errors.New("internal server error")
	}

	responses := make([]model.CashBankTransactionResponse, len(cashBankTransactions))
	for i, cashBankTransaction := range cashBankTransactions {
		responses[i] = *converter.CashBankTransactionToResponse(&cashBankTransaction)
	}

	return responses, total, nil
}

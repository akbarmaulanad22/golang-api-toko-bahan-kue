package usecase

import (
	"context"
	"errors"
	"tokobahankue/internal/model"
	"tokobahankue/internal/model/converter"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
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

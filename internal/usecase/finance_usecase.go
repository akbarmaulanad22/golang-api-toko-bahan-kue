package usecase

import (
	"context"
	"errors"
	"tokobahankue/internal/model"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type FinanceUseCase struct {
	DB                *gorm.DB
	Log               *logrus.Logger
	Validate          *validator.Validate
	FinanceRepository *repository.FinanceRepository
}

func NewFinanceUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	financeRepository *repository.FinanceRepository) *FinanceUseCase {
	return &FinanceUseCase{
		DB:                db,
		Log:               logger,
		Validate:          validate,
		FinanceRepository: financeRepository,
	}
}

func (c *FinanceUseCase) GetOwnerSummary(ctx context.Context, request *model.SearchFinanceSummaryOwnerRequest) (*model.FinanceSummaryOwnerResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	dailyFinances, err := c.FinanceRepository.GetOwnerSummary(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting daily finances")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting daily finances")
		return nil, errors.New("internal server error")
	}

	return dailyFinances, nil
}

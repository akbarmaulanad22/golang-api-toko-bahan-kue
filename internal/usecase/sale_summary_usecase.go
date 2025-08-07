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

type SaleSummaryUseCase struct {
	DB                    *gorm.DB
	Log                   *logrus.Logger
	Validate              *validator.Validate
	SaleSummaryRepository *repository.SaleSummaryRepository
}

func NewSaleSummaryUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	saleSummaryRepository *repository.SaleSummaryRepository) *SaleSummaryUseCase {
	return &SaleSummaryUseCase{
		DB:                    db,
		Log:                   logger,
		Validate:              validate,
		SaleSummaryRepository: saleSummaryRepository,
	}
}

func (c *SaleSummaryUseCase) BranchSalesSummary(ctx context.Context) ([]model.BranchSalesSummaryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	saleSummarys, err := c.SaleSummaryRepository.BranchSalesSummary(tx)
	if err != nil {
		c.Log.WithError(err).Error("error getting sale summary")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting sale summary")
		return nil, errors.New("internal server error")
	}

	return saleSummarys, nil
}

func (c *SaleSummaryUseCase) DailySalesSummaryByBranchID(ctx context.Context, request *model.ListDailySalesSummaryRequest) ([]model.DailySalesSummaryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	saleSummarys, err := c.SaleSummaryRepository.DailySalesSummaryByBranchID(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting daily sale summary")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting daily sale summary")
		return nil, errors.New("internal server error")
	}

	return saleSummarys, nil
}

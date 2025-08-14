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

type FinancialReportUseCase struct {
	DB                        *gorm.DB
	Log                       *logrus.Logger
	Validate                  *validator.Validate
	FinancialReportRepository *repository.FinancialReportRepository
}

func NewFinancialReportUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	financialReportRepository *repository.FinancialReportRepository) *FinancialReportUseCase {
	return &FinancialReportUseCase{
		DB:                        db,
		Log:                       logger,
		Validate:                  validate,
		FinancialReportRepository: financialReportRepository,
	}
}

func (c *FinancialReportUseCase) SearchDailyReports(ctx context.Context, request *model.SearchDailyFinancialReportRequest) ([]model.DailyFinancialReportResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	dailyFinancialReports, err := c.FinancialReportRepository.SearchDailyFinancialReport(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting daily financial reports")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting daily financial reports")
		return nil, errors.New("internal server error")
	}

	return dailyFinancialReports, nil
}

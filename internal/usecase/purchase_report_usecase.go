package usecase

import (
	"context"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PurchaseReportUseCase struct {
	DB                       *gorm.DB
	Log                      *logrus.Logger
	Validate                 *validator.Validate
	PurchaseReportRepository *repository.PurchaseReportRepository
}

func NewPurchaseReportUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	saleReportRepository *repository.PurchaseReportRepository) *PurchaseReportUseCase {
	return &PurchaseReportUseCase{
		DB:                       db,
		Log:                      logger,
		Validate:                 validate,
		PurchaseReportRepository: saleReportRepository,
	}
}

func (c *PurchaseReportUseCase) SearchDaily(ctx context.Context, request *model.SearchPurchasesReportRequest) ([]model.PurchasesDailyReportResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	purchasesReports, total, err := c.PurchaseReportRepository.SearchDaily(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting daily purchases reports")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting daily purchases reports")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	return purchasesReports, total, nil
}

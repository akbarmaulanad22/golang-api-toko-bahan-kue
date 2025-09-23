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
		return nil, 0, errors.New("bad request")
	}

	purchasesReports, total, err := c.PurchaseReportRepository.SearchDaily(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting daily purchases reports")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting daily purchases reports")
		return nil, 0, errors.New("internal server error")
	}

	return purchasesReports, total, nil
}

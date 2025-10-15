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

type SaleReportUseCase struct {
	DB                   *gorm.DB
	Log                  *logrus.Logger
	Validate             *validator.Validate
	SaleReportRepository *repository.SaleReportRepository
}

func NewSaleReportUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	saleReportRepository *repository.SaleReportRepository) *SaleReportUseCase {
	return &SaleReportUseCase{
		DB:                   db,
		Log:                  logger,
		Validate:             validate,
		SaleReportRepository: saleReportRepository,
	}
}

func (c *SaleReportUseCase) SearchDaily(ctx context.Context, request *model.SearchSalesReportRequest) ([]model.SalesDailyReportResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	salesReports, total, err := c.SaleReportRepository.SearchDaily(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting daily sales reports")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting daily sales reports")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	return salesReports, total, nil
}

func (c *SaleReportUseCase) SearchTopSellerProduct(ctx context.Context, request *model.SearchSalesReportRequest) ([]model.SalesTopSellerReportResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	salesReports, total, err := c.SaleReportRepository.SearchTopSellerProduct(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting top seller product sales reports")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting top seller product sales reports")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	return salesReports, total, nil
}

func (c *SaleReportUseCase) SearchTopSellerCategory(ctx context.Context, request *model.SearchSalesReportRequest) ([]model.SalesCategoryResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	salesReports, total, err := c.SaleReportRepository.SearchTopSellerCategory(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting top seller category sales reports")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting top seller category sales reports")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	return salesReports, total, nil
}

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
		return nil, 0, errors.New("bad request")
	}

	salesReports, total, err := c.SaleReportRepository.SearchDaily(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting daily sales reports")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting daily sales reports")
		return nil, 0, errors.New("internal server error")
	}

	return salesReports, total, nil
}

func (c *SaleReportUseCase) SearchTopSeller(ctx context.Context, request *model.SearchSalesReportRequest) ([]model.SalesTopSellerReportResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	salesReports, total, err := c.SaleReportRepository.SearchTopSeller(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting top seller sales reports")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting top seller sales reports")
		return nil, 0, errors.New("internal server error")
	}

	return salesReports, total, nil
}

func (c *SaleReportUseCase) SearchCategory(ctx context.Context, request *model.SearchSalesReportRequest) ([]model.SalesCategoryResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	salesReports, total, err := c.SaleReportRepository.SearchCategory(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting top seller per category sales reports")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting top seller per category sales reports")
		return nil, 0, errors.New("internal server error")
	}

	return salesReports, total, nil
}

// func (c *SaleUseCase) GetBranchSalesReport(ctx context.Context) ([]model.BranchSalesReportResponse, error) {
// 	tx := c.DB.WithContext(ctx).Begin()
// 	defer tx.Rollback()

// 	salesReports, err := c.SaleRepository.SummaryAllBranch(tx)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error getting branch sales report")
// 		return nil, errors.New("internal server error")
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		c.Log.WithError(err).Error("error getting branch sales report")
// 		return nil, errors.New("internal server error")
// 	}

// 	return salesReports, nil
// }

// func (c *SaleUseCase) ListBestSellingProductByBranchID(ctx context.Context, request *model.ListBestSellingProductRequest) ([]model.BestSellingProductResponse, error) {
// 	tx := c.DB.WithContext(ctx).Begin()
// 	defer tx.Rollback()
// 	bestSellingProducts, err := c.SaleRepository.FindhBestSellingProductsByBranchID(tx, request)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error getting best selling products")
// 		return nil, errors.New("internal server error")
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		c.Log.WithError(err).Error("error getting best selling products")
// 		return nil, errors.New("internal server error")
// 	}

// 	return bestSellingProducts, nil
// }

// func (c *SaleUseCase) ListBestSellingProductGlobal(ctx context.Context) ([]model.BestSellingProductResponse, error) {
// 	tx := c.DB.WithContext(ctx).Begin()
// 	defer tx.Rollback()
// 	bestSellingProducts, err := c.SaleRepository.FindhBestSellingProductsGlobal(tx)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error getting best selling products")
// 		return nil, errors.New("internal server error")
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		c.Log.WithError(err).Error("error getting best selling products")
// 		return nil, errors.New("internal server error")
// 	}

// 	return bestSellingProducts, nil
// }

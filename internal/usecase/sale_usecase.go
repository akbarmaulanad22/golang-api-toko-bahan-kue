package usecase

import (
	"context"
	"errors"
	"strings"
	"time"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
	"tokobahankue/internal/model/converter"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SaleUseCase struct {
	DB             *gorm.DB
	Log            *logrus.Logger
	Validate       *validator.Validate
	SaleRepository *repository.SaleRepository
}

func NewSaleUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	saleRepository *repository.SaleRepository) *SaleUseCase {
	return &SaleUseCase{
		DB:             db,
		Log:            logger,
		Validate:       validate,
		SaleRepository: saleRepository,
	}
}

func (c *SaleUseCase) Create(ctx context.Context, request *model.CreateSaleRequest) (*model.SaleResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	sale := &entity.Sale{
		Code:         uuid.New().String(),
		CustomerName: request.CustomerName,
		Status:       request.Status,
		CashValue:    request.CashValue,
		DebitValue:   request.DebitValue,
		// PaidAt:        0,
		BranchID: request.BranchID,
	}

	for _, d := range request.SaleDetails {
		sale.SaleDetails = append(sale.SaleDetails, entity.SaleDetail{
			SaleCode:    sale.Code,
			SizeID:      d.SizeID,
			Qty:         d.Qty,
			IsCancelled: false,
		})
	}

	if err := c.SaleRepository.Create(tx, sale); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Tangani duplikat
			switch {
			case strings.Contains(mysqlErr.Message, "for key 'sales.PRIMARY'"): // sku
				c.Log.Warn("Code already exists")
				return nil, errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return nil, errors.New("conflict")
			}
		}

		c.Log.WithError(err).Error("error creating sale")
		return nil, errors.New("internal server error")
	}

	if err := c.SaleRepository.FindByCode(tx, sale, sale.Code); err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating sale")
		return nil, errors.New("internal server error")
	}

	return converter.SaleToResponse(sale), nil
}

func (c *SaleUseCase) Update(ctx context.Context, request *model.UpdateSaleRequest) (*model.SaleResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	sale := new(entity.Sale)
	if err := c.SaleRepository.FindByCode(tx, sale, request.Code); err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return nil, errors.New("not found")
	}

	if sale.Status == request.Status {
		return converter.SaleToResponse(sale), nil
	}

	createdTime := time.UnixMilli(sale.CreatedAt)
	now := time.Now()

	// Hitung durasi sejak dibuat
	duration := now.Sub(createdTime)

	// Jika status BUKAN PENDING dan sudah lewat 24 jam => tolak
	if sale.Status != model.PENDING && duration.Hours() >= 24 {
		c.Log.WithField("sale_code", sale.Code).Error("error updating sale: exceeded 24-hour window")
		return nil, errors.New("forbidden")
	}

	// Lanjut update status
	sale.Status = request.Status
	if err := c.SaleRepository.Update(tx, sale); err != nil {
		c.Log.WithError(err).Error("error updating sale")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating sale")
		return nil, errors.New("internal server error")
	}

	return converter.SaleToResponse(sale), nil
}

func (c *SaleUseCase) Get(ctx context.Context, request *model.GetSaleRequest) (*model.SaleResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	sale := new(entity.Sale)
	if err := c.SaleRepository.FindByCode(tx, sale, request.Code); err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return nil, errors.New("internal server error")
	}

	return converter.SaleToResponse(sale), nil
}

func (c *SaleUseCase) Search(ctx context.Context, request *model.SearchSaleRequest) ([]model.SaleResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	sales, total, err := c.SaleRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting sales")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting sales")
		return nil, 0, errors.New("internal server error")
	}

	responses := make([]model.SaleResponse, len(sales))
	for i, sale := range sales {
		responses[i] = *converter.SaleToResponse(&sale)
	}

	return responses, total, nil
}

func (c *SaleUseCase) SearchReports(ctx context.Context, request *model.SearchSaleReportRequest) ([]model.SaleReportResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	salesReports, total, err := c.SaleRepository.SearchReports(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting sales reports")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting sales reports")
		return nil, 0, errors.New("internal server error")
	}

	return salesReports, total, nil
}

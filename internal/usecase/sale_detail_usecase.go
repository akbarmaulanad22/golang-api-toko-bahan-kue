package usecase

import (
	"context"
	"errors"
	"time"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SaleDetailUseCase struct {
	DB                   *gorm.DB
	Log                  *logrus.Logger
	Validate             *validator.Validate
	SaleDetailRepository *repository.SaleDetailRepository
	SaleRepository       *repository.SaleRepository
}

func NewSaleDetailUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	saleDetailRepository *repository.SaleDetailRepository,
	saleRepository *repository.SaleRepository,
) *SaleDetailUseCase {
	return &SaleDetailUseCase{
		DB:                   db,
		Log:                  logger,
		Validate:             validate,
		SaleDetailRepository: saleDetailRepository,
		SaleRepository:       saleRepository,
	}
}

func (c *SaleDetailUseCase) Cancel(ctx context.Context, request *model.CancelSaleDetailRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	sale, err := c.SaleRepository.FindByCode(tx, request.SaleCode)
	if err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return errors.New("not found")
	}

	createdTime := time.UnixMilli(sale.CreatedAt)
	now := time.Now()

	// Hitung durasi sejak dibuat
	duration := now.Sub(createdTime)

	// Jika status BUKAN PENDING dan sudah lewat 24 jam => tolak
	if duration.Hours() >= 24 {
		c.Log.WithField("sale_code", sale.Code).Error("error updating sale: exceeded 24-hour window")
		return errors.New("forbidden")
	}

	// ambil price dan qty detail
	detail := new(entity.SaleDetail)
	if err := c.SaleDetailRepository.FindPriceBySizeIDAndSaleCode(tx, sale.Code, request.SizeID, detail); err != nil {
		c.Log.WithError(err).Error("error updating sale detail")
		return errors.New("internal server error")
	}

	if detail.IsCancelled {
		c.Log.WithError(err).Error("error updating sale detail")
		return errors.New("forbidden")
	}

	// Lanjut update status
	if err := c.SaleDetailRepository.Cancel(tx, sale.Code, request.SizeID); err != nil {
		c.Log.WithError(err).Error("error updating sale detail")
		return errors.New("internal server error")
	}

	// penyesuaian total price sale dengan data detail
	sale.TotalPrice -= (detail.SellPrice * float64(detail.Qty))

	if err := c.SaleRepository.UpdateTotalPrice(tx, sale.Code, sale.TotalPrice); err != nil {
		c.Log.WithError(err).Error("error updating sale")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating sale detail")
		return errors.New("internal server error")
	}

	return nil
}

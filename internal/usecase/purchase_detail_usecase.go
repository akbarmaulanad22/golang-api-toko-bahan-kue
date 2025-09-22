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

type PurchaseDetailUseCase struct {
	DB                       *gorm.DB
	Log                      *logrus.Logger
	Validate                 *validator.Validate
	PurchaseDetailRepository *repository.PurchaseDetailRepository
	PurchaseRepository       *repository.PurchaseRepository
}

func NewPurchaseDetailUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	purchaseDetailRepository *repository.PurchaseDetailRepository,
	purchaseRepository *repository.PurchaseRepository,
) *PurchaseDetailUseCase {
	return &PurchaseDetailUseCase{
		DB:                       db,
		Log:                      logger,
		Validate:                 validate,
		PurchaseDetailRepository: purchaseDetailRepository,
		PurchaseRepository:       purchaseRepository,
	}
}

func (c *PurchaseDetailUseCase) Cancel(ctx context.Context, request *model.CancelPurchaseDetailRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	purchase, err := c.PurchaseRepository.FindByCode(tx, request.PurchaseCode)
	if err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return errors.New("not found")
	}

	createdTime := time.UnixMilli(purchase.CreatedAt)
	now := time.Now()

	// Hitung durasi sejak dibuat
	duration := now.Sub(createdTime)

	// Jika status BUKAN PENDING dan sudah lewat 24 jam => tolak
	if duration.Hours() >= 24 {
		c.Log.WithField("purchase_code", purchase.Code).Error("error updating purchase: exceeded 24-hour window")
		return errors.New("forbidden")
	}

	// ambil price dan qty detail
	detail := new(entity.PurchaseDetail)
	if err := c.PurchaseDetailRepository.FindPriceBySizeIDAndPurchaseCode(tx, purchase.Code, request.SizeID, detail); err != nil {
		c.Log.WithError(err).Error("error updating purchase detail")
		return errors.New("internal server error")
	}

	if detail.IsCancelled {
		c.Log.WithError(err).Error("error updating purchase detail")
		return errors.New("forbidden")
	}

	// Lanjut update status
	if err := c.PurchaseDetailRepository.Cancel(tx, purchase.Code, request.SizeID); err != nil {
		c.Log.WithError(err).Error("error updating purchase detail")
		return errors.New("internal server error")
	}

	// penyesuaian total price purchase dengan data detail
	purchase.TotalPrice -= (detail.BuyPrice * float64(detail.Qty))

	if err := c.PurchaseRepository.UpdateTotalPrice(tx, purchase.Code, purchase.TotalPrice); err != nil {
		c.Log.WithError(err).Error("error updating purchase")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating purchase detail")
		return errors.New("internal server error")
	}

	return nil
}

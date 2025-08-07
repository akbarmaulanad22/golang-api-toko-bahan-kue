package usecase

import (
	"context"
	"errors"
	"strings"
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

type PurchaseUseCase struct {
	DB                 *gorm.DB
	Log                *logrus.Logger
	Validate           *validator.Validate
	PurchaseRepository *repository.PurchaseRepository
}

func NewPurchaseUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	purchaseRepository *repository.PurchaseRepository) *PurchaseUseCase {
	return &PurchaseUseCase{
		DB:                 db,
		Log:                logger,
		Validate:           validate,
		PurchaseRepository: purchaseRepository,
	}
}

func (c *PurchaseUseCase) Create(ctx context.Context, request *model.CreatePurchaseRequest) (*model.PurchaseResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	purchase := &entity.Purchase{
		Code: uuid.New().String(),

		SalesName:  request.SalesName,
		Status:     request.Status,
		CashValue:  request.CashValue,
		DebitValue: request.DebitValue,
		// PaidAt:        0,
		BranchID:      request.BranchID,
		DistributorID: request.DistributorID,
	}

	for _, d := range request.PurchaseDetails {
		purchase.PurchaseDetails = append(purchase.PurchaseDetails, entity.PurchaseDetail{
			PurchaseCode: purchase.Code,
			SizeID:       d.SizeID,
			Qty:          d.Qty,
			IsCancelled:  false,
		})
	}

	if err := c.PurchaseRepository.Create(tx, purchase); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Tangani duplikat
			switch {
			case strings.Contains(mysqlErr.Message, "for key 'purchases.PRIMARY'"): // sku
				c.Log.Warn("Code already exists")
				return nil, errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return nil, errors.New("conflict")
			}
		}

		c.Log.WithError(err).Error("error creating purchase")
		return nil, errors.New("internal server error")
	}

	if err := c.PurchaseRepository.FindByCode(tx, purchase, purchase.Code); err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating purchase")
		return nil, errors.New("internal server error")
	}

	return converter.PurchaseToResponse(purchase), nil
}

func (c *PurchaseUseCase) Update(ctx context.Context, request *model.UpdatePurchaseRequest) (*model.PurchaseResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	purchase := new(entity.Purchase)
	if err := c.PurchaseRepository.FindByCode(tx, purchase, request.Code); err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return nil, errors.New("not found")
	}

	if purchase.Status == request.Status {
		return converter.PurchaseToResponse(purchase), nil
	}

	purchase.Status = request.Status

	if err := c.PurchaseRepository.Update(tx, purchase); err != nil {
		c.Log.WithError(err).Error("error updating purchase")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating purchase")
		return nil, errors.New("internal server error")
	}

	return converter.PurchaseToResponse(purchase), nil
}

func (c *PurchaseUseCase) Get(ctx context.Context, request *model.GetPurchaseRequest) (*model.PurchaseResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	purchase := new(entity.Purchase)
	if err := c.PurchaseRepository.FindByCode(tx, purchase, request.Code); err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		return nil, errors.New("internal server error")
	}

	return converter.PurchaseToResponse(purchase), nil
}

func (c *PurchaseUseCase) Search(ctx context.Context, request *model.SearchPurchaseRequest) ([]model.PurchaseResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	purchases, total, err := c.PurchaseRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting purchases")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting purchases")
		return nil, 0, errors.New("internal server error")
	}

	responses := make([]model.PurchaseResponse, len(purchases))
	for i, purchase := range purchases {
		responses[i] = *converter.PurchaseToResponse(&purchase)
	}

	return responses, total, nil
}

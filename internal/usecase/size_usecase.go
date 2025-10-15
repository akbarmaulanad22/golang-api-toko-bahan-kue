package usecase

import (
	"context"
	"errors"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/model/converter"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SizeUseCase struct {
	DB             *gorm.DB
	Log            *logrus.Logger
	Validate       *validator.Validate
	SizeRepository *repository.SizeRepository
}

func NewSizeUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	sizeRepository *repository.SizeRepository) *SizeUseCase {
	return &SizeUseCase{
		DB:             db,
		Log:            logger,
		Validate:       validate,
		SizeRepository: sizeRepository,
	}
}

func (c *SizeUseCase) Create(ctx context.Context, request *model.CreateSizeRequest) (*model.SizeResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	size := &entity.Size{
		Name:       request.Name,
		SellPrice:  request.SellPrice,
		BuyPrice:   request.BuyPrice,
		ProductSKU: request.ProductSKU,
	}

	if err := c.SizeRepository.Create(tx, size); err != nil {
		c.Log.WithError(err).Error("error creating size")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case 1062:
				return nil, model.NewAppErr("conflict", "size already exists")
			case 1452:
				return nil, model.NewAppErr("referenced resource not found", "the specified product does not exist.")
			}
		}

		return nil, errors.New("internal server error")
	}

	if err := c.SizeRepository.FindByIdAndProductSKU(tx, size, size.ID, request.ProductSKU); err != nil {
		c.Log.WithError(err).Error("error getting size")
		return nil, helper.GetNotFoundMessage("size", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating size")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.SizeToResponse(size), nil
}

func (c *SizeUseCase) Update(ctx context.Context, request *model.UpdateSizeRequest) (*model.SizeResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	size := new(entity.Size)
	if err := c.SizeRepository.FindByIdAndProductSKU(tx, size, request.ID, request.ProductSKU); err != nil {
		c.Log.WithError(err).Error("error getting size")
		return nil, helper.GetNotFoundMessage("size", err)
	}

	if size.Name == request.Name && size.SellPrice == request.SellPrice && size.BuyPrice == request.BuyPrice {
		return converter.SizeToResponse(size), nil
	}

	size.Name = request.Name
	size.SellPrice = request.SellPrice
	size.BuyPrice = request.BuyPrice

	if err := c.SizeRepository.Update(tx, size); err != nil {
		c.Log.WithError(err).Error("error updating size")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, model.NewAppErr("conflict", "size already exists")
		}

		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating size")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.SizeToResponse(size), nil
}

func (c *SizeUseCase) Get(ctx context.Context, request *model.GetSizeRequest) (*model.SizeResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	size := new(entity.Size)
	if err := c.SizeRepository.FindByIdAndProductSKU(tx, size, request.ID, request.ProductSKU); err != nil {
		c.Log.WithError(err).Error("error getting size")
		return nil, helper.GetNotFoundMessage("size", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting size")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.SizeToResponse(size), nil
}

func (c *SizeUseCase) Delete(ctx context.Context, request *model.DeleteSizeRequest) error {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	size := new(entity.Size)
	if err := c.SizeRepository.FindByIdAndProductSKU(tx, size, request.ID, request.ProductSKU); err != nil {
		c.Log.WithError(err).Error("error getting size")
		return helper.GetNotFoundMessage("size", err)
	}

	if err := c.SizeRepository.Delete(tx, size); err != nil {
		c.Log.WithError(err).Error("error deleting size")
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting size")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

func (c *SizeUseCase) Search(ctx context.Context, request *model.SearchSizeRequest) ([]model.SizeResponse, int64, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	sizes, total, err := c.SizeRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting sizes")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting sizes")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	responses := make([]model.SizeResponse, len(sizes))
	for i, size := range sizes {
		responses[i] = *converter.SizeToResponse(&size)
	}

	return responses, total, nil
}

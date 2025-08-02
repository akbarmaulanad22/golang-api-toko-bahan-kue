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
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	size := &entity.Size{
		Name:       request.Name,
		SellPrice:  request.SellPrice,
		BuyPrice:   request.BuyPrice,
		ProductSKU: request.ProductSKU,
	}

	if err := c.SizeRepository.Create(tx, size); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Tangani duplikat
			switch {
			case strings.Contains(mysqlErr.Message, "for key"): // name
				c.Log.Warn("Size name already exists")
				return nil, errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return nil, errors.New("conflict")
			}
		}

		c.Log.WithError(err).Error("error creating size")
		return nil, errors.New("internal server error")
	}

	if err := c.SizeRepository.FindByIdAndProductSKU(tx, size, size.ID, request.ProductSKU); err != nil {
		c.Log.WithError(err).Error("error getting size")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating size")
		return nil, errors.New("internal server error")
	}

	return converter.SizeToResponse(size), nil
}

func (c *SizeUseCase) Update(ctx context.Context, request *model.UpdateSizeRequest) (*model.SizeResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	size := new(entity.Size)
	if err := c.SizeRepository.FindByIdAndProductSKU(tx, size, request.ID, request.ProductSKU); err != nil {
		c.Log.WithError(err).Error("error getting size")
		return nil, errors.New("not found")
	}

	if size.Name == request.Name && size.SellPrice == request.SellPrice && size.BuyPrice == request.BuyPrice {
		return converter.SizeToResponse(size), nil
	}

	size.Name = request.Name
	size.SellPrice = request.SellPrice
	size.BuyPrice = request.BuyPrice

	if err := c.SizeRepository.Update(tx, size); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Tangani duplikat
			switch {
			case strings.Contains(mysqlErr.Message, "for key"): // name
				c.Log.Warn("Size name already exists")
				return nil, errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return nil, errors.New("conflict")
			}
		}

		c.Log.WithError(err).Error("error creating size")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating size")
		return nil, errors.New("internal server error")
	}

	return converter.SizeToResponse(size), nil
}

func (c *SizeUseCase) Get(ctx context.Context, request *model.GetSizeRequest) (*model.SizeResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	size := new(entity.Size)
	if err := c.SizeRepository.FindByIdAndProductSKU(tx, size, request.ID, request.ProductSKU); err != nil {
		c.Log.WithError(err).Error("error getting size")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting size")
		return nil, errors.New("internal server error")
	}

	return converter.SizeToResponse(size), nil
}

func (c *SizeUseCase) Delete(ctx context.Context, request *model.DeleteSizeRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	size := new(entity.Size)
	if err := c.SizeRepository.FindByIdAndProductSKU(tx, size, request.ID, request.ProductSKU); err != nil {
		c.Log.WithError(err).Error("error getting size")
		return errors.New("not found")
	}

	if err := c.SizeRepository.Delete(tx, size); err != nil {
		c.Log.WithError(err).Error("error deleting size")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting size")
		return errors.New("internal server error")
	}

	return nil
}

func (c *SizeUseCase) Search(ctx context.Context, request *model.SearchSizeRequest) ([]model.SizeResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	sizes, total, err := c.SizeRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting sizes")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting sizes")
		return nil, 0, errors.New("internal server error")
	}

	responses := make([]model.SizeResponse, len(sizes))
	for i, size := range sizes {
		responses[i] = *converter.SizeToResponse(&size)
	}

	return responses, total, nil
}

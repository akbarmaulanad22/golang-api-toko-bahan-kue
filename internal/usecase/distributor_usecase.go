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

type DistributorUseCase struct {
	DB                    *gorm.DB
	Log                   *logrus.Logger
	Validate              *validator.Validate
	DistributorRepository *repository.DistributorRepository
}

func NewDistributorUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	distributorRepository *repository.DistributorRepository) *DistributorUseCase {
	return &DistributorUseCase{
		DB:                    db,
		Log:                   logger,
		Validate:              validate,
		DistributorRepository: distributorRepository,
	}
}

func (c *DistributorUseCase) Create(ctx context.Context, request *model.CreateDistributorRequest) (*model.DistributorResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	distributor := &entity.Distributor{
		Name:    request.Name,
		Address: request.Address,
	}

	if err := c.DistributorRepository.Create(tx, distributor); err != nil {
		c.Log.WithError(err).Error("error creating distributor")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, model.NewAppErr("conflict", "distributor name or address already exists")
		}

		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating distributor")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.DistributorToResponse(distributor), nil
}

func (c *DistributorUseCase) Update(ctx context.Context, request *model.UpdateDistributorRequest) (*model.DistributorResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	distributor := new(entity.Distributor)
	if err := c.DistributorRepository.FindById(tx, distributor, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting distributor")
		return nil, helper.GetNotFoundMessage("distributor", err)
	}

	if distributor.Name == request.Name && distributor.Address == request.Address {
		return converter.DistributorToResponse(distributor), nil
	}

	distributor.Name = request.Name
	distributor.Address = request.Address

	if err := c.DistributorRepository.Update(tx, distributor); err != nil {
		c.Log.WithError(err).Error("error updating distributor")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, model.NewAppErr("conflict", "distributor name or address already exists")
		}

		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating distributor")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.DistributorToResponse(distributor), nil
}

func (c *DistributorUseCase) Get(ctx context.Context, request *model.GetDistributorRequest) (*model.DistributorResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	distributor := new(entity.Distributor)
	if err := c.DistributorRepository.FindById(tx, distributor, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting distributor")
		return nil, helper.GetNotFoundMessage("distributor", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting distributor")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.DistributorToResponse(distributor), nil
}

func (c *DistributorUseCase) Delete(ctx context.Context, request *model.DeleteDistributorRequest) error {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	distributor := new(entity.Distributor)
	if err := c.DistributorRepository.FindById(tx, distributor, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting distributor")
		return errors.New("not found")
	}

	if err := c.DistributorRepository.Delete(tx, distributor); err != nil {
		c.Log.WithError(err).Error("error deleting distributor")
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting distributor")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

func (c *DistributorUseCase) Search(ctx context.Context, request *model.SearchDistributorRequest) ([]model.DistributorResponse, int64, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	distributors, total, err := c.DistributorRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting distributors")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting distributors")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	responses := make([]model.DistributorResponse, len(distributors))
	for i, distributor := range distributors {
		responses[i] = *converter.DistributorToResponse(&distributor)
	}

	return responses, total, nil
}

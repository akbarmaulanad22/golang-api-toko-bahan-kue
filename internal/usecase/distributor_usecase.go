package usecase

import (
	"context"
	"errors"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
	"tokobahankue/internal/model/converter"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
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
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	distributor := &entity.Distributor{
		Name:    request.Name,
		Address: request.Address,
	}

	if err := c.DistributorRepository.Create(tx, distributor); err != nil {
		c.Log.WithError(err).Error("error creating distributor")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating distributor")
		return nil, errors.New("internal server error")
	}

	return converter.DistributorToResponse(distributor), nil
}

func (c *DistributorUseCase) Update(ctx context.Context, request *model.UpdateDistributorRequest) (*model.DistributorResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	distributor := new(entity.Distributor)
	if err := c.DistributorRepository.FindById(tx, distributor, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting distributor")
		return nil, errors.New("not found")
	}

	distributor.Name = request.Name
	distributor.Address = request.Address

	if err := c.DistributorRepository.Update(tx, distributor); err != nil {
		c.Log.WithError(err).Error("error updating distributor")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating distributor")
		return nil, errors.New("internal server error")
	}

	return converter.DistributorToResponse(distributor), nil
}

func (c *DistributorUseCase) Get(ctx context.Context, request *model.GetDistributorRequest) (*model.DistributorResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	distributor := new(entity.Distributor)
	if err := c.DistributorRepository.FindById(tx, distributor, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting distributor")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting distributor")
		return nil, errors.New("internal server error")
	}

	return converter.DistributorToResponse(distributor), nil
}

func (c *DistributorUseCase) Delete(ctx context.Context, request *model.DeleteDistributorRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	distributor := new(entity.Distributor)
	if err := c.DistributorRepository.FindById(tx, distributor, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting distributor")
		return errors.New("not found")
	}

	if err := c.DistributorRepository.Delete(tx, distributor); err != nil {
		c.Log.WithError(err).Error("error deleting distributor")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting distributor")
		return errors.New("internal server error")
	}

	return nil
}

func (c *DistributorUseCase) Search(ctx context.Context, request *model.SearchDistributorRequest) ([]model.DistributorResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	distributors, total, err := c.DistributorRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting distributors")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting distributors")
		return nil, 0, errors.New("internal server error")
	}

	responses := make([]model.DistributorResponse, len(distributors))
	for i, distributor := range distributors {
		responses[i] = *converter.DistributorToResponse(&distributor)
	}

	return responses, total, nil
}

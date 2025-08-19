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

type CapitalUseCase struct {
	DB                *gorm.DB
	Log               *logrus.Logger
	Validate          *validator.Validate
	CapitalRepository *repository.CapitalRepository
}

func NewCapitalUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	capitalRepository *repository.CapitalRepository) *CapitalUseCase {
	return &CapitalUseCase{
		DB:                db,
		Log:               logger,
		Validate:          validate,
		CapitalRepository: capitalRepository,
	}
}

func (c *CapitalUseCase) Create(ctx context.Context, request *model.CreateCapitalRequest) (*model.CapitalResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	capital := &entity.Capital{
		Note:     request.Note,
		Amount:   request.Amount,
		BranchID: request.BranchID,
		Type:     request.Type,
	}

	if err := c.CapitalRepository.Create(tx, capital); err != nil {
		c.Log.WithError(err).Error("error creating capital")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating capital")
		return nil, errors.New("internal server error")
	}

	return converter.CapitalToResponse(capital), nil
}

func (c *CapitalUseCase) Update(ctx context.Context, request *model.UpdateCapitalRequest) (*model.CapitalResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	capital := new(entity.Capital)
	if err := c.CapitalRepository.FindById(tx, capital, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting capital")
		return nil, errors.New("not found")
	}

	if capital.Note == request.Note && capital.Amount == request.Amount {
		return converter.CapitalToResponse(capital), nil
	}

	capital.Note = request.Note
	capital.Amount = request.Amount
	capital.Type = request.Type

	if err := c.CapitalRepository.Update(tx, capital); err != nil {
		c.Log.WithError(err).Error("error updating capital")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating capital")
		return nil, errors.New("internal server error")
	}

	return converter.CapitalToResponse(capital), nil
}

func (c *CapitalUseCase) Delete(ctx context.Context, request *model.DeleteCapitalRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	capital := new(entity.Capital)
	if err := c.CapitalRepository.FindById(tx, capital, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting capital")
		return errors.New("not found")
	}

	if err := c.CapitalRepository.Delete(tx, capital); err != nil {
		c.Log.WithError(err).Error("error deleting capital")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting capital")
		return errors.New("internal server error")
	}

	return nil
}

func (c *CapitalUseCase) Search(ctx context.Context, request *model.SearchCapitalRequest) ([]model.CapitalResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	capitals, total, err := c.CapitalRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting capitals")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting capitals")
		return nil, 0, errors.New("internal server error")
	}

	responses := make([]model.CapitalResponse, len(capitals))
	for i, capital := range capitals {
		responses[i] = *converter.CapitalToResponse(&capital)
	}

	return responses, total, nil
}

// func (c *CapitalUseCase) ConsolidateReport(ctx context.Context, request *model.SearchConsolidateCapitalRequest) (*model.ConsolidatedCapitalResponse, error) {
// 	tx := c.DB.WithContext(ctx).Begin()
// 	defer tx.Rollback()

// 	if err := c.Validate.Struct(request); err != nil {
// 		c.Log.WithError(err).Error("error validating request body")
// 		return nil, errors.New("bad request")
// 	}

// 	capitals, err := c.CapitalRepository.ConsolidateReport(tx, request)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error getting capitals consolidated report")
// 		return nil, errors.New("internal server error")
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		c.Log.WithError(err).Error("error getting capitals consolidated report")
// 		return nil, errors.New("internal server error")
// 	}

// 	return capitals, nil
// }

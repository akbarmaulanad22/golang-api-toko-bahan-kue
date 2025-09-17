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

type DebtUseCase struct {
	DB             *gorm.DB
	Log            *logrus.Logger
	Validate       *validator.Validate
	DebtRepository *repository.DebtRepository
}

func NewDebtUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	debtRepository *repository.DebtRepository) *DebtUseCase {
	return &DebtUseCase{
		DB:             db,
		Log:            logger,
		Validate:       validate,
		DebtRepository: debtRepository,
	}
}

func (c *DebtUseCase) Get(ctx context.Context, request *model.GetDebtRequest) (*model.DebtDetailResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	debt, err := c.DebtRepository.FindDetailById(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting debt")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting debt")
		return nil, errors.New("internal server error")
	}

	return debt, nil
}

func (c *DebtUseCase) Search(ctx context.Context, request *model.SearchDebtRequest) ([]model.DebtResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	debts, total, err := c.DebtRepository.SearchRaw(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting debts")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting debts")
		return nil, 0, errors.New("internal server error")
	}

	return debts, total, nil
}

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

type DashboardUseCase struct {
	DB                  *gorm.DB
	Log                 *logrus.Logger
	Validate            *validator.Validate
	DashboardRepository *repository.DashboardRepository
}

func NewDashboardUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	dashboardRepository *repository.DashboardRepository) *DashboardUseCase {
	return &DashboardUseCase{
		DB:                  db,
		Log:                 logger,
		Validate:            validate,
		DashboardRepository: dashboardRepository,
	}
}

func (c *DashboardUseCase) Get(ctx context.Context) (*model.DashboardResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	dashboard := new(model.DashboardResponse)
	if err := c.DashboardRepository.CardCount(tx, dashboard); err != nil {
		c.Log.WithError(err).Error("error getting count card dashboard")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting count card dashboard")
		return nil, errors.New("internal server error")
	}

	return dashboard, nil
}

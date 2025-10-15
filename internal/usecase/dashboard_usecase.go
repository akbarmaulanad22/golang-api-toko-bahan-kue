package usecase

import (
	"context"
	"tokobahankue/internal/helper"
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
		c.Log.WithError(err).Error("error getting dashboard data")
		return nil, helper.GetNotFoundMessage("dashbaord data", err)

	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting dashboard data")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return dashboard, nil
}

package usecase

import (
	"context"
	"tokobahankue/internal/model"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type FinanceUseCase struct {
	DB                *gorm.DB
	Log               *logrus.Logger
	Validate          *validator.Validate
	FinanceRepository *repository.FinanceRepository
}

func NewFinanceUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	financeRepository *repository.FinanceRepository) *FinanceUseCase {
	return &FinanceUseCase{
		DB:                db,
		Log:               logger,
		Validate:          validate,
		FinanceRepository: financeRepository,
	}
}

func (c *FinanceUseCase) GetOwnerSummary(ctx context.Context, request *model.GetFinanceSummaryOwnerRequest) (*model.FinanceSummaryOwnerResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	dailyFinances, err := c.FinanceRepository.GetOwnerSummary(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting finances summary")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting finances summary")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return dailyFinances, nil
}

func (c *FinanceUseCase) GetProfitLoss(ctx context.Context, request *model.GetFinanceBasicRequest) (*model.FinanceProfitLossResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	dailyFinances, err := c.FinanceRepository.GetProfitLoss(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting finances profit loss")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting finances profit loss")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return dailyFinances, nil
}

func (c *FinanceUseCase) GetCashFlow(ctx context.Context, request *model.GetFinanceBasicRequest) (*model.FinanceCashFlowResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	dailyFinances, err := c.FinanceRepository.GetCashFlow(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting finances cash flow")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting finances cash flow")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return dailyFinances, nil
}

func (c *FinanceUseCase) GetBalanceSheet(ctx context.Context, request *model.GetFinanceBalanceSheetRequest) (*model.FinanceBalanceSheetResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	dailyFinances, err := c.FinanceRepository.GetBalanceSheet(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting finances balance sheet")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting finances balance sheet")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return dailyFinances, nil
}

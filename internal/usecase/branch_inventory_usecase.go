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

type BranchInventoryUseCase struct {
	DB                        *gorm.DB
	Log                       *logrus.Logger
	Validate                  *validator.Validate
	BranchInventoryRepository *repository.BranchInventoryRepository
}

func NewBranchInventoryUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	branchRepository *repository.BranchInventoryRepository) *BranchInventoryUseCase {
	return &BranchInventoryUseCase{
		DB:                        db,
		Log:                       logger,
		Validate:                  validate,
		BranchInventoryRepository: branchRepository,
	}
}

// func (c *BranchInventoryUseCase) ListOwnerInventoryByBranch(ctx context.Context) ([]model.BranchInventoryResponse, error) {

// 	tx := c.DB.WithContext(ctx).Begin()
// 	defer tx.Rollback()

// 	inventory, err := c.BranchInventoryRepository.ListOwnerInventoryByBranch(tx)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error getting inventory")
// 		return nil, errors.New("internal server error")
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		c.Log.WithError(err).Error("error getting inventory")
// 		return nil, errors.New("internal server error")
// 	}

// 	return inventory, nil

// }

func (c *BranchInventoryUseCase) List(ctx context.Context, request *model.SearchBranchInventoryRequest) ([]model.BranchInventoryProductResponse, int64, error) {

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	inventory, total, err := c.BranchInventoryRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting inventory")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting inventory")
		return nil, 0, errors.New("internal server error")
	}

	return inventory, total, nil

}

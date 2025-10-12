package usecase

import (
	"context"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
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

func (c *BranchInventoryUseCase) List(ctx context.Context, request *model.SearchBranchInventoryRequest) ([]model.BranchInventoryProductResponse, int64, error) {

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	inventory, total, err := c.BranchInventoryRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting inventory")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting inventory")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	return inventory, total, nil

}

func (c *BranchInventoryUseCase) Create(ctx context.Context, request *model.CreateBranchInventoryRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	inventory := &entity.BranchInventory{
		BranchID: request.BranchID,
		SizeID:   request.SizeID,
	}

	if err := c.BranchInventoryRepository.Create(tx, inventory); err != nil {
		c.Log.WithError(err).Error("error creating inventory")

		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return model.NewAppErr("conflict", "size already exists for this branch")
		}

		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating inventory")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

func (c *BranchInventoryUseCase) Update(ctx context.Context, request *model.UpdateBranchInventoryRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)

	}

	inventory := new(entity.BranchInventory)
	if err := c.BranchInventoryRepository.FindById(tx, inventory, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting inventory")
		return helper.GetNotFoundMessage("inventory", err)

	}

	inventory.BranchID = request.BranchID
	inventory.SizeID = request.SizeID

	if err := c.BranchInventoryRepository.Update(tx, inventory); err != nil {
		c.Log.WithError(err).Error("error updating inventory")

		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return model.NewAppErr("conflict", "size already exists for this branch")
		}

		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating inventory")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

func (c *BranchInventoryUseCase) Delete(ctx context.Context, request *model.DeleteBranchInventoryRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	branchInventory := new(entity.BranchInventory)
	if err := c.BranchInventoryRepository.FindById(tx, branchInventory, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting inventory")
		return helper.GetNotFoundMessage("inventory", err)
	}

	if err := c.BranchInventoryRepository.Delete(tx, branchInventory); err != nil {
		c.Log.WithError(err).Error("error deleting inventory")
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting inventory")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

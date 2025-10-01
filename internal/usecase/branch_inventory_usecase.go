package usecase

import (
	"context"
	"errors"
	"strings"
	"tokobahankue/internal/entity"
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

func (c *BranchInventoryUseCase) Create(ctx context.Context, request *model.CreateBranchInventoryRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	branch := &entity.BranchInventory{
		BranchID: request.BranchID,
		SizeID:   request.SizeID,
	}

	if err := c.BranchInventoryRepository.Create(tx, branch); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Tangani duplikat
			switch {
			case strings.Contains(mysqlErr.Message, "for key 'branch_inventory.branch_id'"):
				c.Log.Warn("Branch Inventory branch_id already exists")
				return errors.New("conflict")
			case strings.Contains(mysqlErr.Message, "for key 'branch_inventory.size_id'"):
				c.Log.Warn("Branch Inventory size_id already exists")
				return errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return errors.New("conflict")
			}
		}

		c.Log.WithError(err).Error("error creating branch")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating branch")
		return errors.New("internal server error")
	}

	return nil
}

func (c *BranchInventoryUseCase) Update(ctx context.Context, request *model.UpdateBranchInventoryRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	branch := new(entity.BranchInventory)
	if err := c.BranchInventoryRepository.FindById(tx, branch, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting branch")
		return errors.New("not found")
	}

	branch.BranchID = request.BranchID
	branch.SizeID = request.SizeID

	if err := c.BranchInventoryRepository.Update(tx, branch); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Tangani duplikat
			switch {
			case strings.Contains(mysqlErr.Message, "for key 'branch_inventory.branch_id'"):
				c.Log.Warn("Branch Inventory branch_id already exists")
				return errors.New("conflict")
			case strings.Contains(mysqlErr.Message, "for key 'branch_inventory.size_id'"):
				c.Log.Warn("Branch Inventory size_id already exists")
				return errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return errors.New("conflict")
			}
		}

		c.Log.WithError(err).Error("error updating branch")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating branch")
		return errors.New("internal server error")
	}

	return nil
}

func (c *BranchInventoryUseCase) Delete(ctx context.Context, request *model.DeleteBranchInventoryRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	branchInventory := new(entity.BranchInventory)
	if err := c.BranchInventoryRepository.FindById(tx, branchInventory, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting branch inventory")
		return errors.New("not found")
	}

	if err := c.BranchInventoryRepository.Delete(tx, branchInventory); err != nil {
		c.Log.WithError(err).Error("error deleting branch inventory")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting branch inventory")
		return errors.New("internal server error")
	}

	return nil
}

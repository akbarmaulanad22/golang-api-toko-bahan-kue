package usecase

import (
	"context"
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

type BranchUseCase struct {
	DB               *gorm.DB
	Log              *logrus.Logger
	Validate         *validator.Validate
	BranchRepository *repository.BranchRepository
}

func NewBranchUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	branchRepository *repository.BranchRepository) *BranchUseCase {
	return &BranchUseCase{
		DB:               db,
		Log:              logger,
		Validate:         validate,
		BranchRepository: branchRepository,
	}
}

func (c *BranchUseCase) Create(ctx context.Context, request *model.CreateBranchRequest) (*model.BranchResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	branch := &entity.Branch{
		Name:    request.Name,
		Address: request.Address,
	}

	if err := c.BranchRepository.Create(tx, branch); err != nil {
		c.Log.WithError(err).Error("error creating branch")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, model.NewAppErr("conflict", "branch name or address already exists")
		}

		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating branch")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.BranchToResponse(branch), nil
}

func (c *BranchUseCase) Update(ctx context.Context, request *model.UpdateBranchRequest) (*model.BranchResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	branch := new(entity.Branch)
	if err := c.BranchRepository.FindById(tx, branch, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting branch")
		return nil, helper.GetNotFoundMessage("branch", err)
	}

	if branch.Name == request.Name && branch.Address == request.Address {
		return converter.BranchToResponse(branch), nil
	}

	branch.Name = request.Name
	branch.Address = request.Address

	if err := c.BranchRepository.Update(tx, branch); err != nil {
		c.Log.WithError(err).Error("error updating branch")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, model.NewAppErr("conflict", "branch name or address already exists")
		}

		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating branch")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.BranchToResponse(branch), nil
}

func (c *BranchUseCase) Get(ctx context.Context, request *model.GetBranchRequest) (*model.BranchResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	branch := new(entity.Branch)
	if err := c.BranchRepository.FindById(tx, branch, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting branch")
		return nil, helper.GetNotFoundMessage("branch", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting branch")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.BranchToResponse(branch), nil
}

func (c *BranchUseCase) Delete(ctx context.Context, request *model.DeleteBranchRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	branch := new(entity.Branch)
	if err := c.BranchRepository.FindById(tx, branch, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting branch")
		return helper.GetNotFoundMessage("branch", err)
	}

	if err := c.BranchRepository.Delete(tx, branch); err != nil {
		c.Log.WithError(err).Error("error deleting branch")
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting branch")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

func (c *BranchUseCase) Search(ctx context.Context, request *model.SearchBranchRequest) ([]model.BranchResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	branches, total, err := c.BranchRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting branches")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting branches")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	responses := make([]model.BranchResponse, len(branches))
	for i, branch := range branches {
		responses[i] = *converter.BranchToResponse(&branch)
	}

	return responses, total, nil
}

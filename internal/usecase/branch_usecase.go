package usecase

import (
	"context"
	"errors"
	"strings"
	"tokobahankue/internal/entity"
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
		return nil, errors.New("bad request")
	}

	branch := &entity.Branch{
		Name:    request.Name,
		Address: request.Address,
	}

	if err := c.BranchRepository.Create(tx, branch); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Tangani duplikat
			switch {
			case strings.Contains(mysqlErr.Message, "for key 'branches.name'"): // name
				c.Log.Warn("Branch name already exists")
				return nil, errors.New("conflict")
			case strings.Contains(mysqlErr.Message, "for key 'branches.address'"): // address
				c.Log.Warn("Branch address already exists")
				return nil, errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return nil, errors.New("conflict")
			}
		}

		c.Log.WithError(err).Error("error creating branch")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating branch")
		return nil, errors.New("internal server error")
	}

	return converter.BranchToResponse(branch), nil
}

func (c *BranchUseCase) Update(ctx context.Context, request *model.UpdateBranchRequest) (*model.BranchResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	branch := new(entity.Branch)
	if err := c.BranchRepository.FindById(tx, branch, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting branch")
		return nil, errors.New("not found")
	}

	if branch.Name == request.Name && branch.Address == request.Address {
		return converter.BranchToResponse(branch), nil
	}

	branch.Name = request.Name
	branch.Address = request.Address

	if err := c.BranchRepository.Update(tx, branch); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Tangani duplikat
			switch {
			case strings.Contains(mysqlErr.Message, "for key 'branches.name'"): // name
				c.Log.Warn("Branch name already exists")
				return nil, errors.New("conflict")
			case strings.Contains(mysqlErr.Message, "for key 'branches.address'"): // address
				c.Log.Warn("Branch address already exists")
				return nil, errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return nil, errors.New("conflict")
			}
		}

		c.Log.WithError(err).Error("error updating branch")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating branch")
		return nil, errors.New("internal server error")
	}

	return converter.BranchToResponse(branch), nil
}

func (c *BranchUseCase) Get(ctx context.Context, request *model.GetBranchRequest) (*model.BranchResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	branch := new(entity.Branch)
	if err := c.BranchRepository.FindById(tx, branch, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting branch")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting branch")
		return nil, errors.New("internal server error")
	}

	return converter.BranchToResponse(branch), nil
}

func (c *BranchUseCase) Delete(ctx context.Context, request *model.DeleteBranchRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	branch := new(entity.Branch)
	if err := c.BranchRepository.FindById(tx, branch, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting branch")
		return errors.New("not found")
	}

	if err := c.BranchRepository.Delete(tx, branch); err != nil {
		c.Log.WithError(err).Error("error deleting branch")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting branch")
		return errors.New("internal server error")
	}

	return nil
}

func (c *BranchUseCase) Search(ctx context.Context, request *model.SearchBranchRequest) ([]model.BranchResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	branchs, total, err := c.BranchRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting branchs")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting branchs")
		return nil, 0, errors.New("internal server error")
	}

	responses := make([]model.BranchResponse, len(branchs))
	for i, branch := range branchs {
		responses[i] = *converter.BranchToResponse(&branch)
	}

	return responses, total, nil
}

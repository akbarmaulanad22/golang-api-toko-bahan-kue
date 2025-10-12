package usecase

import (
	"context"
	"errors"
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

type CategoryUseCase struct {
	DB                 *gorm.DB
	Log                *logrus.Logger
	Validate           *validator.Validate
	CategoryRepository *repository.CategoryRepository
}

func NewCategoryUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	categoryRepository *repository.CategoryRepository) *CategoryUseCase {
	return &CategoryUseCase{
		DB:                 db,
		Log:                logger,
		Validate:           validate,
		CategoryRepository: categoryRepository,
	}
}

func (c *CategoryUseCase) Create(ctx context.Context, request *model.CreateCategoryRequest) (*model.CategoryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	category := &entity.Category{
		Name: request.Name,
	}

	if err := c.CategoryRepository.Create(tx, category); err != nil {
		c.Log.WithError(err).Error("error creating category")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, model.NewAppErr("conflict", "category already exists")
		}

		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating category")
		return nil, errors.New("internal server error")
	}

	return converter.CategoryToResponse(category), nil
}

func (c *CategoryUseCase) Update(ctx context.Context, request *model.UpdateCategoryRequest) (*model.CategoryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	category := new(entity.Category)
	if err := c.CategoryRepository.FindById(tx, category, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting category")
		return nil, helper.GetNotFoundMessage("category", err)
	}

	if category.Name == request.Name {
		return converter.CategoryToResponse(category), nil
	}

	category.Name = request.Name

	if err := c.CategoryRepository.Update(tx, category); err != nil {
		c.Log.WithError(err).Error("error updating category")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, model.NewAppErr("conflict", "category already exists")
		}

		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating category")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.CategoryToResponse(category), nil
}

func (c *CategoryUseCase) Get(ctx context.Context, request *model.GetCategoryRequest) (*model.CategoryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	category := new(entity.Category)
	if err := c.CategoryRepository.FindById(tx, category, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting category")
		return nil, helper.GetNotFoundMessage("category", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting category")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.CategoryToResponse(category), nil
}

func (c *CategoryUseCase) Delete(ctx context.Context, request *model.DeleteCategoryRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	category := new(entity.Category)
	if err := c.CategoryRepository.FindById(tx, category, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting category")
		return helper.GetNotFoundMessage("category", err)
	}

	if err := c.CategoryRepository.Delete(tx, category); err != nil {
		c.Log.WithError(err).Error("error deleting category")
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting category")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

func (c *CategoryUseCase) Search(ctx context.Context, request *model.SearchTopSellerCategoryRequest) ([]model.CategoryResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	categories, total, err := c.CategoryRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting categories")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting categories")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	responses := make([]model.CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = *converter.CategoryToResponse(&category)
	}

	return responses, total, nil
}

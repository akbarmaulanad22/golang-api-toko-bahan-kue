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
		return nil, errors.New("bad request")
	}

	category := &entity.Category{
		Name: request.Name,
	}

	if err := c.CategoryRepository.Create(tx, category); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Tangani duplikat
			switch {
			case strings.Contains(mysqlErr.Message, "for key 'categories.name'"): // name
				c.Log.Warn("Category name already exists")
				return nil, errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return nil, errors.New("conflict")
			}
		}

		c.Log.WithError(err).Error("error creating category")
		return nil, errors.New("internal server error")
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
		return nil, errors.New("bad request")
	}

	category := new(entity.Category)
	if err := c.CategoryRepository.FindById(tx, category, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting category")
		return nil, errors.New("not found")
	}

	if category.Name == request.Name {
		return converter.CategoryToResponse(category), nil
	}

	category.Name = request.Name

	if err := c.CategoryRepository.Update(tx, category); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Tangani duplikat
			switch {
			case strings.Contains(mysqlErr.Message, "for key 'categories.name'"): // name
				c.Log.Warn("Category name already exists")
				return nil, errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return nil, errors.New("conflict")
			}
		}

		c.Log.WithError(err).Error("error updating category")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating category")
		return nil, errors.New("internal server error")
	}

	return converter.CategoryToResponse(category), nil
}

func (c *CategoryUseCase) Get(ctx context.Context, request *model.GetCategoryRequest) (*model.CategoryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	category := new(entity.Category)
	if err := c.CategoryRepository.FindById(tx, category, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting category")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting category")
		return nil, errors.New("internal server error")
	}

	return converter.CategoryToResponse(category), nil
}

func (c *CategoryUseCase) Delete(ctx context.Context, request *model.DeleteCategoryRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	category := new(entity.Category)
	if err := c.CategoryRepository.FindById(tx, category, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting category")
		return errors.New("not found")
	}

	if err := c.CategoryRepository.Delete(tx, category); err != nil {
		c.Log.WithError(err).Error("error deleting category")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting category")
		return errors.New("internal server error")
	}

	return nil
}

func (c *CategoryUseCase) Search(ctx context.Context, request *model.SearchCategoryRequest) ([]model.CategoryResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	categorys, total, err := c.CategoryRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting categorys")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting categorys")
		return nil, 0, errors.New("internal server error")
	}

	responses := make([]model.CategoryResponse, len(categorys))
	for i, category := range categorys {
		responses[i] = *converter.CategoryToResponse(&category)
	}

	return responses, total, nil
}

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

type ProductUseCase struct {
	DB                *gorm.DB
	Log               *logrus.Logger
	Validate          *validator.Validate
	ProductRepository *repository.ProductRepository
}

func NewProductUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	productRepository *repository.ProductRepository) *ProductUseCase {
	return &ProductUseCase{
		DB:                db,
		Log:               logger,
		Validate:          validate,
		ProductRepository: productRepository,
	}
}

func (c *ProductUseCase) Create(ctx context.Context, request *model.CreateProductRequest) (*model.ProductResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	product := &entity.Product{
		CategoryID: request.CategoryID,
		SKU:        request.SKU,
		Name:       request.Name,
	}

	if err := c.ProductRepository.Create(tx, product); err != nil {
		c.Log.WithError(err).Error("error creating product")

		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case 1062:
				return nil, model.NewAppErr("conflict", "product sku or name already exists")
			case 1452:
				return nil, model.NewAppErr("referenced resource not found", "the specified category does not exist.")
			}
		}

		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating product")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.ProductToResponse(product), nil
}

func (c *ProductUseCase) Update(ctx context.Context, request *model.UpdateProductRequest) (*model.ProductResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	product := new(entity.Product)
	if err := c.ProductRepository.FindBySKU(tx, product, request.SKU); err != nil {
		c.Log.WithError(err).Error("error getting product")
		return nil, helper.GetNotFoundMessage("product", err)
	}

	if product.Name == request.Name && product.CategoryID == request.CategoryID {
		return converter.ProductToResponse(product), nil
	}

	product.CategoryID = request.CategoryID
	product.Name = request.Name

	if err := c.ProductRepository.Update(tx, product); err != nil {
		c.Log.WithError(err).Error("error updating product")

		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case 1062:
				return nil, model.NewAppErr("conflict", "product sku or name already exists")
			case 1452:
				return nil, model.NewAppErr("referenced resource not found", "the specified category does not exist.")
			}
		}

		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating product")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.ProductToResponse(product), nil
}

func (c *ProductUseCase) Get(ctx context.Context, request *model.GetProductRequest) (*model.ProductResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)

	}

	product := new(entity.Product)
	if err := c.ProductRepository.FindBySKU(tx, product, request.SKU); err != nil {
		c.Log.WithError(err).Error("error getting product")
		return nil, helper.GetNotFoundMessage("product", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting product")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.ProductToResponse(product), nil
}

func (c *ProductUseCase) Delete(ctx context.Context, request *model.DeleteProductRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	product := new(entity.Product)
	if err := c.ProductRepository.FindBySKU(tx, product, request.SKU); err != nil {
		c.Log.WithError(err).Error("error getting product")
		return helper.GetNotFoundMessage("product", err)
	}

	if err := c.ProductRepository.Delete(tx, product); err != nil {
		c.Log.WithError(err).Error("error deleting product")
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting product")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

func (c *ProductUseCase) Search(ctx context.Context, request *model.SearchProductRequest) ([]model.ProductResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	products, total, err := c.ProductRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting products")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting products")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	responses := make([]model.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = *converter.ProductToResponse(&product)
	}

	return responses, total, nil
}

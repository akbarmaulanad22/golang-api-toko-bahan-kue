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
		return nil, errors.New("bad request")
	}

	product := &entity.Product{
		CategorySlug: request.CategorySlug,
		SKU:          request.SKU,
		Name:         request.Name,
	}

	if err := c.ProductRepository.Create(tx, product); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Tangani duplikat
			switch {
			case strings.Contains(mysqlErr.Message, "for key 'products.PRIMARY'"): // sku
				c.Log.Warn("SKU already exists")
				return nil, errors.New("conflict")
			case strings.Contains(mysqlErr.Message, "for key 'products.name'"): // name
				c.Log.Warn("Product name already exists")
				return nil, errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return nil, errors.New("conflict")
			}
		}

		c.Log.WithError(err).Error("error creating product")
		return nil, errors.New("internal server error")
	}

	if err := c.ProductRepository.FindBySKU(tx, product, product.SKU); err != nil {
		c.Log.WithError(err).Error("error getting product")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating product")
		return nil, errors.New("internal server error")
	}

	return converter.ProductToResponse(product), nil
}

func (c *ProductUseCase) Update(ctx context.Context, request *model.UpdateProductRequest) (*model.ProductResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	product := new(entity.Product)
	if err := c.ProductRepository.FindBySKU(tx, product, request.SKU); err != nil {
		c.Log.WithError(err).Error("error getting product")
		return nil, errors.New("not found")
	}

	if product.Name == request.Name && product.CategorySlug == request.CategorySlug {
		return converter.ProductToResponse(product), nil
	}

	product.CategorySlug = request.CategorySlug
	product.Name = request.Name

	if err := c.ProductRepository.Update(tx, product); err != nil {
		c.Log.WithError(err).Error("error updating product")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating product")
		return nil, errors.New("internal server error")
	}

	return converter.ProductToResponse(product), nil
}

func (c *ProductUseCase) Get(ctx context.Context, request *model.GetProductRequest) (*model.ProductResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	product := new(entity.Product)
	if err := c.ProductRepository.FindBySKU(tx, product, request.SKU); err != nil {
		c.Log.WithError(err).Error("error getting product")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting product")
		return nil, errors.New("internal server error")
	}

	return converter.ProductToResponse(product), nil
}

func (c *ProductUseCase) Delete(ctx context.Context, request *model.DeleteProductRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	product := new(entity.Product)
	if err := c.ProductRepository.FindBySKU(tx, product, request.SKU); err != nil {
		c.Log.WithError(err).Error("error getting product")
		return errors.New("not found")
	}

	if err := c.ProductRepository.Delete(tx, product); err != nil {
		c.Log.WithError(err).Error("error deleting product")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting product")
		return errors.New("internal server error")
	}

	return nil
}

func (c *ProductUseCase) Search(ctx context.Context, request *model.SearchProductRequest) ([]model.ProductResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	products, total, err := c.ProductRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting products")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting products")
		return nil, 0, errors.New("internal server error")
	}

	responses := make([]model.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = *converter.ProductToResponse(&product)
	}

	return responses, total, nil
}

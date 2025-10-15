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

type RoleUseCase struct {
	DB             *gorm.DB
	Log            *logrus.Logger
	Validate       *validator.Validate
	RoleRepository *repository.RoleRepository
}

func NewRoleUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	roleRepository *repository.RoleRepository) *RoleUseCase {
	return &RoleUseCase{
		DB:             db,
		Log:            logger,
		Validate:       validate,
		RoleRepository: roleRepository,
	}
}

func (c *RoleUseCase) Create(ctx context.Context, request *model.CreateRoleRequest) (*model.RoleResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	role := &entity.Role{
		Name: request.Name,
	}

	if err := c.RoleRepository.Create(tx, role); err != nil {
		c.Log.WithError(err).Error("error creating role")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, model.NewAppErr("conflict", "role already exists")
		}
		return nil, model.NewAppErr("internal server error", nil)

	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating role")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.RoleToResponse(role), nil
}

func (c *RoleUseCase) Update(ctx context.Context, request *model.UpdateRoleRequest) (*model.RoleResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	role := new(entity.Role)
	if err := c.RoleRepository.FindById(tx, role, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting role")
		return nil, helper.GetNotFoundMessage("role", err)
	}

	if role.Name == request.Name {
		return converter.RoleToResponse(role), nil
	}

	role.Name = request.Name

	if err := c.RoleRepository.Update(tx, role); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, model.NewAppErr("conflict", "role already exists")
		}
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating role")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.RoleToResponse(role), nil
}

func (c *RoleUseCase) Get(ctx context.Context, request *model.GetRoleRequest) (*model.RoleResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	role := new(entity.Role)
	if err := c.RoleRepository.FindById(tx, role, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting role")
		return nil, helper.GetNotFoundMessage("role", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting role")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.RoleToResponse(role), nil
}

func (c *RoleUseCase) Delete(ctx context.Context, request *model.DeleteRoleRequest) error {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	role := new(entity.Role)
	if err := c.RoleRepository.FindById(tx, role, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting role")
		return helper.GetNotFoundMessage("role", err)
	}

	if err := c.RoleRepository.Delete(tx, role); err != nil {
		c.Log.WithError(err).Error("error deleting role")
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting role")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

func (c *RoleUseCase) Search(ctx context.Context, request *model.SearchRoleRequest) ([]model.RoleResponse, int64, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	roles, total, err := c.RoleRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting roles")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting roles")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	responses := make([]model.RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = *converter.RoleToResponse(&role)
	}

	return responses, total, nil
}

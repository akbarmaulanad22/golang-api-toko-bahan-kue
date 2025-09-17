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
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	role := &entity.Role{
		Name: request.Name,
	}

	if err := c.RoleRepository.Create(tx, role); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Tangani duplikat
			switch {
			case strings.Contains(mysqlErr.Message, "for key 'roles.name'"): // name
				c.Log.Warn("role name already exists")
				return nil, errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return nil, errors.New("conflict")
			}
		}

		c.Log.WithError(err).Error("error creating role")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating role")
		return nil, errors.New("internal server error")
	}

	return converter.RoleToResponse(role), nil
}

func (c *RoleUseCase) Update(ctx context.Context, request *model.UpdateRoleRequest) (*model.RoleResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	role := new(entity.Role)
	if err := c.RoleRepository.FindById(tx, role, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting role")
		return nil, errors.New("not found")
	}

	if role.Name == request.Name {
		return converter.RoleToResponse(role), nil
	}

	role.Name = request.Name

	if err := c.RoleRepository.Update(tx, role); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Tangani duplikat
			switch {
			case strings.Contains(mysqlErr.Message, "for key 'roles.name'"): // name
				c.Log.Warn("role name already exists")
				return nil, errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return nil, errors.New("conflict")
			}
		}

		c.Log.WithError(err).Error("error updating role")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating role")
		return nil, errors.New("internal server error")
	}

	return converter.RoleToResponse(role), nil
}

func (c *RoleUseCase) Get(ctx context.Context, request *model.GetRoleRequest) (*model.RoleResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	role := new(entity.Role)
	if err := c.RoleRepository.FindById(tx, role, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting role")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting role")
		return nil, errors.New("internal server error")
	}

	return converter.RoleToResponse(role), nil
}

func (c *RoleUseCase) Delete(ctx context.Context, request *model.DeleteRoleRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	role := new(entity.Role)
	if err := c.RoleRepository.FindById(tx, role, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting role")
		return errors.New("not found")
	}

	if err := c.RoleRepository.Delete(tx, role); err != nil {
		c.Log.WithError(err).Error("error deleting role")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting role")
		return errors.New("internal server error")
	}

	return nil
}

func (c *RoleUseCase) Search(ctx context.Context, request *model.SearchRoleRequest) ([]model.RoleResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	roles, total, err := c.RoleRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting roles")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting roles")
		return nil, 0, errors.New("internal server error")
	}

	responses := make([]model.RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = *converter.RoleToResponse(&role)
	}

	return responses, total, nil
}

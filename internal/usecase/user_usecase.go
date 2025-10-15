package usecase

import (
	"context"
	"errors"
	"strings"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/model/converter"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserUseCase struct {
	DB             *gorm.DB
	Log            *logrus.Logger
	Validate       *validator.Validate
	UserRepository *repository.UserRepository
}

func NewUserUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	userRepository *repository.UserRepository) *UserUseCase {
	return &UserUseCase{
		DB:             db,
		Log:            logger,
		Validate:       validate,
		UserRepository: userRepository,
	}
}

func (c *UserUseCase) Verify(ctx context.Context, request *model.VerifyUserRequest) (*model.Auth, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	err := c.Validate.Struct(request)
	if err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	user := new(entity.User)
	if err := c.UserRepository.FindByToken(tx, user, request.Token); err != nil {
		c.Log.WithError(err).Error("error getting user")
		return nil, helper.GetNotFoundMessage("user", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error verify user")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return &model.Auth{Username: user.Username, BranchID: user.BranchID, Role: user.Role.Name}, nil
}

func (c *UserUseCase) Current(ctx context.Context, request *model.GetUserRequest) (*model.UserResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	user := new(entity.User)
	if err := c.UserRepository.FindByUsername(tx, user, request.Username); err != nil {
		c.Log.WithError(err).Error("error getting user")
		return nil, helper.GetNotFoundMessage("user", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting user")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.UserToResponse(user), nil
}

func (c *UserUseCase) Login(ctx context.Context, request *model.LoginUserRequest) (*model.UserResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	user := new(entity.User)
	if err := c.UserRepository.FindByUsername(tx, user, request.Username); err != nil {
		c.Log.WithError(err).Error("error getting user")
		return nil, model.NewAppErr("unauthorized", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		c.Log.WithError(err).Error("error compare user password")
		return nil, model.NewAppErr("unauthorized", nil)
	}

	user.Token = uuid.New().String()
	if err := c.UserRepository.Update(tx, user); err != nil {
		c.Log.WithError(err).Error("error login user")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error login user")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.UserToTokenResponse(user), nil
}

func (c *UserUseCase) Logout(ctx context.Context, request *model.LogoutUserRequest) (bool, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return false, helper.GetValidationMessage(err)
	}

	user := new(entity.User)
	if err := c.UserRepository.FindByUsername(tx, user, request.Username); err != nil {
		c.Log.WithError(err).Error("error getting user")
		return false, helper.GetNotFoundMessage("branch", err)
	}

	user.Token = ""

	if err := c.UserRepository.Update(tx, user); err != nil {
		c.Log.WithError(err).Error("error logout user")
		return false, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error logout user")
		return false, model.NewAppErr("internal server error", nil)
	}

	return true, nil
}

// general CRUD

func (c *UserUseCase) Create(ctx context.Context, request *model.RegisterUserRequest) (*model.UserResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	err := c.Validate.Struct(request)
	if err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	password, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Log.WithError(err).Error("error generate bcrype hash user password")
		return nil, model.NewAppErr("internal server error", nil)
	}

	user := &entity.User{
		Username: request.Username,
		Password: string(password),
		Name:     request.Name,
		Address:  request.Address,
		RoleID:   request.RoleID,
		BranchID: request.BranchID,
	}

	if err := c.UserRepository.Create(tx, user); err != nil {
		c.Log.WithError(err).Error("error creating user")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case 1452:
				return nil, model.NewAppErr("referenced resource not found", "the specified role or branch does not exist.")
			case 1062:
				return nil, model.NewAppErr("conflict", "username already exists")
			}
		}

		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating user")
		return nil, errors.New("internal server error")
	}

	return converter.UserToResponse(user), nil
}

func (c *UserUseCase) Search(ctx context.Context, request *model.SearchUserRequest) ([]model.UserResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	users, total, err := c.UserRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting users")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting users")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	responses := make([]model.UserResponse, len(users))
	for i, user := range users {
		responses[i] = *converter.UserToResponse(&user)
	}

	return responses, total, nil
}

func (c *UserUseCase) Get(ctx context.Context, request *model.GetUserRequest) (*model.UserResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	user := new(entity.User)
	if err := c.UserRepository.FindByUsername(tx, user, request.Username); err != nil {
		c.Log.WithError(err).Error("error getting user")
		return nil, helper.GetNotFoundMessage("user", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting user")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.UserToResponse(user), nil
}

func (c *UserUseCase) Update(ctx context.Context, request *model.UpdateUserRequest) (*model.UserResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	user := new(entity.User)
	if err := c.UserRepository.FindByUsername(tx, user, request.Username); err != nil {
		c.Log.WithError(err).Error("error getting user")
		return nil, helper.GetNotFoundMessage("user", err)
	}

	if strings.ToUpper(user.Role.Name) == "OWNER" {
		c.Log.Warn("cannot update owner data")
		return nil, model.NewAppErr("forbidden", "cannot update owner")
	}

	user.Name = request.Name
	user.Address = request.Address
	user.RoleID = request.RoleID
	user.BranchID = request.BranchID

	password, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Log.WithError(err).Error("error generate bcrype hash user password")
		return nil, model.NewAppErr("internal server error", nil)
	}
	user.Password = string(password)

	if err := c.UserRepository.Update(tx, user); err != nil {
		c.Log.WithError(err).Error("error updating user")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case 1452:
				return nil, model.NewAppErr("referenced resource not found", "the specified role or branch does not exist.")
			case 1062:
				return nil, model.NewAppErr("conflict", "username already exists")
			}
		}
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating user")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return converter.UserToResponse(user), nil
}

func (c *UserUseCase) Delete(ctx context.Context, request *model.DeleteUserRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return helper.GetValidationMessage(err)
	}

	user := new(entity.User)
	if err := c.UserRepository.FindByUsername(tx, user, request.Username); err != nil {
		c.Log.WithError(err).Error("error getting user")
		return helper.GetNotFoundMessage("user", err)
	}

	if strings.ToUpper(user.Role.Name) == "OWNER" {
		c.Log.Warn("cannot delete owner data")
		return model.NewAppErr("forbidden", "cannot delete owner")
	}

	if err := c.UserRepository.Delete(tx, user); err != nil {
		c.Log.WithError(err).Error("error deleting user")
		return model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting user")
		return model.NewAppErr("internal server error", nil)
	}

	return nil
}

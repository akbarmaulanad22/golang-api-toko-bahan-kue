package http

import (
	"encoding/json"
	"math"
	"net/http"
	"tokobahankue/internal/delivery/http/middleware"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type UserController struct {
	Log     *logrus.Logger
	UseCase *usecase.UserUseCase
}

func NewUserController(useCase *usecase.UserUseCase, logger *logrus.Logger) *UserController {
	return &UserController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *UserController) Register(w http.ResponseWriter, r *http.Request) error {
	var request model.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error creating user")
		return err
	}

	return helper.WriteJSON(w, http.StatusCreated, model.WebResponse[*model.UserResponse]{Data: response})
}

func (c *UserController) Current(w http.ResponseWriter, r *http.Request) error {
	auth := middleware.GetUser(r)

	request := &model.GetUserRequest{
		Username: auth.Username,
	}

	response, err := c.UseCase.Current(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("error to get current user")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.UserResponse]{Data: response})
}

func (c *UserController) Login(w http.ResponseWriter, r *http.Request) error {
	var request model.LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	response, err := c.UseCase.Login(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error login user")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.UserResponse]{Data: response})
}

func (c *UserController) Logout(w http.ResponseWriter, r *http.Request) error {
	auth := middleware.GetUser(r)

	request := &model.LogoutUserRequest{
		Username: auth.Username,
	}

	_, err := c.UseCase.Logout(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("error to logout user")
		return err
	}

	return helper.WriteJSON(w, http.StatusNoContent, nil)
}

func (c *UserController) List(w http.ResponseWriter, r *http.Request) error {
	params := r.URL.Query()

	page := helper.ParseIntOrDefault(params.Get("page"), 1)
	size := helper.ParseIntOrDefault(params.Get("size"), 10)
	role := helper.ParseIntOrDefault(params.Get("role_id"), 0)
	branch := helper.ParseIntOrDefault(params.Get("branch_id"), 0)

	request := &model.SearchUserRequest{
		Search:   params.Get("search"),
		RoleID:   uint(role),
		BranchID: uint(branch),
		Page:     page,
		Size:     size,
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching user")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.UserResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *UserController) Get(w http.ResponseWriter, r *http.Request) error {
	username := mux.Vars(r)["username"]

	if username == "" {
		c.Log.Warnf("error to parse username parameter: missing or invalid username")
		return model.NewAppErr("invalid username parameter", nil)
	}

	request := &model.GetUserRequest{
		Username: username,
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting user")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.UserResponse]{Data: response})
}

func (c *UserController) Update(w http.ResponseWriter, r *http.Request) error {

	username := mux.Vars(r)["username"]

	if username == "" {
		c.Log.Warnf("error to parse username parameter: missing or invalid username")
		return model.NewAppErr("invalid username parameter", nil)
	}

	request := new(model.UpdateUserRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	request.Username = username

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("error to update user")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.UserResponse]{Data: response})
}

func (c *UserController) Delete(w http.ResponseWriter, r *http.Request) error {
	username := mux.Vars(r)["username"]

	if username == "" {
		c.Log.Warnf("error to parse username parameter: missing or invalid username")
		return model.NewAppErr("invalid username parameter", nil)
	}

	request := &model.DeleteUserRequest{
		Username: username,
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting user")
		return err
	}

	return helper.WriteJSON(w, http.StatusNoContent, nil)
}

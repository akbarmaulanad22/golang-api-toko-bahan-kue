package http

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
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

func (c *UserController) Register(w http.ResponseWriter, r *http.Request) {
	var request model.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to register user: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.UserResponse]{Data: response})
}

func (c *UserController) Current(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetUser(r)

	request := &model.GetUserRequest{
		Username: auth.Username,
	}

	response, err := c.UseCase.Current(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to get current user")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.UserResponse]{Data: response})
}

func (c *UserController) Login(w http.ResponseWriter, r *http.Request) {
	var request model.LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	response, err := c.UseCase.Login(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to login user: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.UserResponse]{Data: response})
}

func (c *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetUser(r)

	request := &model.LogoutUserRequest{
		Username: auth.Username,
	}

	response, err := c.UseCase.Logout(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to logout user")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[bool]{Data: response})
}

func (c *UserController) List(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	pageStr := params.Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	pageInt, _ := strconv.Atoi(pageStr)

	sizeStr := params.Get("size")
	if sizeStr == "" {
		sizeStr = "10"
	}
	sizeInt, _ := strconv.Atoi(sizeStr)

	roleIDStr := params.Get("role_id")
	if roleIDStr == "" {
		roleIDStr = "0"
	}
	roleIDInt, _ := strconv.Atoi(roleIDStr)

	branchIDStr := params.Get("branch_id")
	if branchIDStr == "" {
		branchIDStr = "0"
	}
	branchIDInt, _ := strconv.Atoi(branchIDStr)

	request := &model.SearchUserRequest{
		Search:   params.Get("search"),
		RoleID:   uint(roleIDInt),
		BranchID: uint(branchIDInt),
		Page:     pageInt,
		Size:     sizeInt,
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching user")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.UserResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *UserController) Get(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]

	if username == "" {
		http.Error(w, "invalid username", http.StatusBadRequest)
		return
	}

	request := &model.GetUserRequest{
		Username: username,
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting user")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.UserResponse]{Data: response})
}

func (c *UserController) Update(w http.ResponseWriter, r *http.Request) {

	username := mux.Vars(r)["username"]

	if username == "" {
		http.Error(w, "invalid username", http.StatusBadRequest)
		return
	}

	c.Log.Debug(username)

	request := new(model.UpdateUserRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	request.Username = username

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to update user")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.UserResponse]{Data: response})
}

func (c *UserController) Delete(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]

	if username == "" {
		http.Error(w, "invalid username", http.StatusBadRequest)
		return
	}

	request := &model.DeleteUserRequest{
		Username: username,
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting user")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[bool]{Data: true})
}

package http

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type CategoryController struct {
	Log     *logrus.Logger
	UseCase *usecase.CategoryUseCase
}

func NewCategoryController(useCase *usecase.CategoryUseCase, logger *logrus.Logger) *CategoryController {
	return &CategoryController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *CategoryController) Create(w http.ResponseWriter, r *http.Request) error {
	var request model.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error creating category")
		return err
	}

	return helper.WriteJSON(w, http.StatusCreated, model.WebResponse[*model.CategoryResponse]{Data: response})
}

func (c *CategoryController) List(w http.ResponseWriter, r *http.Request) error {
	params := r.URL.Query()

	page := helper.ParseIntOrDefault(params.Get("page"), 1)
	size := helper.ParseIntOrDefault(params.Get("size"), 10)

	request := &model.SearchTopSellerCategoryRequest{
		Name: params.Get("search"),
		Page: page,
		Size: size,
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching category")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.CategoryResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *CategoryController) Get(w http.ResponseWriter, r *http.Request) error {
	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return model.NewAppErr("invalid id parameter", nil)

	}
	request := &model.GetCategoryRequest{
		ID: uint(idInt),
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting category")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.CategoryResponse]{Data: response})
}

func (c *CategoryController) Update(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: %+v", err)
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := new(model.UpdateCategoryRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	request.ID = uint(idInt)

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("error to update category")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.CategoryResponse]{Data: response})
}

func (c *CategoryController) Delete(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: %+v", err)
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.DeleteCategoryRequest{
		ID: uint(idInt),
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting category")
		return err
	}

	return helper.WriteJSON(w, http.StatusNoContent, nil)
}

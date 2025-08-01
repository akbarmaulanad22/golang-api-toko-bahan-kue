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

func (c *CategoryController) Create(w http.ResponseWriter, r *http.Request) {
	var request model.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create category: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.CategoryResponse]{Data: response})
}

func (c *CategoryController) List(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	pageStr, ok := params["page"]
	if !ok || pageStr == "" {
		pageStr = "1"
	}

	pageInt, err := strconv.Atoi(pageStr)
	if err != nil {
		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
		return
	}

	sizeStr, ok := params["size"]
	if !ok || sizeStr == "" {
		sizeStr = "10"
	}

	sizeInt, err := strconv.Atoi(sizeStr)
	if err != nil {
		http.Error(w, "invalid size parameter", http.StatusBadRequest)
		return
	}

	request := &model.SearchCategoryRequest{
		Name: params["name"],
		Page: pageInt,
		Size: sizeInt,
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching category")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.CategoryResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *CategoryController) Get(w http.ResponseWriter, r *http.Request) {
	request := &model.GetCategoryRequest{
		Slug: mux.Vars(r)["slug"],
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting category")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.CategoryResponse]{Data: response})
}

func (c *CategoryController) Update(w http.ResponseWriter, r *http.Request) {

	request := new(model.UpdateCategoryRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	request.Slug = mux.Vars(r)["slug"]

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to update category")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.CategoryResponse]{Data: response})
}

func (c *CategoryController) Delete(w http.ResponseWriter, r *http.Request) {

	request := &model.DeleteCategoryRequest{
		Slug: mux.Vars(r)["slug"],
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting category")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[bool]{Data: true})
}

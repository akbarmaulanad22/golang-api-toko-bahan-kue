package http

import (
	"encoding/json"
	"math"
	"net/http"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type ProductController struct {
	Log     *logrus.Logger
	UseCase *usecase.ProductUseCase
}

func NewProductController(useCase *usecase.ProductUseCase, logger *logrus.Logger) *ProductController {
	return &ProductController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *ProductController) Create(w http.ResponseWriter, r *http.Request) error {
	var request model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("error to create product: %+v", err)
		return err
	}

	return helper.WriteJSON(w, http.StatusCreated, model.WebResponse[*model.ProductResponse]{Data: response})
}

func (c *ProductController) List(w http.ResponseWriter, r *http.Request) error {
	params := r.URL.Query()

	page := helper.ParseIntOrDefault(params.Get("page"), 1)
	size := helper.ParseIntOrDefault(params.Get("size"), 10)

	request := &model.SearchProductRequest{
		Search: params.Get("search"),
		Page:   page,
		Size:   size,
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching product")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.ProductResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *ProductController) Get(w http.ResponseWriter, r *http.Request) error {

	request := &model.GetProductRequest{
		SKU: mux.Vars(r)["sku"],
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting product")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.ProductResponse]{Data: response})
}

func (c *ProductController) Update(w http.ResponseWriter, r *http.Request) error {

	request := new(model.UpdateProductRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	request.SKU = mux.Vars(r)["sku"]

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("error to update product")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.ProductResponse]{Data: response})
}

func (c *ProductController) Delete(w http.ResponseWriter, r *http.Request) error {

	request := &model.DeleteProductRequest{
		SKU: mux.Vars(r)["sku"],
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting product")
		return err
	}

	return helper.WriteJSON(w, http.StatusNoContent, nil)
}

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

func (c *ProductController) Create(w http.ResponseWriter, r *http.Request) {
	var request model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create product: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.ProductResponse]{Data: response})
}

func (c *ProductController) List(w http.ResponseWriter, r *http.Request) {
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

	request := &model.SearchProductRequest{
		// Name: params["name"],
		// SKU:  params["sku"],
		Page: pageInt,
		Size: sizeInt,
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching product")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.ProductResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *ProductController) Get(w http.ResponseWriter, r *http.Request) {

	request := &model.GetProductRequest{
		SKU: mux.Vars(r)["sku"],
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting product")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.ProductResponse]{Data: response})
}

func (c *ProductController) Update(w http.ResponseWriter, r *http.Request) {

	request := new(model.UpdateProductRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	request.SKU = mux.Vars(r)["sku"]

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to update product")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.ProductResponse]{Data: response})
}

func (c *ProductController) Delete(w http.ResponseWriter, r *http.Request) {

	request := &model.DeleteProductRequest{
		SKU: mux.Vars(r)["sku"],
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting product")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[bool]{Data: true})
}

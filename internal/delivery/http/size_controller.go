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

type SizeController struct {
	Log     *logrus.Logger
	UseCase *usecase.SizeUseCase
}

func NewSizeController(useCase *usecase.SizeUseCase, logger *logrus.Logger) *SizeController {
	return &SizeController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *SizeController) Create(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	productSKU, ok := params["productSKU"]
	if !ok || productSKU == "" {
		http.Error(w, "Invalid product sku parameter", http.StatusBadRequest)
		return
	}

	var request model.CreateSizeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	request.ProductSKU = productSKU

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create size: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.SizeResponse]{Data: response})
}

func (c *SizeController) List(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	query := r.URL.Query()

	productSKU, ok := params["productSKU"]
	if !ok || productSKU == "" {
		http.Error(w, "Invalid product sku parameter", http.StatusBadRequest)
		return
	}

	pageStr := query.Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	pageInt, _ := strconv.Atoi(pageStr)

	sizeStr := query.Get("size")
	if sizeStr == "" {
		sizeStr = "10"
	}
	sizeInt, _ := strconv.Atoi(sizeStr)

	request := &model.SearchSizeRequest{
		Name:       query.Get("name"),
		ProductSKU: productSKU,
		Page:       pageInt,
		Size:       sizeInt,
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching size")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.SizeResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *SizeController) Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	productSKU, ok := params["productSKU"]
	if !ok || productSKU == "" {
		http.Error(w, "Invalid product sku parameter", http.StatusBadRequest)
		return
	}

	idStr, ok := params["id"]
	if !ok || idStr == "" {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	request := &model.GetSizeRequest{
		ID:         uint(idInt),
		ProductSKU: productSKU,
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting size")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.SizeResponse]{Data: response})
}

func (c *SizeController) Update(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	productSKU, ok := params["productSKU"]
	if !ok || productSKU == "" {
		http.Error(w, "Invalid product sku parameter", http.StatusBadRequest)
		return
	}

	idStr, ok := params["id"]
	if !ok || idStr == "" {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	request := new(model.UpdateSizeRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	request.ID = uint(idInt)
	request.ProductSKU = productSKU

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to update size")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.SizeResponse]{Data: response})
}

func (c *SizeController) Delete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	productSKU, ok := params["productSKU"]
	if !ok || productSKU == "" {
		http.Error(w, "Invalid product sku parameter", http.StatusBadRequest)
		return
	}

	idStr, ok := params["id"]
	if !ok || idStr == "" {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	request := &model.DeleteSizeRequest{
		ID:         uint(idInt),
		ProductSKU: productSKU,
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting size")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[bool]{Data: true})
}

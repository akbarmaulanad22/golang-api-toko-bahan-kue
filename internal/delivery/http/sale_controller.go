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

type SaleController struct {
	Log     *logrus.Logger
	UseCase *usecase.SaleUseCase
}

func NewSaleController(useCase *usecase.SaleUseCase, logger *logrus.Logger) *SaleController {
	return &SaleController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *SaleController) Create(w http.ResponseWriter, r *http.Request) {

	auth := middleware.GetUser(r)

	var request model.CreateSaleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	request.BranchID = auth.BranchID

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create sale: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.SaleResponse]{Data: response})
}

func (c *SaleController) List(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	pageInt, err := strconv.Atoi(params.Get("page"))
	if err != nil {
		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
		return
	}

	sizeInt, err := strconv.Atoi(params.Get("size"))
	if err != nil {
		http.Error(w, "Invalid size parameter", http.StatusBadRequest)
		return
	}

	startAtInt, err := strconv.Atoi(params.Get("start_at"))
	if err != nil {
		http.Error(w, "Invalid start at parameter", http.StatusBadRequest)
		return
	}

	endAtInt, err := strconv.Atoi(params.Get("end_at"))
	if err != nil {
		http.Error(w, "Invalid end at parameter", http.StatusBadRequest)
		return
	}

	request := &model.SearchSaleRequest{
		Code:         params.Get("code"),
		CustomerName: params.Get("customer_name"),
		Status:       model.StatusPayment(params.Get("status")),
		StartAt:      int64(startAtInt),
		EndAt:        int64(endAtInt),
		Page:         pageInt,
		Size:         sizeInt,
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching sale")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.SaleResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *SaleController) Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	code, ok := params["code"]
	if !ok || code == "" {
		http.Error(w, "Invalid product sku parameter", http.StatusBadRequest)
		return
	}

	request := &model.GetSaleRequest{
		Code: code,
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting sale")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.SaleResponse]{Data: response})
}

func (c *SaleController) Update(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	code, ok := params["code"]
	if !ok || code == "" {
		http.Error(w, "Invalid product sku parameter", http.StatusBadRequest)
		return
	}

	request := new(model.UpdateSaleRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	request.Code = code

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to update sale")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.SaleResponse]{Data: response})
}

package http

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
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

	if auth.BranchID != nil {
		request.BranchID = *auth.BranchID
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create sale: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.SaleResponse]{Data: response})
}

func (c *SaleController) List(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetUser(r)

	params := r.URL.Query()

	page := params.Get("page")
	if page == "" {
		page = "1"
	}
	pageInt, _ := strconv.Atoi(page)

	size := params.Get("size")
	if size == "" {
		size = "10"
	}
	sizeInt, _ := strconv.Atoi(size)

	startAt := params.Get("start_at")
	endAt := params.Get("end_at")

	var (
		startAtMili int64 = 0
		endAtMili   int64 = 0
	)

	if startAt != "" && endAt != "" {
		startAt, err := helper.ParseDateToMilli(startAt, false)
		if err != nil {
			c.Log.WithError(err).Error("invalid start at parameter")
			http.Error(w, err.Error(), helper.GetStatusCode(err))
			return
		}
		startAtMili = startAt

		endAt, err := helper.ParseDateToMilli(endAt, false)
		if err != nil {
			c.Log.WithError(err).Error("invalid end at parameter")
			http.Error(w, err.Error(), helper.GetStatusCode(err))
			return
		}
		endAtMili = endAt
	}

	request := &model.SearchSaleRequest{
		Search:  params.Get("search"),
		Status:  params.Get("status"),
		StartAt: startAtMili,
		EndAt:   endAtMili,
		Page:    pageInt,
		Size:    sizeInt,
	}

	if strings.ToUpper(auth.Role) == "OWNER" {
		branchID := params.Get("branch_id")
		if branchID != "" {
			branchIDInt, err := strconv.Atoi(branchID)
			if err != nil {
				c.Log.WithError(err).Error("invalid branch id parameter")
				http.Error(w, err.Error(), helper.GetStatusCode(err))
				return
			}
			branchIDUint := uint(branchIDInt)
			request.BranchID = &branchIDUint
		}

	} else {
		request.BranchID = auth.BranchID
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
		http.Error(w, "Invalid sale code parameter", http.StatusBadRequest)
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

// func (c *SaleController) Cancel(w http.ResponseWriter, r *http.Request) {

// 	params := mux.Vars(r)

// 	code, ok := params["code"]
// 	if !ok || code == "" {
// 		http.Error(w, "Invalid sale code parameter", http.StatusBadRequest)
// 		return
// 	}

// 	request := model.CancelSaleRequest{
// 		Code: code,
// 	}

// 	response, err := c.UseCase.Cancel(r.Context(), &request)
// 	if err != nil {
// 		c.Log.WithError(err).Warnf("Failed to cancel sale")
// 		http.Error(w, err.Error(), helper.GetStatusCode(err))
// 		return
// 	}

// 	json.NewEncoder(w).Encode(model.WebResponse[*model.SaleResponse]{Data: response})
// }

// func (c *SaleController) ListReport(w http.ResponseWriter, r *http.Request) {
// 	params := r.URL.Query()

// 	page := params.Get("page")
// 	if page == "" {
// 		page = "1"
// 	}

// 	pageInt, err := strconv.Atoi(page)
// 	if err != nil {
// 		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
// 		return
// 	}

// 	size := params.Get("size")
// 	if size == "" {
// 		size = "10"
// 	}

// 	sizeInt, err := strconv.Atoi(size)
// 	if err != nil {
// 		http.Error(w, "Invalid size parameter", http.StatusBadRequest)
// 		return
// 	}

// 	branchID := params.Get("branch_id")
// 	if branchID == "" {
// 		branchID = "0"
// 	}

// 	branchIDInt, err := strconv.Atoi(branchID)
// 	if err != nil {
// 		http.Error(w, "Invalid branch id parameter", http.StatusBadRequest)
// 		return
// 	}

// 	request := &model.SearchSaleReportRequest{
// 		BranchID: uint(branchIDInt),
// 		Search:   params.Get("search"),
// 		StartAt:  params.Get("start_at"),
// 		EndAt:    params.Get("end_at"),
// 		Page:     pageInt,
// 		Size:     sizeInt,
// 	}

// 	responses, total, err := c.UseCase.SearchReports(r.Context(), request)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error searching sales report")
// 		http.Error(w, err.Error(), helper.GetStatusCode(err))
// 		return
// 	}

// 	paging := &model.PageMetadata{
// 		Page:      request.Page,
// 		Size:      request.Size,
// 		TotalItem: total,
// 		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
// 	}

// 	json.NewEncoder(w).Encode(model.WebResponse[[]model.SaleReportResponse]{
// 		Data:   responses,
// 		Paging: paging,
// 	})
// }

// func (c *SaleController) ListBranchSaleReport(w http.ResponseWriter, r *http.Request) {

// 	response, err := c.UseCase.GetBranchSalesReport(r.Context())
// 	if err != nil {
// 		c.Log.WithError(err).Error("error getting sale")
// 		http.Error(w, err.Error(), helper.GetStatusCode(err))
// 		return
// 	}

// 	json.NewEncoder(w).Encode(model.WebResponse[[]model.BranchSalesReportResponse]{Data: response})
// }

// func (c *SaleController) ListBestSellingProduct(w http.ResponseWriter, r *http.Request) {

// 	params := r.URL.Query()

// 	branchID := params.Get("branch_id")
// 	if branchID == "" {
// 		branchID = "0"
// 	}

// 	branchIDInt, _ := strconv.Atoi(branchID)

// 	request := &model.ListBestSellingProductRequest{
// 		BranchID: uint(branchIDInt),
// 	}

// 	if request.BranchID != 0 {
// 		response, err := c.UseCase.ListBestSellingProductByBranchID(r.Context(), request)
// 		if err != nil {
// 			c.Log.WithError(err).Error("error getting best seller products")
// 			http.Error(w, err.Error(), helper.GetStatusCode(err))
// 			return
// 		}

// 		json.NewEncoder(w).Encode(model.WebResponse[[]model.BestSellingProductResponse]{Data: response})
// 		return
// 	}

// 	response, err := c.UseCase.ListBestSellingProductGlobal(r.Context())
// 	if err != nil {
// 		c.Log.WithError(err).Error("error getting best seller products")
// 		http.Error(w, err.Error(), helper.GetStatusCode(err))
// 		return
// 	}

// 	json.NewEncoder(w).Encode(model.WebResponse[[]model.BestSellingProductResponse]{Data: response})
// }

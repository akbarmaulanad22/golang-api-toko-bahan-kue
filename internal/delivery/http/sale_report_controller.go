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

	"github.com/sirupsen/logrus"
)

type SaleReportController struct {
	Log     *logrus.Logger
	UseCase *usecase.SaleReportUseCase
}

func NewSaleReportController(useCase *usecase.SaleReportUseCase, logger *logrus.Logger) *SaleReportController {
	return &SaleReportController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *SaleReportController) ListDaily(w http.ResponseWriter, r *http.Request) {

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

	startAtStr := params.Get("start_at")
	endAtStr := params.Get("end_at")

	var (
		startAtMili int64
		endAtMili   int64
	)

	if startAtStr != "" && endAtStr != "" {
		startMilli, err := helper.ParseDateToMilli(startAtStr, false)
		if err != nil {
			http.Error(w, "Invalid start_at format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		startAtMili = startMilli

		endMilli, err := helper.ParseDateToMilli(endAtStr, true)
		if err != nil {
			http.Error(w, "Invalid start_at format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		endAtMili = endMilli
	}

	request := &model.SearchSalesReportRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
		Page:    pageInt,
		Size:    sizeInt,
	}

	auth := middleware.GetUser(r)
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

	responses, total, err := c.UseCase.SearchDaily(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching daily sales report")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.SalesDailyReportResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *SaleReportController) ListTopSeller(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()

	page := params.Get("page")
	if page == "" {
		page = "1"
	}
	pageInt, _ := strconv.Atoi(page)

	size := params.Get("size")
	if size == "" {
		size = "5"
	}
	sizeInt, _ := strconv.Atoi(size)

	startAtStr := params.Get("start_at")
	endAtStr := params.Get("end_at")

	var (
		startAtMili int64
		endAtMili   int64
	)

	if startAtStr != "" && endAtStr != "" {
		startMilli, err := helper.ParseDateToMilli(startAtStr, false)
		if err != nil {
			http.Error(w, "Invalid start_at format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		startAtMili = startMilli

		endMilli, err := helper.ParseDateToMilli(endAtStr, true)
		if err != nil {
			http.Error(w, "Invalid start_at format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		endAtMili = endMilli
	}

	request := &model.SearchSalesReportRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
		Page:    pageInt,
		Size:    sizeInt,
	}

	auth := middleware.GetUser(r)
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

	responses, total, err := c.UseCase.SearchTopSellerProduct(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching top seller sales report")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.SalesTopSellerReportResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *SaleReportController) ListCategory(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()

	page := params.Get("page")
	if page == "" {
		page = "1"
	}
	pageInt, _ := strconv.Atoi(page)

	size := params.Get("size")
	if size == "" {
		size = "5"
	}
	sizeInt, _ := strconv.Atoi(size)

	startAtStr := params.Get("start_at")
	endAtStr := params.Get("end_at")

	var (
		startAtMili int64
		endAtMili   int64
	)

	if startAtStr != "" && endAtStr != "" {
		startMilli, err := helper.ParseDateToMilli(startAtStr, false)
		if err != nil {
			http.Error(w, "Invalid start_at format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		startAtMili = startMilli

		endMilli, err := helper.ParseDateToMilli(endAtStr, true)
		if err != nil {
			http.Error(w, "Invalid start_at format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		endAtMili = endMilli
	}

	request := &model.SearchSalesReportRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
		Page:    pageInt,
		Size:    sizeInt,
	}

	auth := middleware.GetUser(r)
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

	responses, total, err := c.UseCase.SearchTopSellerCategory(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching top seller sales report")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.SalesCategoryResponse]{
		Data:   responses,
		Paging: paging,
	})
}

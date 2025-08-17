package http

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
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

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
		return
	}

	size := params.Get("size")
	if size == "" {
		size = "10"
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		http.Error(w, "Invalid size parameter", http.StatusBadRequest)
		return
	}

	var branchID *uint
	branchIDStr := params.Get("branch_id")
	if branchIDStr != "" {
		idInt, err := strconv.Atoi(branchIDStr)
		if err != nil {
			http.Error(w, "Invalid branch id parameter", http.StatusBadRequest)
			return
		}
		idUint := uint(idInt)
		branchID = &idUint
	}

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
		BranchID: branchID,
		StartAt:  startAtMili,
		EndAt:    endAtMili,
		Page:     pageInt,
		Size:     sizeInt,
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

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
		return
	}

	size := params.Get("size")
	if size == "" {
		size = "10"
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		http.Error(w, "Invalid size parameter", http.StatusBadRequest)
		return
	}

	var branchID *uint
	branchIDStr := params.Get("branch_id")
	if branchIDStr != "" {
		idInt, err := strconv.Atoi(branchIDStr)
		if err != nil {
			http.Error(w, "Invalid branch id parameter", http.StatusBadRequest)
			return
		}
		idUint := uint(idInt)
		branchID = &idUint
	}

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
		BranchID: branchID,
		StartAt:  startAtMili,
		EndAt:    endAtMili,
		Page:     pageInt,
		Size:     sizeInt,
	}

	responses, total, err := c.UseCase.SearchTopSeller(r.Context(), request)
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

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
		return
	}

	size := params.Get("size")
	if size == "" {
		size = "10"
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		http.Error(w, "Invalid size parameter", http.StatusBadRequest)
		return
	}

	var branchID *uint
	branchIDStr := params.Get("branch_id")
	if branchIDStr != "" {
		idInt, err := strconv.Atoi(branchIDStr)
		if err != nil {
			http.Error(w, "Invalid branch id parameter", http.StatusBadRequest)
			return
		}
		idUint := uint(idInt)
		branchID = &idUint
	}

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
		BranchID: branchID,
		StartAt:  startAtMili,
		EndAt:    endAtMili,
		Page:     pageInt,
		Size:     sizeInt,
	}

	responses, total, err := c.UseCase.SearchCategory(r.Context(), request)
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

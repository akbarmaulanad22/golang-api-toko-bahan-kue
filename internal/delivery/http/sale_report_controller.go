package http

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"
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
		idUint := uint(idInt) // konversi ke uint
		branchID = &idUint    // ambil alamat untuk pointer
	}

	// Parse start_at dan end_at dari query string
	startAtStr := params.Get("start_at")
	endAtStr := params.Get("end_at")

	layout := "2006-01-02" // format tanggal, misal "2025-08-16"
	var startAt, endAt time.Time

	if startAtStr != "" {
		startAt, err = time.Parse(layout, startAtStr)
		if err != nil {
			http.Error(w, "Invalid start_at format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	if endAtStr != "" {
		endAt, err = time.Parse(layout, endAtStr)
		if err != nil {
			http.Error(w, "Invalid end_at format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	request := &model.SearchSalesDailyReportRequest{
		BranchID: branchID,
		Search:   params.Get("search"),
		StartAt:  startAt,
		EndAt:    endAt,
		Page:     pageInt,
		Size:     sizeInt,
	}

	responses, total, err := c.UseCase.SearchDaily(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching daily reports")
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

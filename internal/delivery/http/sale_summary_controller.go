package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type SaleSummaryController struct {
	Log     *logrus.Logger
	UseCase *usecase.SaleSummaryUseCase
}

func NewSaleSummaryController(useCase *usecase.SaleSummaryUseCase, logger *logrus.Logger) *SaleSummaryController {
	return &SaleSummaryController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *SaleSummaryController) ListBranch(w http.ResponseWriter, r *http.Request) {

	responses, err := c.UseCase.BranchSalesSummary(r.Context())
	if err != nil {
		c.Log.WithError(err).Error("error get sale summary")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.BranchSalesSummaryResponse]{
		Data: responses,
	})
}

func (c *SaleSummaryController) ListDailySaleSummary(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	branchIDStr, ok := params["branchID"]
	if !ok || branchIDStr == "" {
		http.Error(w, "Invalid branch id parameter", http.StatusBadRequest)
		return
	}

	branchIDInt, err := strconv.Atoi(branchIDStr)
	if err != nil {
		http.Error(w, "Invalid branch id parameter", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page == "" {
		page = "1"
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
		return
	}

	size := query.Get("size")
	if size == "" {
		size = "10"
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		http.Error(w, "Invalid size parameter", http.StatusBadRequest)
		return
	}

	request := &model.ListDailySalesSummaryRequest{
		BranchID: uint(branchIDInt),
		StartAt:  query.Get("start_at"),
		EndAt:    query.Get("end_at"),
		Page:     pageInt,
		Size:     sizeInt,
	}

	responses, err := c.UseCase.DailySalesSummaryByBranchID(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error get daily sale summary")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.DailySalesSummaryResponse]{
		Data: responses,
	})
}

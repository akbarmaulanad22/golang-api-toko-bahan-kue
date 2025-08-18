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

type ExpenseController struct {
	Log     *logrus.Logger
	UseCase *usecase.ExpenseUseCase
}

func NewExpenseController(useCase *usecase.ExpenseUseCase, logger *logrus.Logger) *ExpenseController {
	return &ExpenseController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *ExpenseController) Create(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetUser(r)

	var request model.CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	request.BranchID = auth.BranchID

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create expense: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.ExpenseResponse]{Data: response})
}

func (c *ExpenseController) List(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	auth := middleware.GetUser(r)

	var branchID *uint

	if auth.Role == "Owner" {
		branchIDStr := params.Get("branch_id")
		if branchIDStr != "" {
			branchIDInt, err := strconv.Atoi(branchIDStr)
			if err != nil {
				http.Error(w, "Invalid branch ID parameter", http.StatusBadRequest)
				return
			}
			tmp := uint(branchIDInt)
			branchID = &tmp
		} else {
			branchID = nil // nil artinya semua cabang
		}
	} else {
		tmp := auth.BranchID
		branchID = &tmp
	}

	c.Log.Warnf("BRANCH ID = %d", branchID)

	pageStr := params.Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	pageInt, err := strconv.Atoi(pageStr)
	if err != nil {
		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
		return
	}

	sizeStr := params.Get("size")
	if sizeStr == "" {
		sizeStr = "10"
	}

	sizeInt, err := strconv.Atoi(sizeStr)
	if err != nil {
		http.Error(w, "invalid size parameter", http.StatusBadRequest)
		return
	}

	request := &model.SearchExpenseRequest{
		BranchID:    branchID,
		Description: params.Get("description"),
		Page:        pageInt,
		Size:        sizeInt,
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching expense")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.ExpenseResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *ExpenseController) Update(w http.ResponseWriter, r *http.Request) {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	request := new(model.UpdateExpenseRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	request.ID = uint(idInt)

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to update expense")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.ExpenseResponse]{Data: response})
}

func (c *ExpenseController) Delete(w http.ResponseWriter, r *http.Request) {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	request := &model.DeleteExpenseRequest{
		ID: uint(idInt),
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting expense")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[bool]{Data: true})
}

func (c *ExpenseController) ConsolidatedReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	startStr := params["start_at"] // contoh: "2022-02-02"
	endStr := params["end_at"]     // contoh: "2022-02-02"

	var startAt int64
	if startStr != "" {
		startDate, err := helper.ParseDateToMilli(startStr, false)
		if err != nil {
			http.Error(w, "invalid start_at format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		startAt = startDate
	}

	var endAt int64
	if endStr != "" {
		endDate, err := helper.ParseDateToMilli(endStr, false)
		if err != nil {
			http.Error(w, "invalid end_at format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		endAt = endDate
	}

	request := &model.SearchConsolidateExpenseRequest{
		StartAt: startAt,
		EndAt:   endAt,
	}

	responses, err := c.UseCase.ConsolidateReport(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching expense")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(responses)
}

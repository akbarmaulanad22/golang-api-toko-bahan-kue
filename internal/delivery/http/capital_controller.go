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

type CapitalController struct {
	Log     *logrus.Logger
	UseCase *usecase.CapitalUseCase
}

func NewCapitalController(useCase *usecase.CapitalUseCase, logger *logrus.Logger) *CapitalController {
	return &CapitalController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *CapitalController) Create(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetUser(r)

	var request model.CreateCapitalRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	request.BranchID = auth.BranchID

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create capital: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.CapitalResponse]{Data: response})
}

func (c *CapitalController) List(w http.ResponseWriter, r *http.Request) {
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

	request := &model.SearchCapitalRequest{
		BranchID: branchID,
		Note:     params.Get("note"),
		Page:     pageInt,
		Size:     sizeInt,
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching capital")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.CapitalResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *CapitalController) Update(w http.ResponseWriter, r *http.Request) {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	request := new(model.UpdateCapitalRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	request.ID = uint(idInt)

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to update capital")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.CapitalResponse]{Data: response})
}

func (c *CapitalController) Delete(w http.ResponseWriter, r *http.Request) {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	request := &model.DeleteCapitalRequest{
		ID: uint(idInt),
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting capital")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[bool]{Data: true})
}

// func (c *CapitalController) ConsolidatedReport(w http.ResponseWriter, r *http.Request) {
// 	params := mux.Vars(r)

// 	startStr := params["start_at"] // contoh: "2022-02-02"
// 	endStr := params["end_at"]     // contoh: "2022-02-02"

// 	var startAt int64
// 	if startStr != "" {
// 		startDate, err := helper.ParseDateToMilli(startStr, false)
// 		if err != nil {
// 			http.Error(w, "invalid start_at format, expected YYYY-MM-DD", http.StatusBadRequest)
// 			return
// 		}
// 		startAt = startDate
// 	}

// 	var endAt int64
// 	if endStr != "" {
// 		endDate, err := helper.ParseDateToMilli(endStr, false)
// 		if err != nil {
// 			http.Error(w, "invalid end_at format, expected YYYY-MM-DD", http.StatusBadRequest)
// 			return
// 		}
// 		endAt = endDate
// 	}

// 	request := &model.SearchConsolidateCapitalRequest{
// 		StartAt: startAt,
// 		EndAt:   endAt,
// 	}

// 	responses, err := c.UseCase.ConsolidateReport(r.Context(), request)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error searching capital")
// 		http.Error(w, err.Error(), helper.GetStatusCode(err))
// 		return
// 	}

// 	json.NewEncoder(w).Encode(responses)
// }

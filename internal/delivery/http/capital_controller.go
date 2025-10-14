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

func (c *CapitalController) Create(w http.ResponseWriter, r *http.Request) error {

	var request model.CreateCapitalRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	auth := middleware.GetUser(r)
	request.BranchID = auth.BranchID

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create capital: %+v", err)
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.CapitalResponse]{Data: response})
}

func (c *CapitalController) List(w http.ResponseWriter, r *http.Request) error {
	params := r.URL.Query()

	page := helper.ParseIntOrDefault(params.Get("page"), 1)
	size := helper.ParseIntOrDefault(params.Get("size"), 10)

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
			return model.NewAppErr("invalid start at parameter", nil)
		}
		startAtMili = startAt

		endAt, err := helper.ParseDateToMilli(endAt, true)
		if err != nil {
			c.Log.WithError(err).Error("invalid end at parameter")
			return model.NewAppErr("invalid end at parameter", nil)
		}
		endAtMili = endAt
	}

	request := &model.SearchCapitalRequest{
		Note:    params.Get("search"),
		Page:    page,
		Size:    size,
		StartAt: startAtMili,
		EndAt:   endAtMili,
	}

	branchID := params.Get("branch_id")
	auth := middleware.GetUser(r)
	if strings.ToUpper(auth.Role) == "OWNER" && branchID != "" {
		branchIDInt, err := strconv.Atoi(branchID)
		if err != nil {
			c.Log.WithError(err).Error("invalid branch id parameter")
			return model.NewAppErr("invalid branch id parameter", nil)
		}
		branchIDUint := uint(branchIDInt)
		request.BranchID = &branchIDUint
	} else {
		request.BranchID = auth.BranchID
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching capital")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.CapitalResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *CapitalController) Update(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := new(model.UpdateCapitalRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	auth := middleware.GetUser(r)
	request.ID = uint(idInt)
	request.BranchID = auth.BranchID

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to update capital")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.CapitalResponse]{Data: response})
}

func (c *CapitalController) Delete(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.DeleteCapitalRequest{
		ID: uint(idInt),
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting capital")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[bool]{Data: true})
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

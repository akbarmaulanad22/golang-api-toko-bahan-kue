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

func (c *ExpenseController) Create(w http.ResponseWriter, r *http.Request) error {

	var request model.CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	auth := middleware.GetUser(r)
	request.BranchID = auth.BranchID

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error creating expense")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.ExpenseResponse]{Data: response})
}

func (c *ExpenseController) List(w http.ResponseWriter, r *http.Request) error {
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
			c.Log.Warnf("error to parse start at parameter: %+v", err)
			return model.NewAppErr("invalid start at parameter", nil)
		}
		startAtMili = startAt

		endAt, err := helper.ParseDateToMilli(endAt, true)
		if err != nil {
			c.Log.Warnf("error to parse end at parameter: %+v", err)
			return model.NewAppErr("invalid end at parameter", nil)
		}
		endAtMili = endAt
	}

	request := &model.SearchExpenseRequest{
		Search:  params.Get("search"),
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
			c.Log.Warnf("error to parse branch id parameter: %+v", err)
			return model.NewAppErr("invalid branch id parameter", nil)
		}
		branchIDUint := uint(branchIDInt)
		request.BranchID = &branchIDUint
	} else {
		request.BranchID = auth.BranchID
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching expense")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.ExpenseResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *ExpenseController) Update(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: %+v", err)
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := new(model.UpdateExpenseRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	request.ID = uint(idInt)

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("error to update expense")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.ExpenseResponse]{Data: response})
}

func (c *ExpenseController) Delete(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: %+v", err)
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.DeleteExpenseRequest{
		ID: uint(idInt),
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting expense")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[bool]{Data: true})
}

func (c *ExpenseController) ConsolidatedReport(w http.ResponseWriter, r *http.Request) error {
	params := r.URL.Query()

	startAt := params.Get("start_at")
	endAt := params.Get("end_at")

	var (
		startAtMili int64 = 0
		endAtMili   int64 = 0
	)

	if startAt != "" && endAt != "" {
		startAt, err := helper.ParseDateToMilli(startAt, false)
		if err != nil {
			c.Log.Warnf("error to parse start at parameter: %+v", err)
			return model.NewAppErr("invalid start at parameter", nil)
		}
		startAtMili = startAt

		endAt, err := helper.ParseDateToMilli(endAt, true)
		if err != nil {
			c.Log.Warnf("error to parse end at parameter: %+v", err)
			return model.NewAppErr("invalid end at parameter", nil)
		}
		endAtMili = endAt
	}

	request := &model.SearchConsolidateExpenseRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
	}

	responses, err := c.UseCase.ConsolidateReport(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching expense")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, responses)
}

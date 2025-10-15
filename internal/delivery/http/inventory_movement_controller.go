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

type InventoryMovementController struct {
	Log     *logrus.Logger
	UseCase *usecase.InventoryMovementUseCase
}

func NewInventoryMovementController(useCase *usecase.InventoryMovementUseCase, logger *logrus.Logger) *InventoryMovementController {
	return &InventoryMovementController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *InventoryMovementController) Create(w http.ResponseWriter, r *http.Request) error {
	var request model.BulkCreateInventoryMovementRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	auth := middleware.GetUser(r)
	if auth.BranchID != nil {
		request.BranchID = *auth.BranchID
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error creating inventory movement")
		return err
	}

	return helper.WriteJSON(w, http.StatusCreated, model.WebResponse[*model.BulkInventoryMovementResponse]{Data: response})
}

func (c *InventoryMovementController) CreateStockOpname(w http.ResponseWriter, r *http.Request) error {
	var request model.BulkCreateInventoryMovementRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	request.ReferenceType = "ADJUST"

	auth := middleware.GetUser(r)
	if auth.BranchID != nil {
		request.BranchID = *auth.BranchID
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error creating inventory movement")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.BulkInventoryMovementResponse]{Data: response})
}

func (c *InventoryMovementController) List(w http.ResponseWriter, r *http.Request) error {
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

	request := &model.SearchInventoryMovementRequest{
		Page:    page,
		Size:    size,
		Type:    params.Get("type"),
		Search:  params.Get("search"),
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
		c.Log.WithError(err).Error("error searching inventory movements")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.InventoryMovementResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *InventoryMovementController) Summary(w http.ResponseWriter, r *http.Request) error {

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

	request := &model.SearchInventoryMovementRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
	}

	response, err := c.UseCase.Summary(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting summary inventory movement")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, response)
}

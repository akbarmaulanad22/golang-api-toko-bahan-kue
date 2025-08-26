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

func (c *InventoryMovementController) Create(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetUser(r)

	var request model.BulkCreateInventoryMovementRequest
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

	json.NewEncoder(w).Encode(model.WebResponse[*model.BulkInventoryMovementResponse]{Data: response})
}

func (c *InventoryMovementController) List(w http.ResponseWriter, r *http.Request) {
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

	request := &model.SearchInventoryMovementRequest{
		BranchID: branchID,
		Page:     pageInt,
		Size:     sizeInt,
		Type:     params.Get("type"),
		Search:   params.Get("search"),
		StartAt:  startAtMili,
		EndAt:    endAtMili,
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching inventory movements")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.InventoryMovementResponse]{
		Data:   responses,
		Paging: paging,
	})
}

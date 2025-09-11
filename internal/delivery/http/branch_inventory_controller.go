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

type BranchInventoryController struct {
	Log     *logrus.Logger
	UseCase *usecase.BranchInventoryUseCase
}

func NewBranchInventoryController(useCase *usecase.BranchInventoryUseCase, logger *logrus.Logger) *BranchInventoryController {
	return &BranchInventoryController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *BranchInventoryController) List(w http.ResponseWriter, r *http.Request) {

	auth := middleware.GetUser(r)

	// if auth.Role == "Owner" {
	// 	responses, err := c.UseCase.ListOwnerInventoryByBranch(r.Context())
	// 	if err != nil {
	// 		c.Log.WithError(err).Error("error searching branch")
	// 		http.Error(w, err.Error(), helper.GetStatusCode(err))
	// 		return
	// 	}

	// 	json.NewEncoder(w).Encode(model.WebResponse[[]model.BranchInventoryResponse]{
	// 		Data: responses,
	// 	})

	// 	return
	// }

	params := r.URL.Query()

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

	request := &model.SearchBranchInventoryRequest{
		BranchID: auth.BranchID,
		Search:   params.Get("search"),
		Page:     pageInt,
		Size:     sizeInt,
	}

	responses, total, err := c.UseCase.List(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching branch inventory")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.BranchInventoryProductResponse]{
		Data:   responses,
		Paging: paging,
	})
}

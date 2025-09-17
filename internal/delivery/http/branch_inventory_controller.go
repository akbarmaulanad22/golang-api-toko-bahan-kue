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

	params := r.URL.Query()

	pageStr := params.Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	pageInt, _ := strconv.Atoi(pageStr)

	sizeStr := params.Get("size")
	if sizeStr == "" {
		sizeStr = "10"
	}
	sizeInt, _ := strconv.Atoi(sizeStr)

	request := &model.SearchBranchInventoryRequest{
		Search: params.Get("search"),
		Page:   pageInt,
		Size:   sizeInt,
	}

	auth := middleware.GetUser(r)
	if strings.ToUpper(auth.Role) == "OWNER" {
		branchID := params.Get("branch_id")
		if branchID != "" {
			branchIDInt, err := strconv.Atoi(branchID)
			if err != nil {
				c.Log.WithError(err).Error("invalid branch id parameter")
				http.Error(w, err.Error(), helper.GetStatusCode(err))
				return
			}
			branchIDUint := uint(branchIDInt)
			request.BranchID = &branchIDUint
		}

	} else {
		request.BranchID = auth.BranchID
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

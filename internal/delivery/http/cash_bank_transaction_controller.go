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

type CashBankTransactionController struct {
	Log     *logrus.Logger
	UseCase *usecase.CashBankTransactionUseCase
}

func NewCashBankTransactionController(useCase *usecase.CashBankTransactionUseCase, logger *logrus.Logger) *CashBankTransactionController {
	return &CashBankTransactionController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *CashBankTransactionController) List(w http.ResponseWriter, r *http.Request) {
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

	request := &model.SearchCashBankTransactionRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
		Page:    pageInt,
		Size:    sizeInt,
		Search:  params.Get("search"),
		Type:    params.Get("type"),
		Source:  params.Get("source"),
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

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching cash bank transaction")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.CashBankTransactionResponse]{
		Data:   responses,
		Paging: paging,
	})
}

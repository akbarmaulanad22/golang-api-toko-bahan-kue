package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tokobahankue/internal/delivery/http/middleware"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/usecase"

	"github.com/sirupsen/logrus"
)

type FinanceController struct {
	Log     *logrus.Logger
	UseCase *usecase.FinanceUseCase
}

func NewFinanceController(useCase *usecase.FinanceUseCase, logger *logrus.Logger) *FinanceController {
	return &FinanceController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *FinanceController) GetSummary(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()

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

	request := model.GetFinanceSummaryOwnerRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
	}

	response, err := c.UseCase.GetOwnerSummary(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error getting finance summary")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.FinanceSummaryOwnerResponse]{Data: response})
}

func (c *FinanceController) GetProfitLoss(w http.ResponseWriter, r *http.Request) {

	auth := middleware.GetUser(r)

	params := r.URL.Query()

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

	request := model.GetFinanceBasicRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
		// BranchID: auth.BranchID,
		Role: auth.Role,
	}

	response, err := c.UseCase.GetProfitLoss(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error getting finance profit loss")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.FinanceProfitLossResponse]{Data: response})
}

func (c *FinanceController) GetCashFlow(w http.ResponseWriter, r *http.Request) {

	auth := middleware.GetUser(r)

	params := r.URL.Query()

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

	request := model.GetFinanceBasicRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
		// BranchID: auth.BranchID,
		Role: auth.Role,
	}

	response, err := c.UseCase.GetCashFlow(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error getting finance cash flow")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.FinanceCashFlowResponse]{Data: response})
}

func (c *FinanceController) GetBalanceSheet(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()

	asOfStr := params.Get("as_of")
	if asOfStr == "" {
		asOfStr = time.Now().Format("2006-01-02")
	}

	asOfMilli, err := helper.ParseDateToMilli(asOfStr, false)
	if err != nil {
		http.Error(w, "Invalid as_of format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	auth := middleware.GetUser(r)
	request := model.GetFinanceBalanceSheetRequest{
		AsOf: asOfMilli,
		Role: auth.Role,
	}

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

	response, err := c.UseCase.GetBalanceSheet(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error getting finance balance sheet")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.FinanceBalanceSheetResponse]{Data: response})
}

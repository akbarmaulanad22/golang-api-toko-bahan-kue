package http

import (
	"encoding/json"
	"net/http"
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
		StartAt:  startAtMili,
		EndAt:    endAtMili,
		BranchID: auth.BranchID,
		Role:     auth.Role,
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
		StartAt:  startAtMili,
		EndAt:    endAtMili,
		BranchID: auth.BranchID,
		Role:     auth.Role,
	}

	response, err := c.UseCase.GetCashFlow(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error getting finance cash flow")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.FinanceCashFlowResponse]{Data: response})
}

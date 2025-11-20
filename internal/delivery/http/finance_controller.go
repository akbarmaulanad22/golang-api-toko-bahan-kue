package http

import (
	"net/http"
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

func (c *FinanceController) GetSummary(w http.ResponseWriter, r *http.Request) error {

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

	request := model.GetFinanceSummaryOwnerRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
	}

	response, err := c.UseCase.GetOwnerSummary(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error getting finance summary")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.FinanceSummaryOwnerResponse]{Data: response})
}

func (c *FinanceController) GetProfitLoss(w http.ResponseWriter, r *http.Request) error {

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

	auth := middleware.GetUser(r)
	request := model.GetFinanceBasicRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
		Role:    strings.ToUpper(auth.Role),
	}

	if strings.ToUpper(auth.Role) != "OWNER" {
		request.BranchID = *auth.BranchID
	}

	response, err := c.UseCase.GetProfitLoss(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error getting finance profit loss")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.FinanceProfitLossResponse]{Data: response})
}

func (c *FinanceController) GetCashFlow(w http.ResponseWriter, r *http.Request) error {

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

	auth := middleware.GetUser(r)
	request := model.GetFinanceBasicRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
		Role:    strings.ToUpper(auth.Role),
	}

	if strings.ToUpper(auth.Role) != "OWNER" {
		request.BranchID = *auth.BranchID
	}

	response, err := c.UseCase.GetCashFlow(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error getting finance cash flow")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.FinanceCashFlowResponse]{Data: response})
}

func (c *FinanceController) GetBalanceSheet(w http.ResponseWriter, r *http.Request) error {

	params := r.URL.Query()

	asOfStr := params.Get("as_of")
	if asOfStr == "" {
		asOfStr = time.Now().Format("2006-01-02")
	}

	asOfMilli, err := helper.ParseDateToMilli(asOfStr, true)
	if err != nil {
		c.Log.Warnf("error to parse as of parameter: %+v", err)
		return model.NewAppErr("invalid as of parameter. Use YYYY-MM-DD", nil)
	}

	auth := middleware.GetUser(r)
	request := model.GetFinanceBalanceSheetRequest{
		AsOf: asOfMilli,
		Role: strings.ToUpper(auth.Role),
	}

	if strings.ToUpper(auth.Role) != "OWNER" {
		request.BranchID = auth.BranchID
	}

	response, err := c.UseCase.GetBalanceSheet(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error getting finance balance sheet")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.FinanceBalanceSheetResponse]{Data: response})
}

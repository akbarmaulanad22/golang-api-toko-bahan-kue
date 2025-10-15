package http

import (
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

func (c *CashBankTransactionController) List(w http.ResponseWriter, r *http.Request) error {
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

	request := &model.SearchCashBankTransactionRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
		Page:    page,
		Size:    size,
		Search:  params.Get("search"),
		Type:    params.Get("type"),
		Source:  params.Get("source"),
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
		c.Log.WithError(err).Error("error searching cash bank transaction")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.CashBankTransactionResponse]{
		Data:   responses,
		Paging: paging,
	})
}

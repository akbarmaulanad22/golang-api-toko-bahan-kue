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

type PurchaseReportController struct {
	Log     *logrus.Logger
	UseCase *usecase.PurchaseReportUseCase
}

func NewPurchaseReportController(useCase *usecase.PurchaseReportUseCase, logger *logrus.Logger) *PurchaseReportController {
	return &PurchaseReportController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *PurchaseReportController) ListDaily(w http.ResponseWriter, r *http.Request) error {

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
			c.Log.WithError(err).Error("invalid start at parameter")
			return model.NewAppErr("invalid start at parameter", nil)
		}
		startAtMili = startAt

		endAt, err := helper.ParseDateToMilli(endAt, true)
		if err != nil {
			c.Log.WithError(err).Error("invalid end at parameter")
			return model.NewAppErr("invalid end at parameter", nil)
		}
		endAtMili = endAt
	}
	request := &model.SearchPurchasesReportRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
		Page:    pageInt,
		Size:    sizeInt,
	}

	branchID := params.Get("branch_id")
	auth := middleware.GetUser(r)
	if strings.ToUpper(auth.Role) == "OWNER" && branchID != "" {
		branchIDInt, err := strconv.Atoi(branchID)
		if err != nil {
			c.Log.WithError(err).Error("invalid branch id parameter")
			return model.NewAppErr("invalid branch id parameter", nil)
		}
		branchIDUint := uint(branchIDInt)
		request.BranchID = &branchIDUint
	} else {
		request.BranchID = auth.BranchID
	}

	responses, total, err := c.UseCase.SearchDaily(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching daily sales report")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.PurchasesDailyReportResponse]{
		Data:   responses,
		Paging: paging,
	})
}

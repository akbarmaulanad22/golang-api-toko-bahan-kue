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

type SaleReportController struct {
	Log     *logrus.Logger
	UseCase *usecase.SaleReportUseCase
}

func NewSaleReportController(useCase *usecase.SaleReportUseCase, logger *logrus.Logger) *SaleReportController {
	return &SaleReportController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *SaleReportController) ListDaily(w http.ResponseWriter, r *http.Request) error {

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

	request := &model.SearchSalesReportRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
		Page:    page,
		Size:    size,
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

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.SalesDailyReportResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *SaleReportController) ListTopSeller(w http.ResponseWriter, r *http.Request) error {

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

	request := &model.SearchSalesReportRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
		Page:    page,
		Size:    size,
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

	responses, total, err := c.UseCase.SearchTopSellerProduct(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching top seller sales report")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.SalesTopSellerReportResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *SaleReportController) ListCategory(w http.ResponseWriter, r *http.Request) error {

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

	request := &model.SearchSalesReportRequest{
		StartAt: startAtMili,
		EndAt:   endAtMili,
		Page:    page,
		Size:    size,
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

	responses, total, err := c.UseCase.SearchTopSellerCategory(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching top seller sales report")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.SalesCategoryResponse]{
		Data:   responses,
		Paging: paging,
	})
}

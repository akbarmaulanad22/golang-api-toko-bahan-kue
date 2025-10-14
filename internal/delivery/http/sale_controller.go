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

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type SaleController struct {
	Log     *logrus.Logger
	UseCase *usecase.SaleUseCase
}

func NewSaleController(useCase *usecase.SaleUseCase, logger *logrus.Logger) *SaleController {
	return &SaleController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *SaleController) Create(w http.ResponseWriter, r *http.Request) error {

	var request model.CreateSaleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	auth := middleware.GetUser(r)
	if auth.BranchID != nil {
		request.BranchID = *auth.BranchID
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create sale: %+v", err)
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.SaleResponse]{Data: response})
}

func (c *SaleController) List(w http.ResponseWriter, r *http.Request) error {

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

	request := &model.SearchSaleRequest{
		Search:  params.Get("search"),
		Status:  params.Get("status"),
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
			c.Log.WithError(err).Error("invalid branch id parameter")
			return model.NewAppErr("invalid branch id parameter", nil)
		}
		branchIDUint := uint(branchIDInt)
		request.BranchID = &branchIDUint
	} else {
		request.BranchID = auth.BranchID
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching sale")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.SaleResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *SaleController) Get(w http.ResponseWriter, r *http.Request) error {
	params := mux.Vars(r)

	code, ok := params["code"]
	if !ok || code == "" {
		return model.NewAppErr("invalid sale code parameter", nil)
	}

	request := &model.GetSaleRequest{
		Code: code,
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.SaleResponse]{Data: response})
}

func (c *SaleController) Cancel(w http.ResponseWriter, r *http.Request) error {

	params := mux.Vars(r)

	code, ok := params["code"]
	if !ok || code == "" {
		return model.NewAppErr("invalid sale code parameter", nil)
	}

	request := model.CancelSaleRequest{
		Code: code,
	}

	if err := c.UseCase.Cancel(r.Context(), &request); err != nil {
		c.Log.WithError(err).Warnf("Failed to cancel sale")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[bool]{Data: true})
}

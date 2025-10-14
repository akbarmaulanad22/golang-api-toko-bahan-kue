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

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type DebtController struct {
	Log     *logrus.Logger
	UseCase *usecase.DebtUseCase
}

func NewDebtController(useCase *usecase.DebtUseCase, logger *logrus.Logger) *DebtController {
	return &DebtController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *DebtController) List(w http.ResponseWriter, r *http.Request) error {

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

	request := &model.SearchDebtRequest{
		ReferenceType: params.Get("reference_type"),
		Search:        params.Get("search"),
		Status:        params.Get("status"),
		Page:          page,
		Size:          size,
		StartAt:       model.UnixDate(startAtMili),
		EndAt:         model.UnixDate(endAtMili),
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
		c.Log.WithError(err).Error("error searching branch")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.DebtResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *DebtController) Get(w http.ResponseWriter, r *http.Request) error {
	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.GetDebtRequest{
		ID: uint(idInt),
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting branch")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.DebtDetailResponse]{Data: response})
}

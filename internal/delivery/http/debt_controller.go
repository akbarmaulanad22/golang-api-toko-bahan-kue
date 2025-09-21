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

func (c *DebtController) List(w http.ResponseWriter, r *http.Request) {

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
			http.Error(w, err.Error(), helper.GetStatusCode(err))
			return
		}
		startAtMili = startAt

		endAt, err := helper.ParseDateToMilli(endAt, true)
		if err != nil {
			c.Log.WithError(err).Error("invalid end at parameter")
			http.Error(w, err.Error(), helper.GetStatusCode(err))
			return
		}
		endAtMili = endAt
	}

	request := &model.SearchDebtRequest{
		ReferenceType: params.Get("reference_type"),
		Search:        params.Get("search"),
		Status:        params.Get("status"),
		Page:          pageInt,
		Size:          sizeInt,
		StartAt:       model.UnixDate(startAtMili),
		EndAt:         model.UnixDate(endAtMili),
	}

	// if err := request.StartAt.ParseFromString(params.Get("start_at")); err != nil {
	// 	c.Log.WithError(err).Error("error parse start at params")
	// 	http.Error(w, err.Error(), helper.GetStatusCode(err))
	// 	return
	// }
	// if err := request.EndAt.ParseFromString(params.Get("end_at")); err != nil {
	// 	c.Log.WithError(err).Error("error parse end at params")
	// 	http.Error(w, err.Error(), helper.GetStatusCode(err))
	// 	return
	// }

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
		c.Log.WithError(err).Error("error searching branch")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.DebtResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *DebtController) Get(w http.ResponseWriter, r *http.Request) {
	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	request := &model.GetDebtRequest{
		ID: uint(idInt),
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting branch")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.DebtDetailResponse]{Data: response})
}

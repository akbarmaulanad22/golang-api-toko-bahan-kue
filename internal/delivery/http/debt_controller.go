package http

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
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

	pageInt, err := strconv.Atoi(pageStr)
	if err != nil {
		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
		return
	}

	sizeStr := params.Get("size")
	if sizeStr == "" {
		sizeStr = "10"
	}

	sizeInt, err := strconv.Atoi(sizeStr)
	if err != nil {
		http.Error(w, "invalid size parameter", http.StatusBadRequest)
		return
	}

	request := &model.SearchDebtRequest{
		ReferenceType: params.Get("refence_type"),
		ReferenceCode: params.Get("refence_code"),
		Page:          pageInt,
		Size:          sizeInt,
	}

	if err := request.StartAt.ParseFromString(params.Get("start_at")); err != nil {
		c.Log.WithError(err).Error("error parse start at params")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}
	if err := request.EndAt.ParseFromString(params.Get("end_at")); err != nil {
		c.Log.WithError(err).Error("error parse end at params")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
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

	json.NewEncoder(w).Encode(model.WebResponse[*model.DebtResponse]{Data: response})
}

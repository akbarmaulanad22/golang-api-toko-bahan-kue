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

type PurchaseController struct {
	Log     *logrus.Logger
	UseCase *usecase.PurchaseUseCase
}

func NewPurchaseController(useCase *usecase.PurchaseUseCase, logger *logrus.Logger) *PurchaseController {
	return &PurchaseController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *PurchaseController) Create(w http.ResponseWriter, r *http.Request) {

	var request model.CreatePurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	auth := middleware.GetUser(r)
	if auth.BranchID != nil {
		request.BranchID = *auth.BranchID
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create purchase: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.PurchaseResponse]{Data: response})
}

func (c *PurchaseController) List(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()

	page := params.Get("page")
	if page == "" {
		page = "1"
	}
	pageInt, _ := strconv.Atoi(page)

	size := params.Get("size")
	if size == "" {
		size = "10"
	}
	sizeInt, _ := strconv.Atoi(size)

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

	request := &model.SearchPurchaseRequest{
		Search:  params.Get("search"),
		Status:  params.Get("status"),
		StartAt: startAtMili,
		EndAt:   endAtMili,
		Page:    pageInt,
		Size:    sizeInt,
	}

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
		c.Log.WithError(err).Error("error searching purchase")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.PurchaseResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *PurchaseController) Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	code, ok := params["code"]
	if !ok || code == "" {
		http.Error(w, "Invalid purchase code parameter", http.StatusBadRequest)
		return
	}

	request := &model.GetPurchaseRequest{
		Code: code,
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting purchase")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.PurchaseResponse]{Data: response})
}

func (c *PurchaseController) Cancel(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	code, ok := params["code"]
	if !ok || code == "" {
		http.Error(w, "Invalid purchase code parameter", http.StatusBadRequest)
		return
	}

	request := model.CancelPurchaseRequest{
		Code: code,
	}

	if err := c.UseCase.Cancel(r.Context(), &request); err != nil {
		c.Log.WithError(err).Warnf("Failed to cancel purchase")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[bool]{Data: true})
}

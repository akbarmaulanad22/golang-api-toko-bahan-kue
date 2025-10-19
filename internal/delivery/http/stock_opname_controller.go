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

type StockOpnameController struct {
	Log     *logrus.Logger
	UseCase *usecase.StockOpnameUseCase
}

func NewStockOpnameController(useCase *usecase.StockOpnameUseCase, logger *logrus.Logger) *StockOpnameController {
	return &StockOpnameController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *StockOpnameController) Create(w http.ResponseWriter, r *http.Request) error {
	auth := middleware.GetUser(r)

	var request model.CreateStockOpnameRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}
	request.CreatedBy = auth.Username
	request.BranchID = *auth.BranchID

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error creating stock opname")
		return err
	}

	return helper.WriteJSON(w, http.StatusCreated, model.WebResponse[*model.StockOpnameResponse]{Data: response})
}

func (c *StockOpnameController) Update(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: %+v", err)
		return model.NewAppErr("invalid id parameter", nil)
	}

	var request model.UpdateStockOpnameRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	request.ID = uint(idInt)

	response, err := c.UseCase.Update(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error creating stock opname")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.StockOpnameResponse]{Data: response})
}

func (c *StockOpnameController) Delete(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: %+v", err)
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.DeleteStockOpnameRequest{
		ID: uint(idInt),
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting stock opname")
		return err
	}

	return helper.WriteJSON(w, http.StatusNoContent, nil)
}

func (c *StockOpnameController) Get(w http.ResponseWriter, r *http.Request) error {
	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: %+v", err)
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.GetStockOpnameRequest{
		ID: uint(idInt),
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting stock opname")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.StockOpnameResponse]{Data: response})
}

func (c *StockOpnameController) List(w http.ResponseWriter, r *http.Request) error {
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

	request := &model.SearchStockOpnameRequest{
		CreatedBy: params.Get("search"),
		Page:      page,
		Size:      size,
		Status:    params.Get("search"),
		DateFrom:  startAtMili,
		DateTo:    endAtMili,
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
		c.Log.WithError(err).Error("error searching stock opname")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.StockOpnameResponse]{
		Data:   responses,
		Paging: paging,
	})
}

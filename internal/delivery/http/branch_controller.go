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

type BranchController struct {
	Log     *logrus.Logger
	UseCase *usecase.BranchUseCase
}

func NewBranchController(useCase *usecase.BranchUseCase, logger *logrus.Logger) *BranchController {
	return &BranchController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *BranchController) Create(w http.ResponseWriter, r *http.Request) error {
	var request model.CreateBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error creating branch")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.BranchResponse]{Data: response})
}

func (c *BranchController) List(w http.ResponseWriter, r *http.Request) error {
	params := r.URL.Query()

	page := helper.ParseIntOrDefault(params.Get("page"), 1)
	size := helper.ParseIntOrDefault(params.Get("size"), 10)

	request := &model.SearchBranchRequest{
		Search: params.Get("search"),
		Page:   page,
		Size:   size,
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

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.BranchResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *BranchController) Get(w http.ResponseWriter, r *http.Request) error {
	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: %+v", err)
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.GetBranchRequest{
		ID: uint(idInt),
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting branch")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.BranchResponse]{Data: response})
}

func (c *BranchController) Update(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: %+v", err)
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := new(model.UpdateBranchRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	request.ID = uint(idInt)

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("error to update branch")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.BranchResponse]{Data: response})
}

func (c *BranchController) Delete(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: %+v", err)
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.DeleteBranchRequest{
		ID: uint(idInt),
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting branch")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[bool]{Data: true})
}

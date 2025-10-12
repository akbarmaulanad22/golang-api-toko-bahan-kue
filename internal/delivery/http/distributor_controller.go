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

type DistributorController struct {
	Log     *logrus.Logger
	UseCase *usecase.DistributorUseCase
}

func NewDistributorController(useCase *usecase.DistributorUseCase, logger *logrus.Logger) *DistributorController {
	return &DistributorController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *DistributorController) Create(w http.ResponseWriter, r *http.Request) error {
	var request model.CreateDistributorRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create distributor: %+v", err)
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.DistributorResponse]{Data: response})
}

func (c *DistributorController) List(w http.ResponseWriter, r *http.Request) error {
	params := r.URL.Query()

	page := helper.ParseIntOrDefault(params.Get("page"), 1)
	size := helper.ParseIntOrDefault(params.Get("size"), 10)

	request := &model.SearchDistributorRequest{
		Search: params.Get("search"),
		Page:   page,
		Size:   size,
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching distributor")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.DistributorResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *DistributorController) Get(w http.ResponseWriter, r *http.Request) error {
	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.GetDistributorRequest{
		ID: uint(idInt),
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting distributor")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.DistributorResponse]{Data: response})
}

func (c *DistributorController) Update(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := new(model.UpdateDistributorRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	request.ID = uint(idInt)

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to update distributor")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.DistributorResponse]{Data: response})
}

func (c *DistributorController) Delete(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.DeleteDistributorRequest{
		ID: uint(idInt),
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting distributor")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[bool]{Data: true})
}

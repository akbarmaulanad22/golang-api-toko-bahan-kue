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

type SizeController struct {
	Log     *logrus.Logger
	UseCase *usecase.SizeUseCase
}

func NewSizeController(useCase *usecase.SizeUseCase, logger *logrus.Logger) *SizeController {
	return &SizeController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *SizeController) Create(w http.ResponseWriter, r *http.Request) error {
	params := mux.Vars(r)

	productSKU, ok := params["productSKU"]
	if !ok || productSKU == "" {
		c.Log.Warnf("error to parse product sku parameter: missing or invalid product sku")
		return model.NewAppErr("invalid product sku parameter", nil)
	}

	var request model.CreateSizeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	request.ProductSKU = productSKU

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error creating size")
		return err
	}

	return helper.WriteJSON(w, http.StatusCreated, model.WebResponse[*model.SizeResponse]{Data: response})
}

func (c *SizeController) List(w http.ResponseWriter, r *http.Request) error {
	params := mux.Vars(r)

	query := r.URL.Query()

	productSKU, ok := params["productSKU"]
	if !ok || productSKU == "" {
		c.Log.Warnf("error to parse product sku parameter: missing or invalid product sku")
		return model.NewAppErr("invalid product sku parameter", nil)
	}
	page := helper.ParseIntOrDefault(query.Get("page"), 1)
	size := helper.ParseIntOrDefault(query.Get("size"), 10)

	request := &model.SearchSizeRequest{
		Name:       query.Get("name"),
		ProductSKU: productSKU,
		Page:       page,
		Size:       size,
	}

	responses, total, err := c.UseCase.Search(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching size")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.SizeResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *SizeController) Get(w http.ResponseWriter, r *http.Request) error {
	params := mux.Vars(r)
	productSKU, ok := params["productSKU"]
	if !ok || productSKU == "" {
		c.Log.Warnf("error to parse product sku parameter: missing or invalid product sku")
		return model.NewAppErr("invalid product sku parameter", nil)
	}

	idInt, err := strconv.Atoi(params["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: missing or invalid id")
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.GetSizeRequest{
		ID:         uint(idInt),
		ProductSKU: productSKU,
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting size")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.SizeResponse]{Data: response})
}

func (c *SizeController) Update(w http.ResponseWriter, r *http.Request) error {

	params := mux.Vars(r)
	productSKU, ok := params["productSKU"]
	if !ok || productSKU == "" {
		c.Log.Warnf("error to parse product sku parameter: missing or invalid product sku")
		return model.NewAppErr("invalid product sku parameter", nil)
	}

	idInt, err := strconv.Atoi(params["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: missing or invalid id")
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := new(model.UpdateSizeRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	request.ID = uint(idInt)
	request.ProductSKU = productSKU

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("error to update size")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.SizeResponse]{Data: response})
}

func (c *SizeController) Delete(w http.ResponseWriter, r *http.Request) error {
	params := mux.Vars(r)

	productSKU, ok := params["productSKU"]
	if !ok || productSKU == "" {
		c.Log.Warnf("error to parse product sku parameter: missing or invalid product sku")
		return model.NewAppErr("invalid product sku parameter", nil)
	}

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: missing or invalid id")
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.DeleteSizeRequest{
		ID:         uint(idInt),
		ProductSKU: productSKU,
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting size")
		return err
	}

	return helper.WriteJSON(w, http.StatusNoContent, nil)
}

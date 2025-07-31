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

func (c *BranchController) Create(w http.ResponseWriter, r *http.Request) {
	var request model.CreateBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create branch: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.BranchResponse]{Data: response})
}

func (c *BranchController) List(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	pageStr, ok := params["page"]
	if !ok || pageStr == "" {
		pageStr = "1"
	}

	pageInt, err := strconv.Atoi(pageStr)
	if err != nil {
		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
		return
	}

	sizeStr, ok := params["size"]
	if !ok || sizeStr == "" {
		sizeStr = "10"
	}

	sizeInt, err := strconv.Atoi(sizeStr)
	if err != nil {
		http.Error(w, "invalid size parameter", http.StatusBadRequest)
		return
	}

	request := &model.SearchBranchRequest{
		Name:    params["name"],
		Address: params["address"],
		Page:    pageInt,
		Size:    sizeInt,
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

	json.NewEncoder(w).Encode(model.WebResponse[[]model.BranchResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *BranchController) Get(w http.ResponseWriter, r *http.Request) {
	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	request := &model.GetBranchRequest{
		ID: uint(idInt),
	}

	response, err := c.UseCase.Get(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error getting branch")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.BranchResponse]{Data: response})
}

func (c *BranchController) Update(w http.ResponseWriter, r *http.Request) {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	request := new(model.UpdateBranchRequest)
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	request.ID = uint(idInt)

	response, err := c.UseCase.Update(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to update branch")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.BranchResponse]{Data: response})
}

func (c *BranchController) Delete(w http.ResponseWriter, r *http.Request) {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	request := &model.DeleteBranchRequest{
		ID: uint(idInt),
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting branch")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[bool]{Data: true})
}

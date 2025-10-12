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

type BranchInventoryController struct {
	Log     *logrus.Logger
	UseCase *usecase.BranchInventoryUseCase
}

func NewBranchInventoryController(useCase *usecase.BranchInventoryUseCase, logger *logrus.Logger) *BranchInventoryController {
	return &BranchInventoryController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *BranchInventoryController) Create(w http.ResponseWriter, r *http.Request) error {

	var request model.CreateBranchInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	if err := c.UseCase.Create(r.Context(), &request); err != nil {
		c.Log.Warnf("Failed to create branch inventory: %+v", err)
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[bool]{Data: true})
}

func (c *BranchInventoryController) List(w http.ResponseWriter, r *http.Request) error {

	params := r.URL.Query()

	page := helper.ParseIntOrDefault(params.Get("page"), 1)
	size := helper.ParseIntOrDefault(params.Get("size"), 10)
	branchID := params.Get("branch_id")

	request := &model.SearchBranchInventoryRequest{
		Search: params.Get("search"),
		Page:   page,
		Size:   size,
	}

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

	responses, total, err := c.UseCase.List(r.Context(), request)
	if err != nil {
		c.Log.WithError(err).Error("error searching branch inventory")
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[[]model.BranchInventoryProductResponse]{
		Data:   responses,
		Paging: paging,
	})
}

func (c *BranchInventoryController) Update(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return model.NewAppErr("invalid id parameter", nil)
	}

	var request model.UpdateBranchInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}

	request.ID = uint(idInt)

	if err := c.UseCase.Update(r.Context(), &request); err != nil {
		c.Log.Warnf("Failed to update branch inventory: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[bool]{Data: true})
}

func (c *BranchInventoryController) Delete(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.DeleteBranchInventoryRequest{
		ID: uint(idInt),
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting branch")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[bool]{Data: true})
}

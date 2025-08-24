package http

import (
	"encoding/json"
	"net/http"
	"strconv"
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

func (c *BranchInventoryController) ListOwnerInventoryByBranch(w http.ResponseWriter, r *http.Request) {

	responses, err := c.UseCase.ListOwnerInventoryByBranch(r.Context())
	if err != nil {
		c.Log.WithError(err).Error("error searching branch")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.BranchInventoryResponse]{
		Data: responses,
	})
}

func (c *BranchInventoryController) ListAdminInventory(w http.ResponseWriter, r *http.Request) {

	branchID := mux.Vars(r)["branchID"]
	if branchID == "" {
		c.Log.Warn("invalid branch id param")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	branchIDInt, _ := strconv.Atoi(branchID)

	var request model.BranchInventoryAdminRequest
	request.BranchID = uint(branchIDInt)

	responses, err := c.UseCase.ListAdminInventory(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error searching branch")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.BranchInventoryResponse]{
		Data: responses,
	})
}

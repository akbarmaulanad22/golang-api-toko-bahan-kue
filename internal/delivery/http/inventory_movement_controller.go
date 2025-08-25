package http

import (
	"encoding/json"
	"net/http"
	"tokobahankue/internal/delivery/http/middleware"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/usecase"

	"github.com/sirupsen/logrus"
)

type InventoryMovementController struct {
	Log     *logrus.Logger
	UseCase *usecase.InventoryMovementUseCase
}

func NewInventoryMovementController(useCase *usecase.InventoryMovementUseCase, logger *logrus.Logger) *InventoryMovementController {
	return &InventoryMovementController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *InventoryMovementController) Create(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetUser(r)

	var request model.BulkCreateInventoryMovementRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	request.BranchID = auth.BranchID

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create capital: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.BulkInventoryMovementResponse]{Data: response})
}

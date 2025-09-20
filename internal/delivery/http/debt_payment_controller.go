package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"tokobahankue/internal/delivery/http/middleware"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type DebtPaymentController struct {
	Log     *logrus.Logger
	UseCase *usecase.DebtPaymentUseCase
}

func NewDebtPaymentController(useCase *usecase.DebtPaymentUseCase, logger *logrus.Logger) *DebtPaymentController {
	return &DebtPaymentController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *DebtPaymentController) Create(w http.ResponseWriter, r *http.Request) {

	var request model.CreateDebtPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	debtIDInt, err := strconv.Atoi(mux.Vars(r)["debtID"])
	if err != nil {
		http.Error(w, "invalid debt ID", http.StatusBadRequest)
		return
	}

	auth := middleware.GetUser(r)
	request.DebtID = uint(debtIDInt)
	request.PaymentDate = time.Now().UnixMilli()
	request.BranchID = auth.BranchID

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.Warnf("Failed to create debt payment: %+v", err)
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.DebtPaymentResponse]{Data: response})
}

func (c *DebtPaymentController) Delete(w http.ResponseWriter, r *http.Request) {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	request := &model.DeleteDebtPaymentRequest{
		ID: uint(idInt),
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting debt payment")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[bool]{Data: true})
}

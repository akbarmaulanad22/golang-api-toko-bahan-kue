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

func (c *DebtPaymentController) Create(w http.ResponseWriter, r *http.Request) error {

	var request model.CreateDebtPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.Log.Warnf("error to parse request body: %+v", err)
		return model.NewAppErr("invalid request body", nil)
	}
	debtIDInt, err := strconv.Atoi(mux.Vars(r)["debtID"])
	if err != nil {
		c.Log.Warnf("error to parse debt id parameter: %+v", err)
		return model.NewAppErr("invalid debt id parameter", nil)
	}

	auth := middleware.GetUser(r)
	request.DebtID = uint(debtIDInt)
	request.PaymentDate = time.Now().UnixMilli()
	request.BranchID = auth.BranchID

	response, err := c.UseCase.Create(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error creating debt payment")
		return err
	}

	return helper.WriteJSON(w, http.StatusCreated, model.WebResponse[*model.DebtPaymentResponse]{Data: response})
}

func (c *DebtPaymentController) Delete(w http.ResponseWriter, r *http.Request) error {

	idInt, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		c.Log.Warnf("error to parse id parameter: %+v", err)
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := &model.DeleteDebtPaymentRequest{
		ID: uint(idInt),
	}

	if err := c.UseCase.Delete(r.Context(), request); err != nil {
		c.Log.WithError(err).Error("error deleting debt payment")
		return err
	}

	return helper.WriteJSON(w, http.StatusNoContent, nil)
}

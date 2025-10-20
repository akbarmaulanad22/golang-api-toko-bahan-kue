package http

import (
	"net/http"
	"strconv"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type PurchaseDetailController struct {
	Log     *logrus.Logger
	UseCase *usecase.PurchaseDetailUseCase
}

func NewPurchaseDetailController(useCase *usecase.PurchaseDetailUseCase, logger *logrus.Logger) *PurchaseDetailController {
	return &PurchaseDetailController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *PurchaseDetailController) Cancel(w http.ResponseWriter, r *http.Request) error {

	params := mux.Vars(r)

	code, ok := params["code"]
	if !ok || code == "" {
		c.Log.Warnf("error to parse sale code parameter: missing or invalid sale code")
		return model.NewAppErr("invalid sale code parameter", nil)
	}

	id, ok := params["id"]
	if !ok || code == "" {
		c.Log.Warnf("error to parse id parameter: missing or invalid id")
		return model.NewAppErr("invalid id parameter", nil)
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.Log.Warnf("error to parse id: %+v", err)
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := model.CancelPurchaseDetailRequest{
		ID:           uint(idInt),
		PurchaseCode: code,
	}

	if err := c.UseCase.Cancel(r.Context(), &request); err != nil {
		c.Log.WithError(err).Warnf("error to cancel sale")
		return err
	}

	return helper.WriteJSON(w, http.StatusNoContent, nil)
}

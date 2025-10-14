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
		return model.NewAppErr("invalid sale code parameter", nil)
	}

	sizeID, ok := params["sizeID"]
	if !ok || code == "" {
		return model.NewAppErr("invalid size id parameter", nil)
	}

	sizeIDInt, err := strconv.Atoi(sizeID)
	if err != nil {
		return model.NewAppErr("invalid size id parameter", nil)
	}

	request := model.CancelPurchaseDetailRequest{
		SizeID:       uint(sizeIDInt),
		PurchaseCode: code,
	}

	if err := c.UseCase.Cancel(r.Context(), &request); err != nil {
		c.Log.WithError(err).Warnf("Failed to cancel sale")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[bool]{Data: true})
}

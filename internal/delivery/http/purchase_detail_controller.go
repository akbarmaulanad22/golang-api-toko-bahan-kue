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

func (c *PurchaseDetailController) Cancel(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	code, ok := params["code"]
	if !ok || code == "" {
		http.Error(w, "Invalid sale code parameter", http.StatusBadRequest)
		return
	}

	sizeID, ok := params["sizeID"]
	if !ok || code == "" {
		http.Error(w, "Invalid size id parameter", http.StatusBadRequest)
		return
	}

	sizeIDInt, err := strconv.Atoi(sizeID)
	if err != nil {
		http.Error(w, "Invalid sale code parameter", http.StatusBadRequest)
		return
	}

	request := model.CancelPurchaseDetailRequest{
		SizeID:       uint(sizeIDInt),
		PurchaseCode: code,
	}

	if err := c.UseCase.Cancel(r.Context(), &request); err != nil {
		c.Log.WithError(err).Warnf("Failed to cancel sale")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[bool]{Data: true})
}

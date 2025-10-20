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

type SaleDetailController struct {
	Log     *logrus.Logger
	UseCase *usecase.SaleDetailUseCase
}

func NewSaleDetailController(useCase *usecase.SaleDetailUseCase, logger *logrus.Logger) *SaleDetailController {
	return &SaleDetailController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *SaleDetailController) Cancel(w http.ResponseWriter, r *http.Request) error {

	params := mux.Vars(r)

	code, ok := params["code"]
	if !ok || code == "" {
		c.Log.Warnf("error to parse sale code parameter: missing or invalid sale code")
		return model.NewAppErr("invalid sale code parameter", nil)
	}

	idInt, err := strconv.Atoi(params["id"])
	if err != nil {
		c.Log.Warnf("error to parse id: %+v", err)
		return model.NewAppErr("invalid id parameter", nil)
	}

	request := model.CancelSaleDetailRequest{
		SaleCode: code,
		ID:       uint(idInt),
	}

	if err := c.UseCase.Cancel(r.Context(), &request); err != nil {
		c.Log.WithError(err).Warnf("error to cancel sale")
		return err
	}

	return helper.WriteJSON(w, http.StatusNoContent, nil)
}

package http

import (
	"net/http"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/usecase"

	"github.com/sirupsen/logrus"
)

type DashboardController struct {
	Log     *logrus.Logger
	UseCase *usecase.DashboardUseCase
}

func NewDashboardController(useCase *usecase.DashboardUseCase, logger *logrus.Logger) *DashboardController {
	return &DashboardController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *DashboardController) Get(w http.ResponseWriter, r *http.Request) error {

	response, err := c.UseCase.Get(r.Context())
	if err != nil {
		c.Log.WithError(err).Error("error getting dashboard data")
		return err
	}

	return helper.WriteJSON(w, http.StatusOK, model.WebResponse[*model.DashboardResponse]{Data: response})
}

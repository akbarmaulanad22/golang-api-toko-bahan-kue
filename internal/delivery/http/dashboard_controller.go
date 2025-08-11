package http

import (
	"encoding/json"
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

func (c *DashboardController) Get(w http.ResponseWriter, r *http.Request) {

	response, err := c.UseCase.Get(r.Context())
	if err != nil {
		c.Log.WithError(err).Error("error getting count card dashboard")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[*model.DashboardResponse]{Data: response})
}

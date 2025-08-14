package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/usecase"

	"github.com/sirupsen/logrus"
)

type FinancialReportController struct {
	Log     *logrus.Logger
	UseCase *usecase.FinancialReportUseCase
}

func NewFinancialReportController(useCase *usecase.FinancialReportUseCase, logger *logrus.Logger) *FinancialReportController {
	return &FinancialReportController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *FinancialReportController) List(w http.ResponseWriter, r *http.Request) {

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	branchIDStr := r.URL.Query().Get("branch_id") // optional

	// Default: awal dan akhir bulan ini
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	layout := "2006-01-02"
	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse(layout, startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	} else {
		startDate = firstOfMonth
	}

	if endDateStr != "" {
		endDate, err = time.Parse(layout, endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	} else {
		endDate = lastOfMonth
	}

	// Branch ID (optional)
	var branchID *int
	if branchIDStr != "" {
		var bID int
		_, err := fmt.Sscanf(branchIDStr, "%d", &bID)
		if err == nil {
			branchID = &bID
		}
	}

	request := model.SearchDailyFinancialReportRequest{
		StartDate: startDate,
		EndDate:   endDate,
		BranchID:  branchID,
	}

	response, err := c.UseCase.SearchDailyReports(r.Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("error getting count card financial Report")
		http.Error(w, err.Error(), helper.GetStatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(model.WebResponse[[]model.DailyFinancialReportResponse]{Data: response})
}

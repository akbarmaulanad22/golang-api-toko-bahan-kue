package http

import (
	"tokobahankue/internal/usecase"

	"github.com/sirupsen/logrus"
)

type PurchaseController struct {
	Log     *logrus.Logger
	UseCase *usecase.PurchaseUseCase
}

func NewPurchaseController(useCase *usecase.PurchaseUseCase, logger *logrus.Logger) *PurchaseController {
	return &PurchaseController{
		Log:     logger,
		UseCase: useCase,
	}
}

// func (c *PurchaseController) Create(w http.ResponseWriter, r *http.Request) {

// 	auth := middleware.GetUser(r)

// 	var request model.CreatePurchaseRequest
// 	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
// 		c.Log.Warnf("Failed to parse request body: %+v", err)
// 		http.Error(w, "bad request", http.StatusBadRequest)
// 		return
// 	}

// 	request.BranchID = auth.BranchID

// 	response, err := c.UseCase.Create(r.Context(), &request)
// 	if err != nil {
// 		c.Log.Warnf("Failed to create purchase: %+v", err)
// 		http.Error(w, err.Error(), helper.GetStatusCode(err))
// 		return
// 	}

// 	json.NewEncoder(w).Encode(model.WebResponse[*model.PurchaseResponse]{Data: response})
// }

// func (c *PurchaseController) List(w http.ResponseWriter, r *http.Request) {
// 	params := r.URL.Query()

// 	page := params.Get("page")
// 	if page == "" {
// 		page = "1"
// 	}

// 	pageInt, err := strconv.Atoi(page)
// 	if err != nil {
// 		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
// 		return
// 	}

// 	size := params.Get("size")
// 	if size == "" {
// 		size = "10"
// 	}

// 	sizeInt, err := strconv.Atoi(size)
// 	if err != nil {
// 		http.Error(w, "Invalid size parameter", http.StatusBadRequest)
// 		return
// 	}

// 	now := time.Now()
// 	format := "2006-01-02"

// 	startAt := params.Get("start_at")
// 	endAt := params.Get("end_at")

// 	if startAt == "" && endAt == "" {
// 		// default: hari ini dan 30 hari ke depan
// 		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
// 		thirtyDaysLater := today.AddDate(0, 0, 30)

// 		startAt = today.Format(format)
// 		endAt = thirtyDaysLater.Format(format)
// 	}

// 	// parse tanggal dari input user atau default di atas
// 	startTime, _ := time.ParseInLocation(format, startAt, time.Local)
// 	endTime, _ := time.ParseInLocation(format, endAt, time.Local)

// 	// pastikan endTime sampai jam 23:59:59
// 	endTime = endTime.Add(time.Hour*23 + time.Minute*59 + time.Second*59)

// 	request := &model.SearchPurchaseRequest{
// 		Code:      params.Get("code"),
// 		SalesName: params.Get("sales_name"),
// 		Status:    params.Get("status"),
// 		StartAt:   startTime.UnixMilli(),
// 		EndAt:     endTime.UnixMilli(),
// 		Page:      pageInt,
// 		Size:      sizeInt,
// 	}

// 	responses, total, err := c.UseCase.Search(r.Context(), request)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error searching purchase")
// 		http.Error(w, err.Error(), helper.GetStatusCode(err))
// 		return
// 	}

// 	paging := &model.PageMetadata{
// 		Page:      request.Page,
// 		Size:      request.Size,
// 		TotalItem: total,
// 		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
// 	}

// 	json.NewEncoder(w).Encode(model.WebResponse[[]model.PurchaseResponse]{
// 		Data:   responses,
// 		Paging: paging,
// 	})
// }

// func (c *PurchaseController) Get(w http.ResponseWriter, r *http.Request) {
// 	params := mux.Vars(r)

// 	code, ok := params["code"]
// 	if !ok || code == "" {
// 		http.Error(w, "Invalid product sku parameter", http.StatusBadRequest)
// 		return
// 	}

// 	request := &model.GetPurchaseRequest{
// 		Code: code,
// 	}

// 	response, err := c.UseCase.Get(r.Context(), request)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error getting purchase")
// 		http.Error(w, err.Error(), helper.GetStatusCode(err))
// 		return
// 	}

// 	json.NewEncoder(w).Encode(model.WebResponse[*model.PurchaseResponse]{Data: response})
// }

// func (c *PurchaseController) Update(w http.ResponseWriter, r *http.Request) {

// 	params := mux.Vars(r)

// 	code, ok := params["code"]
// 	if !ok || code == "" {
// 		http.Error(w, "Invalid product sku parameter", http.StatusBadRequest)
// 		return
// 	}

// 	request := new(model.UpdatePurchaseRequest)
// 	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
// 		c.Log.Warnf("Failed to parse request body: %+v", err)
// 		http.Error(w, "bad request", http.StatusBadRequest)
// 		return
// 	}

// 	request.Code = code

// 	response, err := c.UseCase.Update(r.Context(), request)
// 	if err != nil {
// 		c.Log.WithError(err).Warnf("Failed to update purchase")
// 		http.Error(w, err.Error(), helper.GetStatusCode(err))
// 		return
// 	}

// 	json.NewEncoder(w).Encode(model.WebResponse[*model.PurchaseResponse]{Data: response})
// }

package usecase

import (
	"context"
	"errors"
	"strings"
	"time"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
	"tokobahankue/internal/model/converter"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SaleUseCase struct {
	DB             *gorm.DB
	Log            *logrus.Logger
	Validate       *validator.Validate
	SaleRepository *repository.SaleRepository
	DebtRepository *repository.DebtRepository
}

func NewSaleUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	saleRepository *repository.SaleRepository, debtRepository *repository.DebtRepository) *SaleUseCase {
	return &SaleUseCase{
		DB:             db,
		Log:            logger,
		Validate:       validate,
		SaleRepository: saleRepository,
		DebtRepository: debtRepository,
	}
}

func (c *SaleUseCase) Create(ctx context.Context, request *model.CreateSaleRequest) (*model.SaleResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Validasi input
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	if request.Debt != nil && len(request.Payments) > 0 {
		c.Log.Warn("error please select one payment method")
		return nil, errors.New("bad request")
	}

	// Buat sale
	saleCode := uuid.NewString()
	sale := &entity.Sale{
		Code:         saleCode,
		CustomerName: request.CustomerName,
		BranchID:     request.BranchID,
	}

	if err := c.SaleRepository.Create(tx, sale); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			switch {
			case strings.Contains(mysqlErr.Message, "for key 'sales.PRIMARY'"):
				c.Log.Warn("Code already exists")
				return nil, errors.New("conflict")
			default:
				c.Log.WithError(err).Error("unexpected duplicate entry")
				return nil, errors.New("conflict")
			}
		}
		c.Log.WithError(err).Error("error creating sale")
		return nil, errors.New("internal server error")
	}

	// Ambil harga dari sizes
	var sizeIDs []uint
	for _, d := range request.Details {
		sizeIDs = append(sizeIDs, d.SizeID)
	}

	var sizes []entity.Size
	if err := tx.Where("id IN ?", sizeIDs).Find(&sizes).Error; err != nil {
		c.Log.WithError(err).Error("error getting sizes")
		return nil, errors.New("internal server error")
	}

	sizeMap := make(map[uint]float64, len(sizes))
	for _, s := range sizes {
		sizeMap[s.ID] = s.SellPrice
	}

	// Sale details + total price
	var (
		details    = make([]entity.SaleDetail, len(request.Details))
		totalPrice float64
	)
	for i, d := range request.Details {
		price := sizeMap[d.SizeID]
		totalPrice += price * float64(d.Qty)
		details[i] = entity.SaleDetail{
			SaleCode:  saleCode,
			SizeID:    d.SizeID,
			Qty:       d.Qty,
			SellPrice: price,
		}
	}
	if err := tx.CreateInBatches(&details, 100).Error; err != nil {
		c.Log.WithError(err).Error("error creating bulk data sale details")
		return nil, errors.New("internal server error")
	}

	// Sale payments
	if len(request.Payments) > 0 {
		var totalPayment float64
		payments := make([]entity.SalePayment, len(request.Payments))
		for i, p := range request.Payments {
			payments[i] = entity.SalePayment{
				SaleCode:      saleCode,
				PaymentMethod: p.PaymentMethod,
				Amount:        p.Amount,
				Note:          p.Note,
			}
			totalPayment += p.Amount
		}

		if totalPayment < totalPrice {
			c.Log.Error("error payment is less than total price")
			return nil, errors.New("bad request")
		}

		if err := tx.CreateInBatches(&payments, 100).Error; err != nil {
			c.Log.WithError(err).Error("error creating bulk data sale payments")
			return nil, errors.New("internal server error")
		}

	}

	// Debt
	if request.Debt != nil {
		var dueDate int64
		if request.Debt.DueDate > 0 {
			dueDate = int64(request.Debt.DueDate)
		} else {
			dueDate = time.Now().Add(7 * 24 * time.Hour).UnixMilli() // default 7 hari
		}

		debt := &entity.Debt{
			ReferenceType: "SALE",
			ReferenceCode: saleCode,
			TotalAmount:   totalPrice,
			PaidAmount:    0,
			DueDate:       dueDate,
			Status:        "PENDING",
		}

		if err := c.DebtRepository.Create(tx, debt); err != nil {
			c.Log.WithError(err).Error("error creating debt")
			return nil, errors.New("internal server error")
		}

		if len(request.Debt.DebtPayments) > 0 {
			debtPayments := make([]entity.DebtPayment, len(request.Debt.DebtPayments))
			for i, p := range request.Debt.DebtPayments {
				debtPayments[i] = entity.DebtPayment{
					DebtID:      debt.ID,
					PaymentDate: time.Now().UnixMilli(),
					Amount:      p.Amount,
					Note:        p.Note,
				}
			}

			if err := tx.CreateInBatches(&debtPayments, 100).Error; err != nil {
				c.Log.WithError(err).Error("error creating debt payments")
				return nil, errors.New("internal server error")
			}
		}
	}

	// Commit transaksi
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error committing transaction")
		return nil, errors.New("internal server error")
	}

	return converter.SaleToResponse(sale), nil
}

func (c *SaleUseCase) Search(ctx context.Context, request *model.SearchSaleRequest) ([]model.SaleResponse, int64, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, errors.New("bad request")
	}

	sales, total, err := c.SaleRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting sales")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting sales")
		return nil, 0, errors.New("internal server error")
	}

	responses := make([]model.SaleResponse, len(sales))
	for i, sale := range sales {
		responses[i] = *converter.SaleToResponse(&sale)
	}

	return responses, total, nil
}

func (c *SaleUseCase) Get(ctx context.Context, request *model.GetSaleRequest) (*model.SaleResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	sale, err := c.SaleRepository.FindByCode(tx, request.Code)
	if err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return nil, errors.New("not found")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return nil, errors.New("internal server error")
	}

	return sale, nil
}

// func (c *SaleUseCase) Cancel(ctx context.Context, request *model.CancelSaleRequest) (*model.SaleResponse, error) {
// 	tx := c.DB.WithContext(ctx).Begin()
// 	defer tx.Rollback()

// 	if err := c.Validate.Struct(request); err != nil {
// 		c.Log.WithError(err).Error("error validating request body")
// 		return nil, errors.New("bad request")
// 	}

// 	sale, err := c.SaleRepository.FindByCode(tx, request.Code)
// 	if err != nil {
// 		c.Log.WithError(err).Error("error getting sale")
// 		return nil, errors.New("not found")
// 	}

// 	createdTime := time.UnixMilli(sale.CreatedAt)
// 	now := time.Now()

// 	// Hitung durasi sejak dibuat
// 	duration := now.Sub(createdTime)

// 	// Jika status BUKAN PENDING dan sudah lewat 24 jam => tolak
// 	if duration.Hours() >= 24 {
// 		c.Log.WithField("sale_code", sale.Code).Error("error updating sale: exceeded 24-hour window")
// 		return nil, errors.New("forbidden")
// 	}

// 	// Lanjut update status
// 	sale.Status = "CANCELLED"
// 	if err := c.SaleRepository.Update(tx, sale); err != nil {
// 		c.Log.WithError(err).Error("error updating sale")
// 		return nil, errors.New("internal server error")
// 	}

// 	if sale.Debt != nil {
// 		debt := new(entity.Debt)
// 		if err := c.DebtRepository.FindById(tx, debt, sale.Debt.ID); err != nil {
// 			c.Log.WithError(err).Error("error getting debt")
// 			return nil, errors.New("internal server error")
// 		}

// 		debt.Status = "VOID"

// 		if err := c.DebtRepository.Update(tx, debt); err != nil {
// 			c.Log.WithError(err).Error("error update debt")
// 			return nil, errors.New("internal server error")
// 		}
// 	}

// 	/*
// 		kalo mau return data setelah terupdate
// 	*/
// 	// if err := c.SaleRepository.FindByCode(tx, sale, request.Code); err != nil {
// 	// 	c.Log.WithError(err).Error("error getting sale")
// 	// 	return nil, errors.New("not found")
// 	// }

// 	if err := tx.Commit().Error; err != nil {
// 		c.Log.WithError(err).Error("error updating sale")
// 		return nil, errors.New("internal server error")
// 	}

// 	return converter.SaleToResponse(sale), nil
// }

// func (c *SaleUseCase) Get(ctx context.Context, request *model.GetSaleRequest) (*model.SaleResponse, error) {
// 	tx := c.DB.WithContext(ctx).Begin()
// 	defer tx.Rollback()

// 	if err := c.Validate.Struct(request); err != nil {
// 		c.Log.WithError(err).Error("error validating request body")
// 		return nil, errors.New("bad request")
// 	}

// 	sale := new(entity.Sale)
// 	if err := c.SaleRepository.FindByCode(tx, sale, request.Code); err != nil {
// 		c.Log.WithError(err).Error("error getting sale")
// 		return nil, errors.New("not found")
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		c.Log.WithError(err).Error("error getting sale")
// 		return nil, errors.New("internal server error")
// 	}

// 	return converter.SaleToResponse(sale), nil
// }

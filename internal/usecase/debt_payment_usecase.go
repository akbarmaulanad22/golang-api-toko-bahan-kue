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
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DebtPaymentUseCase struct {
	DB                    *gorm.DB
	Log                   *logrus.Logger
	Validate              *validator.Validate
	DebtPaymentRepository *repository.DebtPaymentRepository
}

func NewDebtPaymentUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	debtPaymentRepository *repository.DebtPaymentRepository) *DebtPaymentUseCase {
	return &DebtPaymentUseCase{
		DB:                    db,
		Log:                   logger,
		Validate:              validate,
		DebtPaymentRepository: debtPaymentRepository,
	}
}

func (c *DebtPaymentUseCase) Create(ctx context.Context, request *model.CreateDebtPaymentRequest) (*model.DebtPaymentResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	debtPayment := &entity.DebtPayment{
		Note:        request.Note,
		Amount:      request.Amount,
		PaymentDate: request.PaymentDate,
		DebtID:      request.DebtID,
	}

	if err := c.DebtPaymentRepository.Create(tx, debtPayment); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case 1452:
				if strings.Contains(mysqlErr.Message, "FOREIGN KEY (`debt_id`)") {
					c.Log.Warn("debt doesnt exists")
					return nil, errors.New("invalid debt id")
				}
				return nil, errors.New("foreign key constraint failed")
			}
		}

		c.Log.WithError(err).Error("error creating debt payment")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error creating debt payment")
		return nil, errors.New("internal server error")
	}

	return converter.DebtPaymentToResponse(debtPayment), nil
}

// func (c *DebtPaymentUseCase) Update(ctx context.Context, request *model.UpdateDebtPaymentRequest) (*model.DebtPaymentResponse, error) {
// 	tx := c.DB.WithContext(ctx).Begin()
// 	defer tx.Rollback()

// 	if err := c.Validate.Struct(request); err != nil {
// 		c.Log.WithError(err).Error("error validating request body")
// 		return nil, errors.New("bad request")
// 	}

// 	debtPayment := new(entity.DebtPayment)
// 	if err := c.DebtPaymentRepository.FindById(tx, debtPayment, request.ID); err != nil {
// 		c.Log.WithError(err).Error("error getting debt payment")
// 		return nil, errors.New("not found")
// 	}

// 	if debtPayment.Note == request.Note && debtPayment.Amount == request.Amount {
// 		return converter.DebtPaymentToResponse(debtPayment), nil
// 	}

// 	debtPayment.Note = request.Note
// 	debtPayment.Amount = request.Amount
// 	debtPayment.Type = request.Type

// 	if err := c.DebtPaymentRepository.Update(tx, debtPayment); err != nil {
// 		c.Log.WithError(err).Error("error updating debt payment")
// 		return nil, errors.New("internal server error")
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		c.Log.WithError(err).Error("error updating debt payment")
// 		return nil, errors.New("internal server error")
// 	}

// 	return converter.DebtPaymentToResponse(debtPayment), nil
// }

func (c *DebtPaymentUseCase) Delete(ctx context.Context, request *model.DeleteDebtPaymentRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	debtPayment := new(entity.DebtPayment)
	if err := c.DebtPaymentRepository.FindById(tx, debtPayment, request.ID); err != nil {
		c.Log.WithError(err).Error("error getting debt payment")
		return errors.New("not found")
	}

	// Konversi CreatedAt (milli detik) ke time.Time
	createdAtTime := time.UnixMilli(debtPayment.CreatedAt)

	// bandingkan waktu debtPayment.CreatedAt dengan saat ini
	if time.Since(createdAtTime) > time.Hour {
		c.Log.Warnf("error deleting debt payment: %d", request.ID)
		return errors.New("forbidden")
	}

	if err := c.DebtPaymentRepository.Delete(tx, debtPayment); err != nil {
		c.Log.WithError(err).Error("error deleting debt payment")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error deleting debt payment")
		return errors.New("internal server error")
	}

	return nil
}

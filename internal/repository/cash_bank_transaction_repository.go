package repository

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CashBankTransactionRepository struct {
	Repository[entity.CashBankTransaction]
	Log *logrus.Logger
}

func NewCashBankTransactionRepository(log *logrus.Logger) *CashBankTransactionRepository {
	return &CashBankTransactionRepository{
		Log: log,
	}
}

func (r *CashBankTransactionRepository) Search(db *gorm.DB, request *model.SearchCashBankTransactionRequest) ([]entity.CashBankTransaction, int64, error) {
	var users []entity.CashBankTransaction
	if err := db.Scopes(r.FilterCashBankTransaction(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.CashBankTransaction{}).Scopes(r.FilterCashBankTransaction(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *CashBankTransactionRepository) FilterCashBankTransaction(request *model.SearchCashBankTransactionRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {

		startAt := request.StartAt
		endAt := request.EndAt
		if startAt != 0 && endAt != 0 {

			tx = tx.Where(
				"(created_at BETWEEN ? AND ?) OR (transaction_date BETWEEN ? AND ?)",
				startAt, endAt, startAt, endAt,
			)
		}

		return tx
	}
}

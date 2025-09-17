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
	if err := db.Preload("Branch").Scopes(r.FilterCashBankTransaction(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&users).Error; err != nil {
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

		if search := request.Search; search != "" {
			amount := search
			search = "%" + search + "%"
			tx = tx.Where("description LIKE ? OR reference_key LIKE ? OR amount = ?", search, search, amount)
		}

		if transactionType := request.Type; transactionType != "" {
			tx = tx.Where("type = ?", transactionType)
		}

		if source := request.Source; source != "" {
			tx = tx.Where("source = ?", source)
		}

		if branchID := request.BranchID; branchID != nil {
			tx = tx.Where("branch_id = ?", branchID)
		}

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

package repository

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BranchRepository struct {
	Repository[entity.Branch]
	Log *logrus.Logger
}

func NewBranchRepository(log *logrus.Logger) *BranchRepository {
	return &BranchRepository{
		Log: log,
	}
}

func (r *BranchRepository) Search(db *gorm.DB, request *model.SearchBranchRequest) ([]entity.Branch, int64, error) {
	var users []entity.Branch
	if err := db.Scopes(r.FilterBranch(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.Branch{}).Scopes(r.FilterBranch(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *BranchRepository) FilterBranch(request *model.SearchBranchRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if name := request.Name; name != "" {
			name = "%" + name + "%"
			tx = tx.Where("name LIKE ?", name)
		}

		if address := request.Address; address != "" {
			address = "%" + address + "%"
			tx = tx.Where("address LIKE ?", address)
		}

		return tx
	}
}

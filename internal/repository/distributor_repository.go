package repository

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DistributorRepository struct {
	Repository[entity.Distributor]
	Log *logrus.Logger
}

func NewDistributorRepository(log *logrus.Logger) *DistributorRepository {
	return &DistributorRepository{
		Log: log,
	}
}

func (r *DistributorRepository) CountByNameAndAddress(db *gorm.DB, name any, address any) (int64, error) {
	var total int64
	err := db.Model(&entity.Distributor{}).Where("name = ? AND address = ?", name, address).Count(&total).Error
	return total, err
}

func (r *DistributorRepository) Search(db *gorm.DB, request *model.SearchDistributorRequest) ([]entity.Distributor, int64, error) {
	var users []entity.Distributor
	if err := db.Scopes(r.FilterDistributor(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.Distributor{}).Scopes(r.FilterDistributor(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *DistributorRepository) FilterDistributor(request *model.SearchDistributorRequest) func(tx *gorm.DB) *gorm.DB {
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

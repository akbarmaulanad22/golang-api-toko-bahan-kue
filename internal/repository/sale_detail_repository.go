package repository

import (
	"tokobahankue/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SaleDetailRepository struct {
	Log *logrus.Logger
}

func NewSaleDetailRepository(log *logrus.Logger) *SaleDetailRepository {
	return &SaleDetailRepository{
		Log: log,
	}
}

func (r *SaleDetailRepository) CreateBulk(db *gorm.DB, details []entity.SaleDetail) error {
	return db.CreateInBatches(&details, 100).Error
}

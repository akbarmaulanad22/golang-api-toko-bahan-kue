package repository

import (
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DashboardRepository struct {
	Log *logrus.Logger
}

func NewDashboardRepository(log *logrus.Logger) *DashboardRepository {
	return &DashboardRepository{
		Log: log,
	}
}

func (r *DashboardRepository) CardCount(db *gorm.DB, dashboardResponse *model.DashboardResponse) error {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM users) AS total_employees,
			(SELECT COUNT(*) FROM products) AS total_products,
			(SELECT COUNT(*) FROM distributors) AS total_distributors,
			(SELECT COUNT(*) FROM branches) AS total_branches
	`

	return db.Raw(query).Scan(&dashboardResponse).Error
}

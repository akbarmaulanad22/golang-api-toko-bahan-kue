package repository

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	Repository[entity.Category]
	Log *logrus.Logger
}

func NewCategoryRepository(log *logrus.Logger) *CategoryRepository {
	return &CategoryRepository{
		Log: log,
	}
}

func (r *CategoryRepository) FindByName(db *gorm.DB, user *entity.Category, name string) error {
	return db.Where("name = ?", name).First(user).Error
}

func (r *CategoryRepository) Search(db *gorm.DB, request *model.SearchCategoryRequest) ([]entity.Category, int64, error) {
	var users []entity.Category
	if err := db.Scopes(r.FilterCategory(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.Category{}).Scopes(r.FilterCategory(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *CategoryRepository) FilterCategory(request *model.SearchCategoryRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if name := request.Name; name != "" {
			name = "%" + name + "%"
			tx = tx.Where("name LIKE ?", name)
		}
		return tx
	}
}

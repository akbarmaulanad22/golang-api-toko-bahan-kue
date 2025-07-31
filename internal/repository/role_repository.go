package repository

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RoleRepository struct {
	Repository[entity.Role]
	Log *logrus.Logger
}

func NewRoleRepository(log *logrus.Logger) *RoleRepository {
	return &RoleRepository{
		Log: log,
	}
}

func (r *RoleRepository) CountByName(db *gorm.DB, name any) (int64, error) {
	var total int64
	err := db.Model(&entity.Role{}).Where("name = ?", name).Count(&total).Error
	return total, err
}

func (r *RoleRepository) Search(db *gorm.DB, request *model.SearchRoleRequest) ([]entity.Role, int64, error) {
	var users []entity.Role
	if err := db.Scopes(r.FilterRole(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.Role{}).Scopes(r.FilterRole(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *RoleRepository) FilterRole(request *model.SearchRoleRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if name := request.Name; name != "" {
			name = "%" + name + "%"
			tx = tx.Where("name LIKE ?", name)
		}

		return tx
	}
}

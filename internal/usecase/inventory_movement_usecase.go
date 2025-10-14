package usecase

import (
	"context"
	"errors"
	"strings"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type InventoryMovementUseCase struct {
	DB                          *gorm.DB
	Log                         *logrus.Logger
	Validate                    *validator.Validate
	InventoryMovementRepository *repository.InventoryMovementRepository
	BranchInventoryRepository   *repository.BranchInventoryRepository
}

func NewInventoryMovementUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	inventoryMovementRepository *repository.InventoryMovementRepository, branchInventoryRepository *repository.BranchInventoryRepository) *InventoryMovementUseCase {
	return &InventoryMovementUseCase{
		DB:                          db,
		Log:                         logger,
		Validate:                    validate,
		InventoryMovementRepository: inventoryMovementRepository,
		BranchInventoryRepository:   branchInventoryRepository,
	}
}

func (c *InventoryMovementUseCase) Create(ctx context.Context, request *model.BulkCreateInventoryMovementRequest) (*model.BulkInventoryMovementResponse, error) {

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, model.NewAppErr("invalid request body", nil)
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	var responses []model.InventoryMovementResponse

	for _, m := range request.Movements {
		branchInv := &entity.BranchInventory{}
		err := c.BranchInventoryRepository.FindByBranchIDAndSizeID(tx, branchInv, request.BranchID, m.SizeID)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Belum ada → buat baru
			branchInv = &entity.BranchInventory{
				BranchID: request.BranchID,
				SizeID:   m.SizeID,
				Stock:    m.ChangeQty,
			}
			if err := c.BranchInventoryRepository.Create(tx, branchInv); err != nil {
				if mysqlErr, ok := err.(*mysql.MySQLError); ok {
					switch mysqlErr.Number {
					case 1452:
						c.Log.WithError(err).Error("foreign key constraint failed")
						return nil, model.NewAppErr("referenced resource not found", "the specified branch or size does not exist.")
					}
				}
				return nil, model.NewAppErr("internal server error", nil)
			}
		} else if err != nil {
			c.Log.WithError(err).Error("error querying branch inventory")
			return nil, model.NewAppErr("internal server error", nil)
		} else {
			if err := c.BranchInventoryRepository.UpdateStock(tx, branchInv.ID, m.ChangeQty); err != nil {
				return nil, model.NewAppErr("internal server error", nil)
			}
		}

		// Catat movement
		movement := &entity.InventoryMovement{
			BranchInventoryID: branchInv.ID,
			ChangeQty:         m.ChangeQty,
			ReferenceType:     request.ReferenceType,
			ReferenceKey:      request.ReferenceKey,
		}
		if err := c.InventoryMovementRepository.Create(tx, movement); err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); ok {
				switch mysqlErr.Number {
				case 1452:
					if strings.Contains(mysqlErr.Message, "FOREIGN KEY (`branch_inventory_id`)") {
						c.Log.WithError(err).Error("foreign key constraint failed")
						return nil, model.NewAppErr("referenced resource not found", "the specified branch inventory does not exist.")
					}
					return nil, model.NewAppErr("internal server error", nil)
				}
			}
			return nil, model.NewAppErr("internal server error", nil)
		}

		responses = append(responses, model.InventoryMovementResponse{
			ID:        movement.ID,
			ChangeQty: movement.ChangeQty,
			CreatedAt: movement.CreatedAt,
		})
	}

	if err := tx.Commit().Error; err != nil {
		return nil, model.NewAppErr("internal server error", nil)
	}

	return &model.BulkInventoryMovementResponse{
		ReferenceType: request.ReferenceType,
		ReferenceKey:  request.ReferenceKey,
		Movements:     responses,
	}, nil
}

func (c *InventoryMovementUseCase) Search(ctx context.Context, request *model.SearchInventoryMovementRequest) ([]model.InventoryMovementResponse, int64, error) {

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, 0, helper.GetValidationMessage(err)
	}

	inventoryMovements, total, err := c.InventoryMovementRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting inventory movements")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting inventory movements")
		return nil, 0, model.NewAppErr("internal server error", nil)
	}

	return inventoryMovements, total, nil
}

func (c *InventoryMovementUseCase) Summary(ctx context.Context, request *model.SearchInventoryMovementRequest) (*model.InventoryMovementSummaryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, helper.GetValidationMessage(err)
	}

	inventoryMovements, err := c.InventoryMovementRepository.Summary(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting inventory movements summary")
		return nil, model.NewAppErr("internal server error", nil)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting inventory movements summary")
		return nil, model.NewAppErr("internal server error", nil)
	}

	return inventoryMovements, nil
}

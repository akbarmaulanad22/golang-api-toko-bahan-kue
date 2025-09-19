package usecase

import (
	"context"
	"errors"
	"strings"
	"tokobahankue/internal/entity"
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
		return nil, errors.New("bad request")
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
						if strings.Contains(mysqlErr.Message, "FOREIGN KEY (`branch_id`)") {
							c.Log.Warn("branch doesnt exists")
							return nil, errors.New("invalid branch id")
						}
						if strings.Contains(mysqlErr.Message, "FOREIGN KEY (`size_id`)") {
							c.Log.Warn("size doesnt exists")
							return nil, errors.New("invalid size id")
						}
						return nil, errors.New("foreign key constraint failed")
					}
				}

				return nil, errors.New("error creating branch inventory")
			}
		} else if err != nil {
			c.Log.WithError(err).Error("error querying branch inventory")
			return nil, errors.New("internal server error")
		} else {
			// Sudah ada → update stok
			if err := c.BranchInventoryRepository.UpdateStock(tx, branchInv.ID, m.ChangeQty); err != nil {
				return nil, errors.New("error updating stock")
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
						c.Log.Warn("branch inventory doesnt exists")
						return nil, errors.New("invalid branch inventory id")
					}
					return nil, errors.New("foreign key constraint failed")
				}
			}
			return nil, errors.New("error creating inventory movement")
		}

		responses = append(responses, model.InventoryMovementResponse{
			ID:        movement.ID,
			ChangeQty: movement.ChangeQty,
			CreatedAt: movement.CreatedAt,
		})
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("commit transaction failed")
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
		return nil, 0, errors.New("bad request")
	}

	inventoryMovements, total, err := c.InventoryMovementRepository.Search(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting inventory movements")
		return nil, 0, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting inventory movements")
		return nil, 0, errors.New("internal server error")
	}

	return inventoryMovements, total, nil
}

func (c *InventoryMovementUseCase) Summary(ctx context.Context, request *model.SearchInventoryMovementRequest) (*model.InventoryMovementSummaryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	inventoryMovements, err := c.InventoryMovementRepository.Summary(tx, request)
	if err != nil {
		c.Log.WithError(err).Error("error getting inventory movements summary")
		return nil, errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error getting inventory movements summary")
		return nil, errors.New("internal server error")
	}

	return inventoryMovements, nil
}

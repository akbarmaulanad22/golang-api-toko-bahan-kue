package helper

import (
	"errors"
	"fmt"
	"tokobahankue/internal/model"

	"gorm.io/gorm"
)

func GetNotFoundMessage(entity string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NewAppErr(fmt.Sprintf("%s not found", entity), nil)
	}

	return model.NewAppErr("internal server error", nil)
}

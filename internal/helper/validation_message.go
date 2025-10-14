package helper

import (
	"tokobahankue/internal/model"

	"github.com/go-playground/validator/v10"
)

func GetValidationMessage(err error) error {
	details := make([]map[string]string, 0)
	for _, e := range err.(validator.ValidationErrors) {
		details = append(details, map[string]string{
			"field": e.Field(),
			"rule":  e.Tag(),
		})
	}
	return model.NewAppErr("validation error", details)
}

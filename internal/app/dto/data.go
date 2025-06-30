package dto

import (
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

var validate = validator.New()

// ErrorResponse response payload.
type ErrorResponse struct {
	Error string `json:"error"`
}

func decimalGreaterThanZero(fl validator.FieldLevel) bool {
	val, ok := fl.Field().Interface().(decimal.Decimal)
	if !ok {
		return false
	}

	return val.GreaterThan(decimal.Zero)
}

func init() { //nolint:gochecknoinits
	validate.RegisterValidation("decimal_gt_zero", decimalGreaterThanZero) //nolint:errcheck
}

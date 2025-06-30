package dto

import (
	"fmt"
	"net/http"

	"github.com/shopspring/decimal"
)

type CreateTransferRequest struct {
	SourceAccountID      int64           `json:"source_account_id"      validate:"required"`
	DestinationAccountID int64           `json:"destination_account_id" validate:"required"`
	Amount               decimal.Decimal `json:"amount"                 validate:"required,decimal_gt_zero"`
}

func (req *CreateTransferRequest) Bind(_ *http.Request) error {
	err := validate.Struct(req)
	if err != nil {
		return fmt.Errorf("validate transaction create request: %w", err)
	}

	return nil
}

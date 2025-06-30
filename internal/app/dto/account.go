package dto

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

type CreateAccountRequest struct {
	AccountID      int64           `json:"account_id"      validate:"required"`
	InitialBalance decimal.Decimal `json:"initial_balance" validate:"required,decimal_gt_zero"`
}

func (req *CreateAccountRequest) Bind(_ *http.Request) error {
	err := validate.Struct(req)
	if err != nil {
		return fmt.Errorf("validate account create request: %w", err)
	}

	return nil
}

type GetAccountRequest struct {
	ID int64 `json:"id" validate:"required"`
}

func (req *GetAccountRequest) Bind(r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return errors.New("id is required")
	}

	parsedID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid id format: %w", err)
	}

	req.ID = int64(parsedID)

	err = validate.Struct(req)
	if err != nil {
		return fmt.Errorf("validate get account request: %w", err)
	}

	return nil
}

type AccountResponse struct {
	AccountID int64           `json:"account_id"`
	Balance   decimal.Decimal `json:"balance"`
}

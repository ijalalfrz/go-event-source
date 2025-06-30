package repository

import (
	"errors"
	"strings"

	"github.com/ijalalfrz/go-event-source/internal/pkg/exception"
	"github.com/lib/pq"
)

var dbErrorMap = map[string]error{
	"SQLSTATE 23505": exception.ErrRecordNotUnique,
}

type errorMapper struct{}

func (m *errorMapper) mapError(err error) error {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		for dbCode, dbErr := range dbErrorMap {
			if strings.Contains(string(pgErr.Code), dbCode) {
				return dbErr
			}
		}
	}

	return err
}

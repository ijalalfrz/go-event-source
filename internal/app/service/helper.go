package service

import (
	"context"
	"errors"
	"time"

	"github.com/ijalalfrz/go-event-source/internal/app/dto"
)

func getRequestContext(ctx context.Context, requestTimeThreshold time.Duration) (dto.RequestContext, error) {
	reqContext, ok := dto.RequestFromContext(ctx)
	if !ok {
		return dto.RequestContext{}, errors.New("request context not found")
	}

	now := time.Now()
	timeDiff := now.Sub(reqContext.Timestamp)

	if timeDiff > requestTimeThreshold || timeDiff < -requestTimeThreshold {
		return dto.RequestContext{}, ErrInvalidRequestTime
	}

	return reqContext, nil
}

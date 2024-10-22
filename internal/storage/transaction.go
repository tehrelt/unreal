package storage

import (
	"context"
	"errors"
)

type CtxKey string

var (
	ErrNotEnrichedContext    = errors.New("not enriched context")
	ErrNoConnectionInContext = errors.New("no connection in context")
	ErrInvalidValue          = errors.New("invalid value")
)

type Manager interface {
	Do(context.Context, func(ctx context.Context) error) error
}

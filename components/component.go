package components

import (
	"context"
)

type Component interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Healthy() (bool, error)
	Cleanup(ctx context.Context) error
}

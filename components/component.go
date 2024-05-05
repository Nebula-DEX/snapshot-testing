package components

import (
	"context"
)

type Component interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Healthy() (bool, error)
	Cleanup(ctx context.Context) error
}

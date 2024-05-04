package components

import (
	"context"
	"io"
)

type Component interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Stdout(ctx context.Context) (io.ReadCloser, error)
	Stderr(ctx context.Context) (io.ReadCloser, error)
	Healthy() (bool, error)
	Cleanup(ctx context.Context) error
}

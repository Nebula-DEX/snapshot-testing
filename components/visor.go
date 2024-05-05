package components

import (
	"context"
	"io"

	"go.uber.org/zap"
)

type visor struct {
	logger *zap.Logger

	vegavisorHome string
}

func NewVisor(envName string, vegavisorHome string, logger *zap.Logger) (Component, error) {
	return &visor{
		logger: logger,

		vegavisorHome: vegavisorHome,
	}, nil
}

func (v *visor) Name() string {
	return "vegavisor"
}

// Healthy implements Component.
func (v *visor) Healthy() (bool, error) {
	panic("unimplemented")
}

// Logs implements Component.
func (v *visor) Stdout(ctx context.Context) (io.ReadCloser, error) {
	panic("unimplemented")
}

// Logs implements Component.
func (v *visor) Stderr(ctx context.Context) (io.ReadCloser, error) {
	panic("unimplemented")
}

// Start implements Component.
func (v *visor) Start(ctx context.Context) error {
	panic("unimplemented")
}

// Stop implements Component.
func (v *visor) Stop(ctx context.Context) error {
	panic("unimplemented")
}

// Stop implements Component.
func (v *visor) Cleanup(ctx context.Context) error {
	panic("unimplemented")
}

package components

import (
	"context"
	"io"
)

type visor struct {
	envName string

	vegavisorHome  string
	vegaHome       string
	tendermintHome string
}

func NewVisor(envName string, vegavisorHome string, vegaHome string, tendermintHome string) (Component, error) {
	return &visor{
		envName: envName,

		vegavisorHome:  vegavisorHome,
		vegaHome:       vegaHome,
		tendermintHome: tendermintHome,
	}, nil
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

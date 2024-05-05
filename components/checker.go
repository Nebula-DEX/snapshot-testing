package components

import (
	"context"
	"io"
)

type checker struct {
	envName string
}

func NewCHecker(envName string) (Component, error) {
	return &checker{
		envName: envName,
	}, nil
}

func (c *checker) Name() string {
	return "watchdog"
}

// Healthy implements Component.
func (c *checker) Healthy() (bool, error) {
	panic("unimplemented")
}

// Logs implements Component.
func (c *checker) Stdout(ctx context.Context) (io.ReadCloser, error) {
	panic("unimplemented")
}

// Logs implements Component.
func (c *checker) Stderr(ctx context.Context) (io.ReadCloser, error) {
	panic("unimplemented")
}

// Start implements Component.
func (c *checker) Start(ctx context.Context) error {
	panic("unimplemented")
}

// Stop implements Component.
func (c *checker) Stop(ctx context.Context) error {
	panic("unimplemented")
}

// Stop implements Component.
func (c *checker) Cleanup(ctx context.Context) error {
	panic("unimplemented")
}

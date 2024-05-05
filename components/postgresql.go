package components

import (
	"context"
	"fmt"
	"io"

	"github.com/vegaprotocol/snapshot-testing/clients/docker"
	"github.com/vegaprotocol/snapshot-testing/config"
)

type postgresql struct {
	containerName string

	dockerClient *docker.Client
}

func NewPostgresql(dockerClient *docker.Client) (Component, error) {
	return &postgresql{
		dockerClient: dockerClient,
	}, nil
}

// Healthy implements Component.
func (p *postgresql) Healthy() (bool, error) {
	if p.containerName == "" {
		return false, fmt.Errorf("the postgresql has not been started")
	}

	return p.dockerClient.ContainerRunning(context.Background(), p.containerName)
}

// Logs implements Component.
func (p *postgresql) Stdout(ctx context.Context) (io.ReadCloser, error) {
	if p.containerName == "" {
		return nil, fmt.Errorf("the postgresql has not been started")
	}

	return p.dockerClient.Stdout(ctx, p.containerName, true)
}

// Logs implements Component.
func (p *postgresql) Stderr(ctx context.Context) (io.ReadCloser, error) {

	if p.containerName == "" {
		return nil, fmt.Errorf("the postgresql has not been started")
	}

	return p.dockerClient.Stderr(ctx, p.containerName, true)
}

// Start implements Component.
func (p *postgresql) Start(ctx context.Context) error {
	err := p.dockerClient.RunContainer(ctx, config.PostgresqlConfig)
	p.containerName = config.PostgresqlConfig.Name
	if err != nil {
		return fmt.Errorf("failed to start postgresql component: %w", err)
	}

	return nil
}

// Stop implements Component.
func (p *postgresql) Stop(ctx context.Context) error {
	containerExist, err := p.dockerClient.ContainerExist(ctx, config.PostgresqlConfig.Name)
	if err != nil {
		return fmt.Errorf("failed to check if docker container exists: %w", err)
	}

	if containerExist {
		err := p.dockerClient.ContainerRemoveForce(ctx, config.PostgresqlConfig.Name)
		if err != nil {
			return fmt.Errorf("failed to remove existing container: %w", err)
		}
	}

	return nil
}

// Stop implements Component.
func (p *postgresql) Cleanup(ctx context.Context) error {
	return p.Stop(ctx)
}

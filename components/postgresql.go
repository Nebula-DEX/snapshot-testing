package components

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/vegaprotocol/snapshot-testing/clients/docker"
	"github.com/vegaprotocol/snapshot-testing/config"
	"github.com/vegaprotocol/snapshot-testing/logging"
	"go.uber.org/zap"
)

type postgresql struct {
	mainLogger    *zap.Logger
	stdoutLogger  *zap.Logger
	stderrLogger  *zap.Logger
	containerName string
	credentials   config.PostgreSQLCreds

	dockerClient *docker.Client
}

func NewPostgresql(dockerClient *docker.Client, credentials config.PostgreSQLCreds, mainLogger *zap.Logger, stdoutLogger *zap.Logger, stderrLogger *zap.Logger) (Component, error) {
	return &postgresql{
		mainLogger:   mainLogger,
		stdoutLogger: stdoutLogger,
		stderrLogger: stderrLogger,
		dockerClient: dockerClient,
		credentials:  credentials,
	}, nil
}

func (p *postgresql) Name() string {
	return "postgresql"
}

// Healthy implements Component.
func (p *postgresql) Healthy() (bool, error) {
	if p.containerName == "" {
		return false, fmt.Errorf("the postgresql has not been started")
	}

	running, err := p.dockerClient.ContainerRunning(context.Background(), p.containerName)
	if err != nil {
		return false, fmt.Errorf("failed to check if container is running: %w", err)
	}

	return running, nil
}

// Start implements Component.
func (p *postgresql) Start(ctx context.Context) error {
	container := config.PostgresqlConfig
	container.Environment["POSTGRES_USER"] = p.credentials.User
	container.Environment["POSTGRES_DB"] = p.credentials.DbName
	container.Environment["POSTGRES_PASSWORD"] = p.credentials.Pass
	container.Ports[p.credentials.Port] = p.credentials.Port

	err := p.dockerClient.RunContainer(ctx, container)
	p.containerName = container.Name
	if err != nil {
		return fmt.Errorf("failed to start postgresql component: %w", err)
	}

	stdout, err := p.dockerClient.Stdout(ctx, p.containerName, true)
	if err != nil {
		return fmt.Errorf("failed to get stdout stream for postgresql: %w", err)
	}
	stderr, err := p.dockerClient.Stderr(ctx, p.containerName, true)
	if err != nil {
		return fmt.Errorf("failed to get stderr stream for postgresql: %w", err)
	}
	defer stdout.Close()
	defer stderr.Close()

	go func(stream io.Reader) {
		if err := logging.StreamLogs(stream, p.stdoutLogger); err != nil && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			p.mainLogger.Error("failed to stream postgresql stdout", zap.Error(err))
		}
	}(stdout)

	go func(stream io.Reader) {
		if err := logging.StreamLogs(stream, p.stdoutLogger); err != nil && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			p.mainLogger.Error("failed to stream postgresql stdout", zap.Error(err))
		}
	}(stderr)

	<-ctx.Done()

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

func (p *postgresql) Result() map[string]interface{} {
	return map[string]interface{}{}
}

// Stop implements Component.
func (p *postgresql) Cleanup(ctx context.Context) error {
	return p.Stop(ctx)
}

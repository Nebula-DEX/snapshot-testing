package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/vegaprotocol/snapshot-testing/config"
)

var (
	ContainerNotFound = errors.New("container not found")
)

type OutputType string

const (
	Stdout OutputType = "stdout"
	Stderr OutputType = "stderr"
)

type Client struct {
	apiClient *client.Client
}

func NewClient() (*Client, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return nil, fmt.Errorf("failed to create docker api client from env: %w", err)
	}

	return &Client{
		apiClient: apiClient,
	}, nil
}

func NewClientWithApiClient(apiClient *client.Client) (*Client, error) {
	if apiClient == nil {
		return nil, fmt.Errorf("apiClient cannot be nil")
	}

	return &Client{
		apiClient: apiClient,
	}, nil
}

func (c *Client) fullContainerId(ctx context.Context, containerId string) (string, error) {
	list, err := c.apiClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	for _, cont := range list {
		if strings.Contains(cont.ID, containerId) || slices.Contains(cont.Names, containerId) || slices.Contains(cont.Names, fmt.Sprintf("/%s", containerId)) {
			return cont.ID, nil
		}
	}

	return "", ContainerNotFound
}

func (c *Client) ContainerExist(ctx context.Context, containerId string) (bool, error) {
	_, err := c.fullContainerId(ctx, containerId)
	if err == nil {
		return true, nil
	}

	if strings.Contains(err.Error(), "not found") {
		return false, nil
	}

	return false, err
}

func (c *Client) ContainerRunning(ctx context.Context, containerId string) (bool, error) {
	fullContainerId, err := c.fullContainerId(ctx, containerId)
	if err != nil {
		return false, fmt.Errorf("failed to get full container name: %w", err)
	}

	inspect, err := c.apiClient.ContainerInspect(ctx, fullContainerId)
	if err != nil {
		return false, fmt.Errorf("failed to inspect container: %w", err)
	}

	return inspect.State.Running, nil
}

func (c *Client) ContainerStarting(ctx context.Context, containerId string) (bool, error) {
	fullContainerId, err := c.fullContainerId(ctx, containerId)
	if err != nil {
		return false, fmt.Errorf("failed to get full container name: %w", err)
	}

	inspect, err := c.apiClient.ContainerInspect(ctx, fullContainerId)
	if err != nil {
		return false, fmt.Errorf("failed to inspect container: %w", err)
	}

	return (!inspect.State.Running &&
		!inspect.State.Dead &&
		!inspect.State.Restarting &&
		inspect.State.StartedAt != "" &&
		inspect.State.FinishedAt == ""), nil
}

func (c *Client) ContainerRemoveForce(ctx context.Context, containerId string) error {
	return c.apiClient.ContainerRemove(ctx, containerId, container.RemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         true,
	})
}

func (c *Client) RunContainer(ctx context.Context, config config.ContainerConfig) error {
	envs := []string{}
	for k, v := range config.Environment {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	exposedPorts := nat.PortSet{}
	portsMapping := nat.PortMap{}
	for k, v := range config.Ports {
		portsMapping[nat.Port(fmt.Sprintf("%d", k))] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: fmt.Sprintf("%d", v),
			},
		}
		exposedPorts[nat.Port(fmt.Sprintf("%d", k))] = struct{}{}
	}

	resp, err := c.apiClient.ContainerCreate(ctx, &container.Config{
		Image:        config.Image,
		Env:          envs,
		Cmd:          config.Command,
		ExposedPorts: exposedPorts,

		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
	}, &container.HostConfig{
		NetworkMode:  "bridge",
		PortBindings: portsMapping,
	}, &network.NetworkingConfig{}, &v1.Platform{}, config.Name)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	if err := c.apiClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

func (c *Client) logs(ctx context.Context, containerId string, logType OutputType, follow bool) (io.ReadCloser, error) {
	fullContainerId, err := c.fullContainerId(ctx, containerId)
	if err != nil {
		return nil, fmt.Errorf("failed to get full container name: %w", err)
	}

	logStream, err := c.apiClient.ContainerLogs(ctx, fullContainerId, container.LogsOptions{
		Follow:     follow,
		ShowStdout: logType == Stdout,
		ShowStderr: logType == Stderr,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get container logs stream: %w", err)
	}

	return logStream, nil
}

func (c *Client) Stdout(ctx context.Context, containerId string, follow bool) (io.ReadCloser, error) {
	return c.logs(ctx, containerId, Stdout, follow)
}

func (c *Client) Stderr(ctx context.Context, containerId string, follow bool) (io.ReadCloser, error) {
	return c.logs(ctx, containerId, Stderr, follow)
}

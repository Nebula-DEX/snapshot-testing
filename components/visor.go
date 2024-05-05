package components

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"time"

	"github.com/vegaprotocol/snapshot-testing/logging"
	"go.uber.org/zap"
)

type visor struct {
	started  bool
	finished bool

	mainLogger   *zap.Logger
	stdoutLogger *zap.Logger
	stderrLogger *zap.Logger
	commandStop  context.CancelFunc

	vegavisorBinary string
	vegavisorHome   string
}

func NewVisor(
	vegavisorBinary string,
	vegavisorHome string,
	mainLogger *zap.Logger,
	stdoutLogger *zap.Logger,
	stderrLogger *zap.Logger,
) (Component, error) {
	return &visor{
		mainLogger:   mainLogger,
		stdoutLogger: stderrLogger,
		stderrLogger: stderrLogger,

		vegavisorBinary: vegavisorBinary,
		vegavisorHome:   vegavisorHome,
	}, nil
}

func (v *visor) Name() string {
	return "vegavisor"
}

// Healthy implements Component.
func (v *visor) Healthy() (bool, error) {
	// Still not started
	if !v.started {
		return true, nil
	}

	// Program should not finish early
	return !v.finished, nil
}

func (v *visor) waitForPostgreSQL(ctx context.Context) error {
	ticker := time.NewTicker(time.Second * 5)
	v.mainLogger.Info("Waiting for postgresql to startup")

	for {
		address := net.JoinHostPort("127.0.0.1", "5432")
		// 3 second timeout
		conn, err := net.DialTimeout("tcp", address, 3*time.Second)
		if err != nil {
			v.mainLogger.Info("PostgreSQL port still not open")
		}
		if conn != nil {
			_ = conn.Close()
			v.mainLogger.Info("Found PostgreSQL running")
			return nil
		}
		v.mainLogger.Info("PostgreSQL still not running: connection cannot be established")

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return fmt.Errorf("PostgreSQL not found healthy in given time")
		}
	}
}

// Start implements Component.
func (v *visor) Start(ctx context.Context) error {
	commandContext, cancel := context.WithCancel(ctx)
	defer cancel()

	postgreSQLWaitContext, psqlWaitCancel := context.WithTimeout(commandContext, 60*time.Second)
	defer psqlWaitCancel()
	if err := v.waitForPostgreSQL(postgreSQLWaitContext); err != nil {
		return fmt.Errorf("postgreSQL did not start in 60 seconds: %w", err)
	}

	v.commandStop = cancel

	cmd := exec.CommandContext(commandContext, v.vegavisorBinary, []string{"run", "--home", v.vegavisorHome}...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	go func(cmd *exec.Cmd) {
		time.Sleep(30 * time.Second)
		v.started = true
		if err := cmd.Run(); err != nil {
			v.mainLogger.Error("failed to start vegavisor", zap.Error(err))
		}
		v.finished = true
	}(cmd)

	defer stdout.Close()
	defer stderr.Close()

	go func(stream io.Reader) {
		if err := logging.StreamLogs(stream, v.stdoutLogger); err != nil {
			v.mainLogger.Error("failed to stream visor stdout", zap.Error(err))
		}
	}(stdout)

	go func(stream io.Reader) {
		if err := logging.StreamLogs(stream, v.stdoutLogger); err != nil {
			v.mainLogger.Error("failed to stream visor stdout", zap.Error(err))
		}
	}(stderr)

	<-commandContext.Done()

	return nil
}

// Stop implements Component.
func (v *visor) Stop(ctx context.Context) error {
	if v.started {
		v.commandStop()
	}

	return nil
}

// Stop implements Component.
func (v *visor) Cleanup(ctx context.Context) error {
	return v.Stop(ctx)
}

package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/snapshot-testing/clients/docker"
	"github.com/vegaprotocol/snapshot-testing/components"
	"github.com/vegaprotocol/snapshot-testing/config"
	"github.com/vegaprotocol/snapshot-testing/logging"
	"github.com/vegaprotocol/snapshot-testing/networkutils"
	"go.uber.org/zap"
)

var testDuration time.Duration

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Prepare local node and run it for given time.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSnapshotTesting(testDuration); err != nil {
			panic(err)
		}
	},
}

func init() {
	runCmd.PersistentFlags().DurationVar(&testDuration, "duration", 15*time.Minute, "duration of test")
}

func runSnapshotTesting(duration time.Duration) error {
	pathManager := networkutils.NewPathManager(workDir)
	if err := pathManager.CreateDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to prepare working directory: %w", err)
	}

	// We do not want to log this to file
	mainLogger := logging.CreateLogger(zap.InfoLevel, pathManager.LogFile("main.log"), true, true)
	networkConfig, err := config.NetworkConfigForEnvironmentName(environment)
	if err != nil {
		return fmt.Errorf("failed to get network config: %w", err)
	}

	if err := prepareNetwork(mainLogger, pathManager, *networkConfig, config.DefaultCredentials); err != nil {
		return fmt.Errorf("failed to setup local network: %w", err)
	}

	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	psqlStdoutLogger := logging.CreateLogger(zap.InfoLevel, pathManager.LogFile("psql-stdout.log"), false, false)
	psqlStderrLogger := logging.CreateLogger(zap.InfoLevel, pathManager.LogFile("psql-stderr.log"), false, false)

	postgresql, err := components.NewPostgresql(
		dockerClient,
		config.DefaultCredentials,
		mainLogger,
		psqlStdoutLogger,
		psqlStderrLogger,
	)
	if err != nil {
		return fmt.Errorf("failed to create postgresql component: %w", err)
	}

	visorStdoutLogger := logging.CreateLogger(zap.InfoLevel, pathManager.LogFile("visor-stdout.log"), false, false)
	visorStderrLogger := logging.CreateLogger(zap.InfoLevel, pathManager.LogFile("visor-stderr.log"), false, false)

	visor, err := components.NewVisor(
		pathManager.VisorBin(),
		pathManager.VisorHome(),
		mainLogger,
		visorStdoutLogger,
		visorStderrLogger,
	)
	if err != nil {
		return fmt.Errorf("failed to create visor component: %w", err)
	}

	testsComponents := []components.Component{
		postgresql,
		visor,
	}

	if err := components.Run(pathManager, mainLogger, testsComponents); err != nil {
		return fmt.Errorf("failed to run test components: %w", err)
	}

	return nil
}

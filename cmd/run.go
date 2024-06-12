package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/vegaprotocol/snapshot-testing/clients/docker"
	"github.com/vegaprotocol/snapshot-testing/components"
	"github.com/vegaprotocol/snapshot-testing/config"
	"github.com/vegaprotocol/snapshot-testing/logging"
	"github.com/vegaprotocol/snapshot-testing/networkutils"
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

	if err := prepareNetwork(mainLogger.Named("prepare-network"), pathManager, *networkConfig, config.DefaultCredentials); err != nil {
		if shouldSkipFailure(err) {
			snapshotTestingResults := map[string]any{
				"should-skip-failure": true,
			}
			if err := writeResult(duration, mainLogger, snapshotTestingResults, pathManager); err != nil {
				return err
			}
		}
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
		mainLogger.Named("postgresql"),
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
		mainLogger.Named("visor"),
		visorStdoutLogger,
		visorStderrLogger,
	)
	if err != nil {
		return fmt.Errorf("failed to create visor component: %w", err)
	}

	watchdog, err := components.NewWatchdog(networkConfig.DataNodesREST, mainLogger.Named("watchdog"))
	if err != nil {
		return fmt.Errorf("failed to create watchdog component: %w", err)
	}

	testsComponents := []components.Component{
		postgresql,
		visor,
		watchdog,
	}

	testCtx, testCancel := context.WithTimeout(context.Background(), duration)
	defer testCancel()

	if err := components.Run(testCtx, pathManager, mainLogger.Named("controller"), testsComponents); err != nil {
		return fmt.Errorf("failed to run test components: %w", err)
	}

	// Run post-snapshot-testing actions
	snapshotMin, snapshotMax, err := networkutils.LocalSnapshotsRange(&pathManager)
	if err != nil {
		return fmt.Errorf("failed to get snapshot range: %w", err)
	}

	snapshotTestingResults := components.MergeResults(
		postgresql.Result(),
		watchdog.Result(),
		visor.Result(),
	)
	snapshotTestingResults["snapshot-min"] = snapshotMin
	snapshotTestingResults["snapshot-max"] = snapshotMax
	snapshotTestingResults["should-skip-failure"] = false

	return writeResult(duration, mainLogger, snapshotTestingResults, pathManager)
}

func shouldSkipFailure(err error) bool {
	return environment == config.NetworkNameDevnet1 && (errors.Is(err, networkutils.ErrNoHealthyNodeFound) || errors.Is(err, networkutils.ErrNoSnapshotForRestartFound))
}

func writeResult(duration time.Duration, mainLogger *zap.Logger, snapshotTestingResults components.ComponentResults, pathManager networkutils.PathManager) error {
	mainLogger.Sugar().Infof("Snapshot testing finished after %s", duration.String())
	jsonResults, err := json.MarshalIndent(snapshotTestingResults, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot-testing results into JSON: %w", err)
	}
	mainLogger.Sugar().Infof("Result: %s", jsonResults)
	mainLogger.Sugar().Infof("Writing results to the %s file", pathManager.Results())
	if err := os.WriteFile(pathManager.Results(), jsonResults, os.ModePerm); err != nil {
		return fmt.Errorf("failed to write results to %s: %w", pathManager.Results(), err)
	}
	return nil
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/snapshot-testing/config"
	"github.com/vegaprotocol/snapshot-testing/logging"
	"github.com/vegaprotocol/snapshot-testing/networkutils"
	"go.uber.org/zap"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare local node only and print command to start it.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ensureWorkingDirectory(workDir); err != nil {
			panic(err)
		}

		// We do not want to log this to file
		stdoutOnlyLogger := logging.CreateLogger(zap.InfoLevel, logging.DoNotLogToFile, true)
		networkConfig, err := config.NetworkConfigForEnvironmentName(environment)
		if err != nil {
			stdoutOnlyLogger.Fatal("failed to get network config", zap.Error(err))
		}

		pathManager := networkutils.NewPathManager(workDir)

		if err := prepareNetwork(stdoutOnlyLogger, pathManager, *networkConfig, config.DefaultCredentials); err != nil {
			stdoutOnlyLogger.Fatal("failed to setup local network", zap.Error(err))
		}

		stdoutOnlyLogger.Info("")
		stdoutOnlyLogger.Info("")
		stdoutOnlyLogger.Info("To run your local node start: ")
		stdoutOnlyLogger.Sugar().Infof("%s run --home %s", pathManager.VisorBin(), pathManager.VisorHome())
	},
}

func prepareNetwork(
	logger *zap.Logger,
	pathManager networkutils.PathManager,
	networkConfig config.Network,
	postgreSQLCredentials config.PostgreSQLCreds,
) error {
	network, err := networkutils.NewNetwork(logger, networkConfig, pathManager)
	if err != nil {
		return fmt.Errorf("failed to create network utils: %w", err)
	}

	if err := network.SetupLocalNode(postgreSQLCredentials); err != nil {
		return fmt.Errorf("failed to setup local node: %w", err)
	}

	return nil
}

func ensureWorkingDirectory(path string) error {
	// logger.Sugar().Infof("Ensure the working directory(%s) exists", path) // TODO: We have no loger at this point
	if err := os.MkdirAll(workDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create working directory")
	}

	logsDir := filepath.Join(path, "logs")
	// logger.Sugar().Infof("Ensure the directory for logs (%s) exists", logsDir)
	if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create logs directory")
	}

	return nil
}

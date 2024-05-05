package cmd

// import (
// 	"fmt"
// 	"path/filepath"

// 	"github.com/spf13/cobra"
// 	"github.com/vegaprotocol/snapshot-testing/clients/docker"
// 	"github.com/vegaprotocol/snapshot-testing/components"
// 	"github.com/vegaprotocol/snapshot-testing/config"
// 	"github.com/vegaprotocol/snapshot-testing/logging"
// 	"go.uber.org/zap"
// )

// var runCmd = &cobra.Command{
// 	Use:   "run",
// 	Short: "Prepare local node and run it for given time.",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if err := ensureWorkingDirectory(workDir); err != nil {
// 			panic(err)
// 		}

// 		// We do not want to log this to file
// 		stdoutOnlyLogger := logging.CreateLogger(zap.InfoLevel, filepath.Join(workDir, "logs", "main.log"), true)
// 		networkConfig, err := config.NetworkConfigForEnvironmentName(environment)
// 		if err != nil {
// 			stdoutOnlyLogger.Fatal("failed to get network config", zap.Error(err))
// 		}

// 		localNodeDetails, err := prepareNetwork(stdoutOnlyLogger, workDir, *networkConfig, config.DefaultCredentials)
// 		if err != nil {
// 			stdoutOnlyLogger.Fatal("failed to setup local network", zap.Error(err))
// 		}

// 		dockerClient, err := docker.NewClient()
// 		if err != nil {
// 			return fmt.Errorf("failed to create docker client: %w", err)
// 		}

// 		postgresql, err := components.NewPostgresql(dockerClient)

// 	},
// }

package cmd

import "github.com/spf13/cobra"

var (
	workDir     string
	environment string
	configPath  string

	rootCmd = &cobra.Command{
		Use:   "snapshot-testing",
		Short: "Command that runs the snapshot-testing",
		Long: `The command setup local node to start it from the remote snapshot, then
starts it and runs it for given time.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&workDir,
		"work-dir",
		"/tmp/snapshot-testing",
		"the working directory, where binaries are downloaded and local node home directories are generated",
	)

	rootCmd.PersistentFlags().StringVar(
		&environment,
		"environment",
		"mainnet",
		"the environment you want to run testing on, available values are: mainnet, fairground, stagnet1, devnet1",
	)

	rootCmd.PersistentFlags().StringVar(
		&configPath,
		"config-path",
		"",
		"the config-file path for the network, if not empty the --environment flag is ignored(may be local or remote file (with https://))",
	)
	rootCmd.AddCommand(prepareCmd)
	rootCmd.AddCommand(runCmd)
}

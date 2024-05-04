package networkutils

import (
	"fmt"
	"path/filepath"

	"github.com/vegaprotocol/snapshot-testing/tools"
)

func updateVisorConfig(visorHome string, vegaBinary string, vegaHome string, tendermintHome string, workDir string) error {
	currentVersionFile := filepath.Join(visorHome, "genesis", "run-config.toml")
	vegaBinaryAbs, err := filepath.Abs(vegaBinary)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for vega binary: %w", err)
	}
	vegaHomeAbs, err := filepath.Abs(vegaHome)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for vega home: %w", err)
	}
	tendermintHomeAbs, err := filepath.Abs(tendermintHome)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for tendermint home: %w", err)
	}
	workDirAbs, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for working directory: %w", err)
	}

	if err := tools.UpdateConfig(currentVersionFile, "toml", map[string]interface{}{
		"vega.binary.path":      vegaBinaryAbs,
		"data_node.binary.path": vegaBinaryAbs,
		"vega.binary.args":      []string{"start", "--home", vegaHomeAbs, "--tendermint-home", tendermintHomeAbs},
		"data_node.binary.args": []string{"datanode", "start", "--home", vegaHomeAbs},
		"vega.rpc.socketPath":   filepath.Join(workDirAbs, "vega.sock"),
		"vega.rpc.httpPath":     "/rpc",
	}); err != nil {
		return fmt.Errorf("failed to update vegavisor config: %w", err)
	}

	return nil
}

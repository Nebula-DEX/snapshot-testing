package networkutils

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vegaprotocol/snapshot-testing/config"
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

	newConfigValues := map[string]interface{}{
		"vega.binary.path":      vegaBinaryAbs,
		"data_node.binary.path": vegaBinaryAbs,
		"vega.binary.args":      []string{"start", "--home", vegaHomeAbs, "--tendermint-home", tendermintHomeAbs},
		"data_node.binary.args": []string{"datanode", "start", "--home", vegaHomeAbs},
		"vega.rpc.socketPath":   filepath.Join(workDirAbs, "vega.sock"),
		"vega.rpc.httpPath":     "/rpc",
	}

	if err := tools.UpdateConfig(currentVersionFile, "toml", newConfigValues); err != nil {
		return fmt.Errorf("failed to update vegavisor config: %w", err)
	}

	supervisorConfigFilePath := filepath.Join(visorHome, "config.toml")
	newSupervisorConfigValues := map[string]interface{}{
		"maxNumberOfRestarts": 0,
	}

	if err := tools.UpdateConfig(supervisorConfigFilePath, "toml", newSupervisorConfigValues); err != nil {
		return fmt.Errorf("failed to update vegavisor main config: %w", err)
	}

	return nil
}

func updateVegaConfig(vegaHome string, workDir string, startSnapshot Snapshot) error {
	configFilePath := filepath.Join(vegaHome, "config", "node", "config.toml")
	workDirAbs, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for working directory: %w", err)
	}

	newConfigValues := map[string]interface{}{
		"Admin.Server.SocketPath":   filepath.Join(workDirAbs, "vega.sock"),
		"Admin.Server.HTTPPath":     "/rpc",
		"Broker.Socket.Enabled":     true,
		"Broker.Socket.DialTimeout": "4h",
		"Snapshot.StartHeight":      startSnapshot.BlockHeight,
	}

	if err := tools.UpdateConfig(configFilePath, "toml", newConfigValues); err != nil {
		return fmt.Errorf("failed to update vega config: %w", err)
	}

	return nil
}

func updateTendermintConfig(tendermintHome string, rpcPeers []string, seeds []string, snapshot Snapshot, externalAddress string) error {
	configFilePath := filepath.Join(tendermintHome, "config", "config.toml")
	newConfigValues := map[string]interface{}{
		"log_level":              "debug",
		"p2p.seeds":              strings.Join(seeds, ","),
		"p2p.pex":                true,
		"statesync.enable":       true,
		"statesync.rpc_servers":  strings.Join(rpcPeers, ","),
		"statesync.trust_period": "672h0m0s",
		"statesync.trust_height": snapshot.BlockHeight,
		"statesync.trust_hash":   snapshot.BlockHash,
		"p2p.addr_book_strict":   false,
		"p2p.seed_mode":          true,
		"p2p.allow_duplicate_ip": true,
		"p2p.laddr":              "tcp://0.0.0.0:36656",
	}

	if len(externalAddress) > 0 {
		withPortRegex := regexp.MustCompile(`.*:\d{1,5}$`)
		if !withPortRegex.MatchString(externalAddress) {
			// add port if it is missing in the given address
			externalAddress = fmt.Sprintf("%s:36656", externalAddress)
		}

		newConfigValues["p2p.external_address"] = externalAddress
	}

	if err := tools.UpdateConfig(configFilePath, "toml", newConfigValues); err != nil {
		return fmt.Errorf("failed to update tendermint config: %w", err)
	}

	return nil
}

func updateDataNodeConfig(vegaHome string, bootstrapPeers []string, psqlCreds config.PostgreSQLCreds) error {
	configFilePath := filepath.Join(vegaHome, "config", "data-node", "config.toml")
	newConfigValues := map[string]interface{}{
		"SQLStore.RetentionPeriod":                    "standard",
		"SQLStore.ConnectionConfig.Host":              psqlCreds.Host,
		"SQLStore.ConnectionConfig.Port":              psqlCreds.Port,
		"SQLStore.ConnectionConfig.Username":          psqlCreds.User,
		"SQLStore.ConnectionConfig.Password":          psqlCreds.Pass,
		"SQLStore.ConnectionConfig.Database":          psqlCreds.DbName,
		"SQLStore.WipeOnStartup":                      true,
		"NetworkHistory.Store.BootstrapPeers":         bootstrapPeers,
		"NetworkHistory.Initialise.MinimumBlockCount": 1000,
		"NetworkHistory.Initialise.Timeout":           "4h",
		"NetworkHistory.RetryTimeout":                 "15s",
		"API.RateLimit.Rate":                          300.0,
		"API.RateLimit.Burst":                         1000,
		"AutoInitialiseFromNetworkHistory":            true,
	}

	if err := tools.UpdateConfig(configFilePath, "toml", newConfigValues); err != nil {
		return fmt.Errorf("failed to update data-node config: %w", err)
	}

	return nil
}

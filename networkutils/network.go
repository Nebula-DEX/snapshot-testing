package networkutils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"time"

	"github.com/vegaprotocol/snapshot-testing/config"
	"github.com/vegaprotocol/snapshot-testing/tools"
	"go.uber.org/zap"
)

const (
	HealthyTimeThreshold   = time.Second * 300
	HealthyBlocksThreshold = 450
)

const (
	BinariesFolder = "bins"
)

type Network struct {
	logger      *zap.Logger
	conf        config.Network
	pathManager PathManager

	healthyRESTEndpoints []string
	healthyRPCPeers      []string
	restartSnapshot      *Snapshot
	chainId              string

	appVersion string
	height     uint64
}

func NewNetwork(logger *zap.Logger, conf config.Network, pm PathManager) (*Network, error) {
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config for network: %w", err)
	}

	return &Network{
		logger:      logger,
		conf:        conf,
		pathManager: pm,
	}, nil
}

func (n *Network) getHealthyRPCPeers() ([]string, error) {
	if len(n.healthyRPCPeers) > 0 {
		return n.healthyRPCPeers, nil
	}

	n.logger.Info("Looking for a healthy RPC peers")

	networkHeadHeight, err := n.getNetworkHeight()
	if err != nil {
		return nil, fmt.Errorf("failed to get network height: %w", err)
	}

	healthyPeers := []string{}
	for _, rpcPeer := range n.conf.RPCPeers {
		if len(rpcPeer.CoreREST) < 1 {
			n.logger.Sugar().Infof("The %s peer does not have core REST assigned. Skipping", rpcPeer.Endpoint)
			continue
		}
		if isRESTEndpointHealthy(n.logger, networkHeadHeight, rpcPeer.CoreREST) {
			n.logger.Sugar().Infof("The %s RPC peer is healthy", rpcPeer.Endpoint)
			healthyPeers = append(healthyPeers, rpcPeer.Endpoint)
		}
	}

	if len(healthyPeers) < 1 {
		return nil, fmt.Errorf("no healthy RPC peers found")
	}

	n.healthyRPCPeers = healthyPeers

	return healthyPeers, nil
}

func (n *Network) getNetworkHeight() (uint64, error) {
	// We do not care about latest results here
	if n.height > 0 {
		return n.height, nil
	}

	heights := []uint64{}

	n.logger.Info("Fetching statistics from all the REST endpoint to get the network height")
	for _, restURL := range n.conf.DataNodesREST {
		n.logger.Sugar().Infof("Fetching statistics from %s", restURL)
		statistics, err := tools.RetryReturn(3, 500*time.Millisecond, func() (*Statistics, error) {
			return GetStatistics(restURL)
		})

		if err != nil {
			n.logger.Info(fmt.Sprintf("Failed to get statistics from %s", restURL), zap.Error(err))
			continue
		}

		heights = append(heights, statistics.BlockHeight)
		n.logger.Sugar().Infof("Height for %s: %d", restURL, statistics.BlockHeight)
	}

	if len(heights) == 0 {
		return 0, fmt.Errorf("no healthy rest endpoint found")
	}

	maxHeight := slices.Max(heights)
	n.height = maxHeight
	n.logger.Sugar().Infof("The network head height is %d", maxHeight)

	return maxHeight, nil
}

func (n *Network) getChainID() (string, error) {
	// We do not care about latest results here
	if n.chainId != "" {
		return n.chainId, nil
	}

	n.logger.Info("Fetching the network chain id")
	for _, restURL := range n.conf.DataNodesREST {
		n.logger.Sugar().Infof("Fetching statistics from %s", restURL)
		statistics, err := tools.RetryReturn(3, 500*time.Millisecond, func() (*Statistics, error) {
			return GetStatistics(restURL)
		})

		if err != nil {
			n.logger.Info(fmt.Sprintf("Failed to get statistics from %s", restURL), zap.Error(err))
			continue
		}

		if statistics.ChainID != "" {
			n.logger.Sugar().Infof("Found network chain id on node %s: %s", restURL, statistics.ChainID)
			n.chainId = statistics.ChainID

			return n.chainId, nil
		}
	}

	return "", fmt.Errorf("not received any valid response from statistics rest endpoints")
}

func (n *Network) getAppVersion() (string, error) {
	if len(n.conf.BinaryVersionOverride) > 0 {
		n.logger.Sugar().Infof("Binary version is override in the config to version %s", n.conf.BinaryVersionOverride)

		return n.conf.BinaryVersionOverride, nil
	}

	if len(n.appVersion) > 0 {
		return n.appVersion, nil
	}
	n.logger.Info("Fetching the network app version")

	healthyRESTEndpoints, err := n.getHealthyRESTEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get healthy rest endpoints: %w", err)
	}

	for _, restURL := range healthyRESTEndpoints {
		n.logger.Sugar().Infof("Fetching statistics from %s", restURL)
		statistics, err := tools.RetryReturn(3, 500*time.Millisecond, func() (*Statistics, error) {
			return GetStatistics(restURL)
		})

		if err != nil {
			n.logger.Info(fmt.Sprintf("Failed to fetch valid response from %s", restURL), zap.Error(err))
			continue
		}
		if statistics.AppVersion != "" {
			n.logger.Sugar().Infof("Found network app version on node %s: %s", restURL, statistics.AppVersion)
			n.appVersion = statistics.AppVersion

			return statistics.AppVersion, nil
		}
	}

	return "", fmt.Errorf("failed to find the app version for the network: no valid response received from the healthy endpoints")
}

func (n *Network) getHealthyRESTEndpoints() ([]string, error) {
	if len(n.healthyRESTEndpoints) > 0 {
		return n.healthyRESTEndpoints, nil
	}

	n.logger.Info("Getting all healthy rest endpoints for the network")

	networkHeadHeight, err := n.getNetworkHeight()
	if err != nil {
		return nil, fmt.Errorf("failed to get network height: %w", err)
	}

	healthyNodes := []string{}
	for _, restURL := range n.conf.DataNodesREST {
		if isRESTEndpointHealthy(n.logger, networkHeadHeight, restURL) {
			healthyNodes = append(healthyNodes, restURL)
		}
	}

	if len(healthyNodes) == 0 {
		return nil, fmt.Errorf("no healthy rest endpoint found")
	}

	n.logger.Sugar().Infof("All healthy REST endpoints: %v", healthyNodes)
	n.healthyRESTEndpoints = healthyNodes

	return n.healthyRESTEndpoints, nil
}

func (n *Network) binaryArtifactURL(kind string) (string, error) {
	osPart := "linux"

	switch runtime.GOOS {
	case "linux":
		osPart = "linux"
	case "darwin":
		osPart = "darwin"
	default:
		return "", fmt.Errorf("operating system not supported: only windows and linux supported, got %s", runtime.GOOS)
	}

	appVersion, err := n.getAppVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get app version: %w", err)
	}

	archPart := ""

	switch runtime.GOARCH {
	case "amd64":
		archPart = "amd64"
	case "arm", "arm64":
		archPart = "arm64"
	default:
		return "", fmt.Errorf("system architecture not supported: only amd64 and arm64 supported, got %s", runtime.GOARCH)
	}

	url := fmt.Sprintf(
		"https://github.com/%s/releases/download/%s/%s-%s-%s.zip",
		n.conf.ArtifactsRepository,
		appVersion,
		kind,
		osPart,
		archPart,
	)

	return url, nil
}

func (n *Network) DownloadFile(kind string, force bool, cleanup bool) (string, error) {
	n.logger.Sugar().Infof("Preparing URL for %s binary", kind)

	zipOutputFile := filepath.Join(n.pathManager.workDir, fmt.Sprintf("%s.zip", kind))
	if force {
		n.logger.Sugar().Infof("Removing old %s binaries", kind)
		if err := os.RemoveAll(zipOutputFile); err != nil {
			return "", fmt.Errorf("failed to cleanup: %w", err)
		}
	}

	url, err := n.binaryArtifactURL(kind)
	if err != nil {
		return "", fmt.Errorf("failed to get url for %s binary: %w", kind, err)
	}

	n.logger.Sugar().Infof("Downloading the %s file", url)
	if err := tools.DownloadFile(url, zipOutputFile); err != nil {
		return "", fmt.Errorf("failed to download %s binary: %w", kind, err)
	}

	binariesPath := n.pathManager.Binaries()
	n.logger.Sugar().Infof("Extracting downloaded binary to %s", binariesPath)

	if err := os.MkdirAll(binariesPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create binaries folder(%s): %w", binariesPath, err)
	}

	if err := tools.UnzipFile(zipOutputFile, binariesPath); err != nil {
		return "", fmt.Errorf("failed to unzip downloaded binary: %w", err)
	}
	n.logger.Sugar().Infof("Binary extracted to %s", binariesPath)

	// TODO: Maybe we can cleanup on defer???
	n.logger.Sugar().Infof("The %s binary saved in %s", kind, zipOutputFile)
	if cleanup {
		n.logger.Info("Removing temporary files")
		if err := os.RemoveAll(zipOutputFile); err != nil {
			return "", fmt.Errorf("failed to cleanup: %w", err)
		}
	}

	return filepath.Join(binariesPath, kind), nil
}

func (n *Network) downloadVegaBinary() error {
	_, err := n.DownloadFile("vega", true, true)
	if err != nil {
		return fmt.Errorf("failed to download vega binary: %w", err)
	}

	return nil
}

func (n *Network) downloadVegaVisorBinary() error {
	_, err := n.DownloadFile("visor", true, true)
	if err != nil {
		return fmt.Errorf("failed to download visor binary: %w", err)
	}

	return nil
}

// GetRestartSnapshot select snapshot for tendermint trusted block and height.
// It does not select the latest available snapshot. It select random snapshot
// between <X-6000; X-500>, where X is current block
func (n *Network) getRestartSnapshot() (*Snapshot, error) {
	if n.restartSnapshot != nil {
		return n.restartSnapshot, nil
	}
	n.logger.Info("Getting restart snapshot from the network REST API")

	healthyRESTEndpoints, err := n.getHealthyRESTEndpoints()
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy rest endpoints: %w", err)
	}

	networkHeadHeight, err := n.getNetworkHeight()
	if err != nil {
		return nil, fmt.Errorf("failed to get network head height: %w", err)
	}

	for _, endpoint := range healthyRESTEndpoints {
		n.logger.Sugar().Infof("Searching restart snapshot from REST api %s", endpoint)
		response, err := tools.RetryReturn(3, 500*time.Millisecond, func() ([]Snapshot, error) {
			return getSnapshots(endpoint)
		})

		if err != nil {
			n.logger.Info(fmt.Sprintf("cannot get snapshots from the REST endpoint(%s)", endpoint), zap.Error(err))
			continue
		}

		for _, snapshot := range response {
			if snapshot.BlockHeight >= networkHeadHeight-6000 && snapshot.BlockHeight <= networkHeadHeight-500 {
				n.logger.Sugar().Infof("Found restart snapshot %#v", snapshot)
				result := snapshot.Clone()
				n.restartSnapshot = &result

				return n.restartSnapshot, nil
			}
		}
	}

	return nil, fmt.Errorf("no snapshot for restart found")
}

func (n *Network) initLocally(force bool) error {
	if !n.pathManager.AreBinariesDownloaded() {
		return fmt.Errorf("Binaries are not downloaded")
	}

	if n.restartSnapshot == nil {
		return fmt.Errorf("missing restart snapshot")
	}

	chainID, err := n.getChainID()
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}
	if force {
		for _, folderPath := range []string{n.pathManager.VegaHome(), n.pathManager.VisorHome(), n.pathManager.TendermintHome()} {
			n.logger.Sugar().Infof("Removing the %s directory", folderPath)
			if err := os.RemoveAll(folderPath); err != nil {
				return fmt.Errorf("failed to remove the %s directory: %w", folderPath, err)
			}
			n.logger.Sugar().Infof("Removed the %s directory", folderPath)
		}
	}

	visorInitCommand := []string{
		n.pathManager.VisorBin(), "init",
		"--with-data-node",
		"--home", n.pathManager.VisorHome(),
	}
	vegaInitCommand := []string{
		n.pathManager.VegaBin(), "init",
		"--home", n.pathManager.VegaHome(),
		"--tendermint-home", n.pathManager.TendermintHome(),
		"--output", "json",
		"full",
	}
	dataNodeInitCommand := []string{
		n.pathManager.VegaBin(), "datanode", "init",
		"--home", n.pathManager.VegaHome(),
		chainID,
	}

	n.logger.Sugar().Infof("Initializing the vega visor with the following command: %v", visorInitCommand)
	if _, err := tools.ExecuteBinary(visorInitCommand[0], visorInitCommand[1:], nil); err != nil {
		return fmt.Errorf("failed to initialize vegavisor: %w", err)
	}
	n.logger.Sugar().Infof("Vegavisor initialized")

	n.logger.Sugar().Infof("Initializing the vega with the following command: %v", vegaInitCommand)
	if _, err := tools.ExecuteBinary(vegaInitCommand[0], vegaInitCommand[1:], nil); err != nil {
		return fmt.Errorf("failed to initialize vega: %w", err)
	}
	n.logger.Sugar().Infof("Vega initialized")

	n.logger.Sugar().Infof("Initializing the data-node with the following command: %v", dataNodeInitCommand)
	if _, err := tools.ExecuteBinary(dataNodeInitCommand[0], dataNodeInitCommand[1:], nil); err != nil {
		return fmt.Errorf("failed to initialize data-node: %w", err)
	}
	n.logger.Sugar().Infof("DataNode initialized")

	// n.logger.Sugar().Info("Executing vegavisor init with the %s home", visorHome)
	return nil
}

func (n *Network) downloadGenesis(tendermintHome string) error {
	genesisPath := filepath.Join(tendermintHome, "config", "genesis.json")

	n.logger.Sugar().Infof("Downloading genesis file from %s to %s", n.conf.GenesisURL, genesisPath)
	if err := tools.DownloadFile(n.conf.GenesisURL, genesisPath); err != nil {
		return fmt.Errorf("failed to download genesis: %w", err)
	}
	n.logger.Info("Genesis successfully downloaded")

	return nil
}

func (n *Network) SetupLocalNode(psqlCreds config.PostgreSQLCreds) error {

	if err := n.downloadVegaBinary(); err != nil {
		return fmt.Errorf("failed to download vega binary: %w", err)
	}

	if err := n.downloadVegaVisorBinary(); err != nil {
		return fmt.Errorf("failed to download visor binary: %w", err)
	}

	restartSnapshot, err := n.getRestartSnapshot()
	if err != nil {
		return fmt.Errorf("failed to get restart snapshot from the api: %w", err)
	}

	headHeight, err := n.getNetworkHeight()
	if err != nil {
		return fmt.Errorf("failed to get network head height: %w", err)
	}

	chainId, err := n.getChainID()
	if err != nil {
		return fmt.Errorf("failed to get chain id: %d", err)
	}

	rpcPeers, err := n.getHealthyRPCPeers()
	if err != nil {
		return fmt.Errorf("failed to get RPC peers: %w", err)
	}

	appVersion, err := n.getAppVersion()
	if err != nil {
		return fmt.Errorf("failed to get app version: %w", err)
	}

	overrideVersion := "no"
	if n.conf.BinaryVersionOverride != "" {
		overrideVersion = n.conf.BinaryVersionOverride
	}

	n.logger.Sugar().Info("")
	n.logger.Sugar().Info("===================================================")
	n.logger.Sugar().Info("Initializing local node with the following details:")
	n.logger.Sugar().Info("===================================================")
	n.logger.Sugar().Infof("Network head height: %d", headHeight)
	n.logger.Sugar().Infof("Network Chain ID: %s", chainId)
	n.logger.Sugar().Infof("Vega binary: %s", n.pathManager.VegaBin())
	n.logger.Sugar().Infof("Visor binary: %s", n.pathManager.VisorBin())
	n.logger.Sugar().Infof("Vega home: %s", n.pathManager.VegaHome())
	n.logger.Sugar().Infof("Visor home: %s", n.pathManager.VisorHome())
	n.logger.Sugar().Infof("Tendermint home: %s", n.pathManager.TendermintHome())
	n.logger.Sugar().Infof("Snapshot for restart: %#v", restartSnapshot)
	n.logger.Sugar().Infof("RPCPeers: %v", rpcPeers)
	n.logger.Sugar().Infof("Bootstrap peers: %v", n.conf.BootstrapPeers)
	n.logger.Sugar().Infof("Genesis file: %v", n.conf.GenesisURL)
	n.logger.Sugar().Infof("Seeds: %v", n.conf.Seeds)
	n.logger.Sugar().Infof("Network version: %s", appVersion)
	n.logger.Sugar().Infof("Override release: %s", overrideVersion)

	if err := n.initLocally(true); err != nil {
		return fmt.Errorf("failed to initialize node locally: %w", err)
	}

	if err := n.downloadGenesis(n.pathManager.TendermintHome()); err != nil {
		return fmt.Errorf("failed to download genesis: %w", err)
	}

	n.logger.Info("Updating vegavisor config")
	if err := updateVisorConfig(
		n.pathManager.VisorHome(),
		n.pathManager.VegaBin(),
		n.pathManager.VegaHome(),
		n.pathManager.TendermintHome(),
		n.pathManager.WorkDir()); err != nil {
		return fmt.Errorf("failed to update vegavisor config: %w", err)
	}

	n.logger.Info("Updating vega config")
	if err := updateVegaConfig(n.pathManager.VegaHome(), n.pathManager.WorkDir(), *restartSnapshot); err != nil {
		return fmt.Errorf("failed to update vega config: %w", err)
	}

	n.logger.Info("Updating tendermint config")
	if err := updateTendermintConfig(
		n.pathManager.TendermintHome(),
		rpcPeers,
		n.conf.Seeds,
		*restartSnapshot,
	); err != nil {
		return fmt.Errorf("failed to update tendermint config: %w", err)
	}

	n.logger.Info("Updating data-node config")
	if err := updateDataNodeConfig(n.pathManager.VegaHome(), n.conf.BootstrapPeers, psqlCreds); err != nil {
		return fmt.Errorf("failed to update data-node config: %w", err)
	}

	return nil
}

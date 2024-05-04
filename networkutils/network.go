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

type LocalNodeDetails struct {
	VegaBin        string
	VisorBin       string
	VegaHome       string
	TendermintHome string
	VisorHome      string
}

type Network struct {
	logger  *zap.Logger
	conf    config.Network
	workDir string

	healthyRESTEndpoints []string
	healthyRPCPeers      []string
	restartSnapshot      *Snapshot
	vegaBinaryPath       string
	visorBinaryPath      string
	chainId              string

	appVersion string
	height     uint64
}

func NewNetwork(logger *zap.Logger, conf config.Network, workDir string) (*Network, error) {
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config for network: %w", err)
	}

	return &Network{
		logger:  logger,
		conf:    conf,
		workDir: workDir,
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
		statistics, err := tools.RetryReturn(3, 500*time.Millisecond, func() (*statistics, error) {
			return getStatistics(restURL)
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
		statistics, err := tools.RetryReturn(3, 500*time.Millisecond, func() (*statistics, error) {
			return getStatistics(restURL)
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
		statistics, err := tools.RetryReturn(3, 500*time.Millisecond, func() (*statistics, error) {
			return getStatistics(restURL)
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

func (n *Network) downloadBinary(kind string, force bool, cleanup bool) (string, error) {
	n.logger.Sugar().Infof("Preparing URL for %s binary", kind)

	zipOutputFile := filepath.Join(n.workDir, fmt.Sprintf("%s.zip", kind))
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
	if err := tools.DownloadBinary(url, zipOutputFile); err != nil {
		return "", fmt.Errorf("failed to download %s binary: %w", kind, err)
	}

	binariesPath := filepath.Join(n.workDir, BinariesFolder)
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

func (n *Network) downloadVegaBinary() (string, error) {
	path, err := n.downloadBinary("vega", true, true)
	if err != nil {
		return "", fmt.Errorf("failed to download vega binary: %w", err)
	}

	n.vegaBinaryPath = path
	return path, nil
}

func (n *Network) downloadVegaVisorBinary() (string, error) {
	path, err := n.downloadBinary("visor", true, true)
	if err != nil {
		return "", fmt.Errorf("failed to download visor binary: %w", err)
	}

	n.visorBinaryPath = path
	return path, nil
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

func (n *Network) initLocally(force bool) (*LocalNodeDetails, error) {
	if n.vegaBinaryPath == "" {
		return nil, fmt.Errorf("empty vega binary path")
	}

	if n.visorBinaryPath == "" {
		return nil, fmt.Errorf("empty vegavisor binary path")
	}

	if n.restartSnapshot == nil {
		return nil, fmt.Errorf("missing restart snapshot")
	}

	chainID, err := n.getChainID()
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	visorHome := filepath.Join(n.workDir, "vegavisor_home")
	vegaHome := filepath.Join(n.workDir, "vega_home")
	tmHome := filepath.Join(n.workDir, "tendermint_home")

	if force {
		for _, folderPath := range []string{visorHome, vegaHome, tmHome} {
			n.logger.Sugar().Infof("Removing the %s directory", folderPath)
			if err := os.RemoveAll(folderPath); err != nil {
				return nil, fmt.Errorf("failed to remove the %s directory: %w", folderPath, err)
			}
			n.logger.Sugar().Infof("Removed the %s directory", folderPath)
		}
	}

	visorInitCommand := []string{
		n.visorBinaryPath, "init",
		"--with-data-node",
		"--home", visorHome,
	}
	vegaInitCommand := []string{
		n.vegaBinaryPath, "init",
		"--home", vegaHome,
		"--tendermint-home", tmHome,
		"--output", "json",
		"full",
	}
	dataNodeInitCommand := []string{
		n.vegaBinaryPath, "datanode", "init",
		"--home", vegaHome,
		chainID,
	}

	n.logger.Sugar().Infof("Initializing the vega visor with the following command: %v", visorInitCommand)
	if _, err := tools.ExecuteBinary(visorInitCommand[0], visorInitCommand[1:], nil); err != nil {
		return nil, fmt.Errorf("failed to initialize vegavisor: %w", err)
	}
	n.logger.Sugar().Infof("Vegavisor initialized")

	n.logger.Sugar().Infof("Initializing the vega with the following command: %v", vegaInitCommand)
	if _, err := tools.ExecuteBinary(vegaInitCommand[0], vegaInitCommand[1:], nil); err != nil {
		return nil, fmt.Errorf("failed to initialize vega: %w", err)
	}
	n.logger.Sugar().Infof("Vega initialized")

	n.logger.Sugar().Infof("Initializing the data-node with the following command: %v", dataNodeInitCommand)
	if _, err := tools.ExecuteBinary(dataNodeInitCommand[0], dataNodeInitCommand[1:], nil); err != nil {
		return nil, fmt.Errorf("failed to initialize data-node: %w", err)
	}
	n.logger.Sugar().Infof("DataNode initialized")

	// n.logger.Sugar().Info("Executing vegavisor init with the %s home", visorHome)
	return &LocalNodeDetails{
		VegaBin:        n.vegaBinaryPath,
		VisorBin:       n.visorBinaryPath,
		VegaHome:       vegaHome,
		TendermintHome: tmHome,
		VisorHome:      visorHome,
	}, nil
}

func (n *Network) SetupLocalNode() (*LocalNodeDetails, error) {
	vegaPath, err := n.downloadVegaBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to download vega binary: %w", err)
	}

	visorPath, err := n.downloadVegaVisorBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to download visor binary: %w", err)
	}

	restartSnapshot, err := n.getRestartSnapshot()
	if err != nil {
		return nil, fmt.Errorf("failed to get restart snapshot from the api: %w", err)
	}

	headHeight, err := n.getNetworkHeight()
	if err != nil {
		return nil, fmt.Errorf("failed to get network head height: %w", err)
	}

	chainId, err := n.getChainID()
	if err != nil {
		return nil, fmt.Errorf("failed to get chain id: %d", err)
	}

	rpcPeers, err := n.getHealthyRPCPeers()
	if err != nil {
		return nil, fmt.Errorf("failed to get RPC peers: %w", err)
	}

	appVersion, err := n.getAppVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get app version: %w", err)
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
	n.logger.Sugar().Infof("VegaPath: %s", vegaPath)
	n.logger.Sugar().Infof("VisorPath: %s", visorPath)
	n.logger.Sugar().Infof("Snapshot for restart: %#v", restartSnapshot)
	n.logger.Sugar().Infof("RPCPeers: %v", rpcPeers)
	n.logger.Sugar().Infof("Bootstrap peers: %v", n.conf.BootstrapPeers)
	n.logger.Sugar().Infof("Genesis file: %v", n.conf.GenesisURL)
	n.logger.Sugar().Infof("Seeds: %v", n.conf.Seeds)
	n.logger.Sugar().Infof("Network version: %s", appVersion)
	n.logger.Sugar().Infof("Override release: %s", overrideVersion)

	localNodeDetails, err := n.initLocally(true)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize node locally: %w", err)
	}

	n.logger.Info("Updating vegavisor config")
	if err := updateVisorConfig(
		localNodeDetails.VisorHome,
		localNodeDetails.VegaBin,
		localNodeDetails.VegaHome,
		localNodeDetails.TendermintHome,
		n.workDir); err != nil {
		return nil, fmt.Errorf("failed to update vegavisor config: %w", err)
	}

	return localNodeDetails, nil
}

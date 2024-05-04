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
	logger  *zap.Logger
	conf    config.Network
	workDir string

	healthyRESTEndpoints []string
	restartSnapshot      *Snapshot

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

func (n *Network) GetNetworkHeight() (uint64, error) {
	// We do not care about cache here
	if n.height > 0 {
		return n.height, nil
	}

	heights := []uint64{}

	n.logger.Info("Fetching statistics from all the REST endpoint to get the network height")
	for _, restURL := range n.conf.DataNodesREST {
		n.logger.Sugar().Debugf("Fetching statistics from %s", restURL)
		statistics, err := tools.RetryReturn(3, 500*time.Millisecond, func() (*statistics, error) {
			return getStatistics(restURL)
		})

		if err != nil {
			n.logger.Info(fmt.Sprintf("Failed to get statistics from %s", restURL), zap.Error(err))
			continue
		}

		heights = append(heights, statistics.BlockHeight)
		n.logger.Sugar().Debugf("Height for %s: %d", restURL, statistics.BlockHeight)
	}

	if len(heights) == 0 {
		return 0, fmt.Errorf("no healthy rest endpoint found")
	}

	maxHeight := slices.Max(heights)
	n.height = maxHeight
	n.logger.Sugar().Infof("The network head height is %d", maxHeight)

	return maxHeight, nil
}

func (n *Network) GetAppVersion() (string, error) {
	if len(n.conf.BinaryVersionOverride) > 0 {
		n.logger.Sugar().Infof("Binary version is override in the config to version %s", n.conf.BinaryVersionOverride)

		return n.conf.BinaryVersionOverride, nil
	}

	if len(n.appVersion) > 0 {
		return n.appVersion, nil
	}
	n.logger.Info("Fetching the network app version")

	healthyRESTEndpoints, err := n.GetHealthyRESTEndpoints()
	if err != nil {
		return "", fmt.Errorf("failed to get healthy rest endpoints: %w", err)
	}

	for _, restURL := range healthyRESTEndpoints {
		n.logger.Sugar().Debugf("Fetching statistics from %s", restURL)
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

func (n *Network) GetHealthyRESTEndpoints() ([]string, error) {
	if len(n.healthyRESTEndpoints) > 0 {
		return n.healthyRESTEndpoints, nil
	}

	n.logger.Debug("Getting all healthy rest endpoints for the network")

	networkHeadHeight, err := n.GetNetworkHeight()
	if err != nil {
		return nil, fmt.Errorf("failed to get network height: %w", err)
	}

	healthyNodes := []string{}
	for _, restURL := range n.conf.DataNodesREST {
		n.logger.Sugar().Debugf("Fetching statistics from %s", restURL)
		statistics, err := tools.RetryReturn(3, 500*time.Millisecond, func() (*statistics, error) {
			return getStatistics(restURL)
		})

		if err != nil {
			n.logger.Info(fmt.Sprintf("The %s endpoint unhealthy: failed to get statistics endpoint", restURL), zap.Error(err))
			continue
		}

		headBlocksDiff := networkHeadHeight - statistics.BlockHeight
		if statistics.BlockHeight < networkHeadHeight && headBlocksDiff > HealthyBlocksThreshold {
			n.logger.Sugar().Infof(
				"The %s endpoint unhealthy: core height(%d) is %d behind the network head(%d), only %d blocks lag allowed",
				restURL,
				statistics.BlockHeight,
				headBlocksDiff,
				networkHeadHeight,
				HealthyBlocksThreshold,
			)
			continue
		}

		if statistics.DataNodeHeight > 0 {
			blocksDiff := statistics.BlockHeight - statistics.DataNodeHeight
			if statistics.DataNodeHeight < statistics.BlockHeight && blocksDiff > HealthyBlocksThreshold {
				n.logger.Sugar().Infof(
					"The %s endpoint unhealthy: data node is %d blocks behind core, only %d blocks lag allowed",
					restURL,
					blocksDiff,
					HealthyBlocksThreshold,
				)
				continue
			}
		}

		timeDiff := statistics.CurrentTime.Sub(statistics.VegaTime)
		if timeDiff > HealthyTimeThreshold {
			n.logger.Sugar().Infof(
				"The %s endpoint unhealthy: time lag is %s, only %s allowed",
				restURL,
				timeDiff.String(),
				HealthyTimeThreshold.String(),
			)
			continue
		}

		n.logger.Sugar().Infof("The %s endpoint is healthy", restURL)
		healthyNodes = append(healthyNodes, restURL)
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

	appVersion, err := n.GetAppVersion()
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
	n.logger.Sugar().Info("Preparing URL for %s binary", kind)

	zipOutputFile := filepath.Join(n.workDir, fmt.Sprintf("%s.zip", kind))
	if force {
		n.logger.Sugar().Info("Removing old %s binaries", kind)
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

func (n *Network) DownloadVegaBinary() (string, error) {
	return n.downloadBinary("vega", true, true)
}

func (n *Network) DownloadVegaVisorBinary() (string, error) {
	return n.downloadBinary("visor", true, true)
}

// GetRestartSnapshot select snapshot for tendermint trusted block and height.
// It does not select the latest available snapshot. It select random snapshot
// between <X-6000; X-500>, where X is current block
func (n *Network) GetRestartSnapshot() (*Snapshot, error) {
	if n.restartSnapshot != nil {
		return n.restartSnapshot, nil
	}
	n.logger.Info("Getting restart snapshot from the network REST API")

	healthyRESTEndpoints, err := n.GetHealthyRESTEndpoints()
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy rest endpoints: %w", err)
	}

	networkHeadHeight, err := n.GetNetworkHeight()
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

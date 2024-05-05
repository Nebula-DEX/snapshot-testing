package networkutils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vegaprotocol/snapshot-testing/tools"
	"go.uber.org/zap"
)

type rawStatistics struct {
	Statistics struct {
		BlockHeight string
		CurrentTime string
		VegaTime    string
		ChainID     string
		AppVersion  string
	}
}

type Statistics struct {
	BlockHeight    uint64
	DataNodeHeight uint64

	CurrentTime time.Time
	VegaTime    time.Time

	ChainID    string
	AppVersion string
}

func GetStatistics(restURL string) (*Statistics, error) {
	statisticsURL := fmt.Sprintf("%s/statistics", strings.TrimRight(restURL, "/"))

	resp, err := http.Get(statisticsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to send get query to the statistics endpoint: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read statistics response body: %w", err)
	}

	rawResult := &rawStatistics{}
	if err := json.Unmarshal(body, rawResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal statistics response: %w", err)
	}

	blockHeight, err := strconv.ParseUint(rawResult.Statistics.BlockHeight, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed parse block height: %w", err)
	}

	currentTime, err := time.Parse(time.RFC3339Nano, rawResult.Statistics.CurrentTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current time: %w", err)
	}

	vegaTime, err := time.Parse(time.RFC3339Nano, rawResult.Statistics.VegaTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vega time: %w", err)
	}

	dataNodeHeight := uint64(0)
	if dataNodeHeightStr := resp.Header.Get("x-block-height"); dataNodeHeightStr != "" {
		dataNodeHeight, err = strconv.ParseUint(dataNodeHeightStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed parse data node block height: %w", err)
		}
	}

	result := &Statistics{
		BlockHeight:    blockHeight,
		DataNodeHeight: dataNodeHeight,
		CurrentTime:    currentTime,
		VegaTime:       vegaTime,

		ChainID:    rawResult.Statistics.ChainID,
		AppVersion: rawResult.Statistics.AppVersion,
	}

	return result, nil
}

func GetLatestStatistics(restEndpoints []string) (*Statistics, error) {
	if len(restEndpoints) < 1 {
		return nil, fmt.Errorf("no rest endpoint passed")
	}

	var latestStatistics *Statistics

	for _, endpoint := range restEndpoints {
		statistics, err := tools.RetryReturn(3, 500*time.Millisecond, func() (*Statistics, error) {
			return GetStatistics(endpoint)
		})

		if err != nil {
			// TODO: Maybe we can think about logging
			continue
		}

		if latestStatistics == nil || latestStatistics.BlockHeight < statistics.BlockHeight {
			latestStatistics = statistics
		}
	}

	if latestStatistics == nil {
		return nil, fmt.Errorf("all endpoints are unhealthy")
	}

	return latestStatistics, nil
}

type rawSnapshots struct {
	CoreSnapshots struct {
		Edges []struct {
			Node struct {
				BlockHeight string
				BlockHash   string
				CoreVersion string
			}
		}
	}
}

type Snapshot struct {
	BlockHeight uint64
	BlockHash   string
	CoreVersion string
}

func (s Snapshot) Clone() Snapshot {
	return Snapshot{
		BlockHeight: s.BlockHeight,
		BlockHash:   s.BlockHash,
		CoreVersion: s.CoreVersion,
	}
}

func getSnapshots(restURL string) ([]Snapshot, error) {
	snapshotsURL := fmt.Sprintf("%s/api/v2/snapshots", strings.TrimRight(restURL, "/"))

	resp, err := http.Get(snapshotsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to send get query to the statistics endpoint: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read statistics response body: %w", err)
	}

	rawResponse := &rawSnapshots{}
	if err := json.Unmarshal(body, rawResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshots response: %w", err)
	}

	response := []Snapshot{}

	for _, snap := range rawResponse.CoreSnapshots.Edges {
		snapshotHeight, err := strconv.ParseUint(snap.Node.BlockHeight, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse snapshot height(%s): %w", snap.Node.BlockHeight, err)
		}

		response = append(response, Snapshot{
			BlockHeight: snapshotHeight,
			BlockHash:   snap.Node.BlockHash,
			CoreVersion: snap.Node.CoreVersion,
		})
	}

	return response, nil
}

func isRESTEndpointHealthy(logger *zap.Logger, networkHeadHeight uint64, restURL string) bool {
	logger.Sugar().Infof("Fetching statistics from %s", restURL)
	statistics, err := tools.RetryReturn(3, 500*time.Millisecond, func() (*Statistics, error) {
		return GetStatistics(restURL)
	})

	if err != nil {
		logger.Info(fmt.Sprintf("The %s endpoint unhealthy: failed to get statistics endpoint", restURL), zap.Error(err))
		return false
	}

	headBlocksDiff := networkHeadHeight - statistics.BlockHeight
	if statistics.BlockHeight < networkHeadHeight && headBlocksDiff > HealthyBlocksThreshold {
		logger.Sugar().Infof(
			"The %s endpoint unhealthy: core height(%d) is %d behind the network head(%d), only %d blocks lag allowed",
			restURL,
			statistics.BlockHeight,
			headBlocksDiff,
			networkHeadHeight,
			HealthyBlocksThreshold,
		)
		return false
	}

	if statistics.DataNodeHeight > 0 {
		blocksDiff := statistics.BlockHeight - statistics.DataNodeHeight
		if statistics.DataNodeHeight < statistics.BlockHeight && blocksDiff > HealthyBlocksThreshold {
			logger.Sugar().Infof(
				"The %s endpoint unhealthy: data node is %d blocks behind core, only %d blocks lag allowed",
				restURL,
				blocksDiff,
				HealthyBlocksThreshold,
			)
			return false
		}
	}

	timeDiff := statistics.CurrentTime.Sub(statistics.VegaTime)
	if timeDiff > HealthyTimeThreshold {
		logger.Sugar().Infof(
			"The %s endpoint unhealthy: time lag is %s, only %s allowed",
			restURL,
			timeDiff.String(),
			HealthyTimeThreshold.String(),
		)
		return false
	}

	logger.Sugar().Infof("The %s endpoint is healthy", restURL)

	return true
}

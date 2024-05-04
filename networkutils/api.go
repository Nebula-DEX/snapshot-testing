package networkutils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
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

type statistics struct {
	BlockHeight    uint64
	DataNodeHeight uint64

	CurrentTime time.Time
	VegaTime    time.Time

	ChainID    string
	AppVersion string
}

func getStatistics(restURL string) (*statistics, error) {
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

	result := &statistics{
		BlockHeight:    blockHeight,
		DataNodeHeight: dataNodeHeight,
		CurrentTime:    currentTime,
		VegaTime:       vegaTime,

		ChainID:    rawResult.Statistics.ChainID,
		AppVersion: rawResult.Statistics.AppVersion,
	}

	return result, nil
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

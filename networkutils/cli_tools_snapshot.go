package networkutils

import (
	"fmt"
	"slices"
	"time"

	"github.com/vegaprotocol/snapshot-testing/tools"
)

type CliSnapshot struct {
	Height int64 `json:"height"`
}

func LocalSnapshotsRange(pathManager *PathManager) (int64, int64, error) {
	toolsSnapshotCmd := []string{
		"tools", "snapshot",
		"--home", pathManager.VegaHome(),
		"--output", "json",
	}

	response := struct {
		Snapshots []CliSnapshot
	}{}

	if err := tools.RetryRun(3, time.Second*5, func() error {
		if stdout, err := tools.ExecuteBinary(pathManager.VegaBin(), toolsSnapshotCmd, &response); err != nil {
			return fmt.Errorf("failed to execute vega tools snapshot(stdout: %s): %w", stdout, err)
		}

		return nil
	}); err != nil {
		return 0, 0, fmt.Errorf("failed to get snapshot from the cli: %w", err)
	}

	heightSlice := []int64{}
	for _, snapshot := range response.Snapshots {
		heightSlice = append(heightSlice, snapshot.Height)
	}

	return slices.Min(heightSlice), slices.Max(heightSlice), nil
}

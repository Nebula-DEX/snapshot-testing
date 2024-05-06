package components

import (
	"context"
	"fmt"
	"time"

	"github.com/vegaprotocol/snapshot-testing/networkutils"
	"go.uber.org/zap"
)

type HealthyStatus string

const (
	Healthy      HealthyStatus = "HEALTHY"
	MaybeHealthy HealthyStatus = "MAYBE"
	Unhealthy    HealthyStatus = "UNHEALTHY"
)

type event struct {
	time  time.Time
	event string
}

type localNodeStatus struct {
	started   time.Time // We started the watchdog thread
	firstSeen time.Time // We got first response from the /statistics for local node
	catchUp   time.Time // Node catch rest of the network up

	lagging time.Time // When node started lagging
	healthy time.Time // Last healthy event

	events []event
}

func (lns localNodeStatus) healthyStatus() HealthyStatus {
	// Node was up to date and did not lagging on the end
	if !lns.catchUp.IsZero() && lns.healthy.After(lns.lagging) {
		return Healthy
	}

	// Node caught rest of the network up at some point but then started lagging...
	if !lns.catchUp.IsZero() && !lns.healthy.After(lns.lagging) {
		return MaybeHealthy
	}

	return Unhealthy
}

func (lns localNodeStatus) unhealthyReason() string {
	if !lns.catchUp.IsZero() && lns.healthy.After(lns.lagging) {
		return ""
	}

	if !lns.catchUp.IsZero() && !lns.healthy.After(lns.lagging) {
		return "Node caught up at some point but then started lagging"
	}

	if lns.catchUp.IsZero() {
		return "Node never caught up rest of the network"
	}

	if lns.firstSeen.IsZero() {
		return "Node never returned valid response for the /statistics endpoint"
	}

	return "Unknown reason?????"
}

const (
	KeyNodeStatus      = "status"
	KeyUnhealthyReason = "reason"
	KeyStarted         = "test-startup"
	KeyFirstSeen       = "node-startup"
	KeyCatchUp         = "node-catch-up"
	KeyLastLag         = "node-last-lag"
	KeyLastHealthy     = "node-last-healthy"
	KeyCatchUpTime     = "catchup-duration"
)

// Prepare results that can be write into some file
func (lns localNodeStatus) toMap() map[string]interface{} {
	res := map[string]interface{}{
		KeyNodeStatus:      lns.healthyStatus(),
		KeyUnhealthyReason: lns.unhealthyReason(),
		KeyStarted:         lns.started.String(),
		KeyFirstSeen:       lns.firstSeen.String(),
		KeyCatchUp:         lns.catchUp.String(),
		KeyLastLag:         lns.lagging.String(),
		KeyLastHealthy:     lns.healthy.String(),
		KeyCatchUpTime:     "N/A",
	}

	if !lns.catchUp.IsZero() {
		res[KeyCatchUpTime] = lns.catchUp.Sub(lns.started).String()
	}

	return res
}

func (lns *localNodeStatus) PushEvent(str string) {
	if len(str) < 1 {
		return
	}

	lns.events = append(lns.events, event{
		time:  time.Now(),
		event: str,
	})
}

type watchdog struct {
	logger        *zap.Logger
	restEndpoints []string

	stop   context.CancelFunc
	status localNodeStatus

	lastReconciliation time.Time
}

func NewWatchdog(restEndpoints []string, mainLogger *zap.Logger) (Component, error) {
	if len(restEndpoints) < 1 {
		return nil, fmt.Errorf("at least one rest endpoint is required")
	}

	return &watchdog{
		restEndpoints:      restEndpoints,
		logger:             mainLogger,
		lastReconciliation: time.Now(),
	}, nil
}

func (w *watchdog) Name() string {
	return "watchdog"
}

func (w *watchdog) Result() map[string]interface{} {
	return w.status.toMap()
}

// Healthy implements Component.
func (w *watchdog) Healthy() (bool, error) {
	lastReconciliationDiff := time.Since(w.lastReconciliation)

	return lastReconciliationDiff < 30*time.Second, nil
}

// Start implements Component.
func (w *watchdog) Start(ctx context.Context) error {
	w.status.started = time.Now()

	watcherCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	w.stop = cancel

	ticker := time.NewTicker(5 * time.Second)
	for {
		w.lastReconciliation = time.Now()
		ticker.Reset(5 * time.Second)
		select {
		case <-ticker.C:
		// Someone finished execution
		case <-watcherCtx.Done():
			return nil
		}

		networkStatistics, err := networkutils.GetLatestStatistics(w.restEndpoints)
		if err != nil {
			w.logger.Sugar().Infof("Could not get valid response from any available REST endpoints: %v", w.restEndpoints)
			continue
		}

		nodeStatistics, err := networkutils.GetLatestStatistics([]string{"http://localhost:3008"})
		if err != nil {
			w.status.PushEvent("Node unhealthy")
			w.logger.Sugar().Info("Could not get valid response from local node(http://localhost:3008)")
			continue
		}

		if w.status.firstSeen.IsZero() {
			w.status.PushEvent("Node response from /statistics first seen")
			w.status.firstSeen = time.Now()
		}

		if nodeStatistics.BlockHeight < networkStatistics.BlockHeight {
			blocksDiff := networkStatistics.BlockHeight - nodeStatistics.BlockHeight
			if blocksDiff > 500 {
				msg := fmt.Sprintf(
					"Core blocks lag too big: local core(%d) is %d blocks behind rest of the network(%d), 500 blocks allowed",
					nodeStatistics.BlockHeight,
					blocksDiff,
					networkStatistics.BlockHeight,
				)
				w.status.PushEvent(msg)
				w.logger.Info(msg)

				w.status.lagging = time.Now()
				continue
			}
		}

		if nodeStatistics.DataNodeHeight < nodeStatistics.BlockHeight {
			blocksDiff := nodeStatistics.BlockHeight - nodeStatistics.DataNodeHeight

			if blocksDiff > 500 {
				msg := fmt.Sprintf(
					"Data node blocks lag too big: local data-node(%d) is %d blocks behind core(%d), 500 blocks allowed",
					nodeStatistics.DataNodeHeight,
					blocksDiff,
					nodeStatistics.BlockHeight,
				)
				w.status.PushEvent(msg)
				w.logger.Info(msg)

				w.status.lagging = time.Now()
				continue
			}
		}

		w.status.healthy = time.Now()
		if w.status.catchUp.IsZero() {
			msg := fmt.Sprintf("Node caught rest of the network up at block %d", nodeStatistics.BlockHeight)
			w.status.catchUp = time.Now()
			w.status.PushEvent(msg)
			w.logger.Info(msg)
		} else {
			msg := fmt.Sprintf("Local node is healthy, block is %d", nodeStatistics.BlockHeight)
			w.status.PushEvent(msg)
			w.logger.Info(msg)
		}
	}
}

// Stop implements Component.
func (w *watchdog) Stop(ctx context.Context) error {
	if w.stop != nil {
		w.stop()
	}

	return nil
}

// Stop implements Component.
func (w *watchdog) Cleanup(ctx context.Context) error {
	return w.Stop(ctx)
}

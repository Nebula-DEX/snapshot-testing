package components

import (
	"context"
	"fmt"
	"time"

	"github.com/vegaprotocol/snapshot-testing/networkutils"
	"go.uber.org/zap"
)

func Run(ctx context.Context, pathManager networkutils.PathManager, mainLogger *zap.Logger, components []Component) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mainLogger.Info("Running cleanup for all the components")
	// Start all of teh components
	for _, component := range components {
		mainLogger.Sugar().Infof("Starting cleanup for the %s component", component.Name())
		if err := component.Cleanup(ctx); err != nil {
			return fmt.Errorf("failed to cleanup the %s component: %w", component.Name(), err)
		}
	}

	mainLogger.Info("Starting the snapshot-testing components")
	// Start all of the components
	for idx, component := range components {
		mainLogger.Sugar().Infof("Starting the %s component", component.Name())
		go func(component Component) {
			if err := component.Start(ctx); err != nil {
				mainLogger.Fatal(fmt.Sprintf("failed to start the %s component:", component.Name()), zap.Error(err))
			}
		}(components[idx])
	}

	// Stop components when they are not needed anymore
	defer func(components []Component) {
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		for _, component := range components {
			if err := component.Stop(stopCtx); err != nil {
				mainLogger.Error(fmt.Sprintf("Failed to stop the %s component", component.Name()), zap.Error(err))
			}
		}
	}(components)

	ticker := time.NewTicker(30 * time.Second)

	for {
		ticker.Reset(30 * time.Second)
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return nil
		}
		allComponentsHealthy := true
		mainLogger.Info("Running health check")
		for _, component := range components {
			healthy, err := component.Healthy()
			if !healthy {
				allComponentsHealthy = false
				mainLogger.Error(fmt.Sprintf("The %s component is unhealthy", component.Name()), zap.Error(err))
				continue
			}

			mainLogger.Sugar().Infof("The %s component is healthy", component.Name())
		}

		if !allComponentsHealthy {
			return fmt.Errorf("one or more test components failed")
		}
	}
}

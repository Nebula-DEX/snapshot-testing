package components

import (
	"context"
)

type ComponentResults map[string]any

type Component interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Healthy() (bool, error)
	Cleanup(ctx context.Context) error
	Result() ComponentResults
}

func MergeResults(results ...ComponentResults) ComponentResults {
	finalResult := ComponentResults{}

	for idx, singleComponentResults := range results {
		for k := range singleComponentResults {
			finalResult[k] = results[idx][k]
		}
	}

	return finalResult
}

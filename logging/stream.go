package logging

import (
	"bufio"
	"fmt"
	"io"

	"go.uber.org/zap"
)

func StreamLogs(source io.Reader, out *zap.Logger) error {
	if source == nil {
		return fmt.Errorf("source stream is nil")
	}

	scanner := bufio.NewScanner(source)
	for scanner.Scan() {
		out.Info(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to stream logs: %w", err)
	}

	return nil
}

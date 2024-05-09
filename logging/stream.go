package logging

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"

	"go.uber.org/zap"
)

const AppendExtraLogLinesAfterFailureFoundCount = 3

type ExtraInfo struct {
	mut      sync.Mutex
	logLines []string
}

func (ei *ExtraInfo) String() string {
	return strings.Join(ei.logLines, "\n")
}
func (ei ExtraInfo) Empty() bool {
	return len(ei.logLines) < 1
}

func NewExtraInfo() ExtraInfo {
	return ExtraInfo{}
}

func StreamLogs(source io.Reader, out *zap.Logger, extraResults *ExtraInfo) error {
	if source == nil {
		return fmt.Errorf("source stream is nil")
	}

	appendExtraLines := 0
	scanner := bufio.NewScanner(source)
	for scanner.Scan() {
		text := scanner.Text()

		// Search for panics, app hash errors, etc if someone requested for it by passing the ExtraInfo pointer
		if extraResults != nil {
			if foundFailure(text) {
				appendExtraLines = AppendExtraLogLinesAfterFailureFoundCount
			}

			if appendExtraLines > 0 {
				extraResults.mut.Lock()
				extraResults.logLines = append(extraResults.logLines, text)
				extraResults.mut.Unlock()

				appendExtraLines = appendExtraLines - 1
			}
		}

		out.Info(text)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to stream logs: %w", err)
	}

	return nil
}

// TODO: We should pass pointer here...
func foundFailure(logLine string) bool {
	logLine = strings.ToLower(logLine)

	return strings.Contains(logLine, "panic") ||
		strings.Contains(logLine, "consensus failure") ||
		strings.Contains(logLine, "invalid memory") ||
		strings.Contains(logLine, "wrong block.header.lastresultshash")
}

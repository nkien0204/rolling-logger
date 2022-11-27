package test

import (
	"testing"

	"github.com/nkien0204/rolling-logger/rolling"
	"go.uber.org/zap"
)

func BenchmarkNewLogger(t *testing.B) {
	for i := 0; i < t.N; i++ {
		logger := rolling.New()
		writeLog(logger, i)
	}
}

func writeLog(logger *zap.Logger, i int) {
	logger.Info("write to file", zap.Int("value", i))
}

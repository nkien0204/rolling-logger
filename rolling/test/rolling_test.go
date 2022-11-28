package test

import (
	"os"
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

func TestNewLogger(t *testing.T) {
	inputEnv := map[string]string{
		"LOG_ROTATION_TIME": "",
		"LOG_INFO_DIR":      "",
		"LOG_INFO_NAME":     "",
		"LOG_DEBUG_DIR":     "",
		"LOG_DEBUG_NAME":    "",
	}
	for k, v := range inputEnv {
		if err := os.Setenv(k, v); err != nil {
			t.Errorf("setenv failed %s:%s", k, v)
		}
	}
	logger := rolling.New()
	logger.Info("info")
	logger.Error("error")
	logger.Debug("debug")
}

package rolling

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/strftime"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const DEFAULT_DIR string = "log"
const DEFAULT_INFO_NAME string = "logger.log"
const DEFAULT_DEBUG_NAME string = "logger-debug.log"
const (
	DAY_ROTATION  string = "day"
	HOUR_ROTATION string = "hour"
	MIN_ROTATION  string = "min"
)

type rolling struct {
	filename     string
	pattern      *strftime.Strftime
	rotationTime time.Duration
	fileWriter   *os.File
}

var logger *zap.Logger
var once sync.Once

/*
Create new logger only at the first time.
*/
func New() *zap.Logger {
	once.Do(func() {
		logger = initLogger()
	})
	return logger
}

func initLogger() *zap.Logger {
	config := zap.Config{
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "msg",
			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,
			TimeKey:     "ts",
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("2006-01-02 15:04:05"))
			},
			CallerKey:    "file",
			EncodeCaller: zapcore.ShortCallerEncoder,
			EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendInt64(int64(d) / 1000000)
			},
		},
	}
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.InfoLevel
	})
	debugLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.InfoLevel
	})

	var infoFilenamePattern string
	var debugFilenamePattern string
	var rotationTime time.Duration
	switch os.Getenv("LOG_ROTATION_TIME") {
	case DAY_ROTATION:
		infoFilenamePattern, debugFilenamePattern = handleRotation("%Y-%m-%d")
		rotationTime = time.Hour * 24
	case HOUR_ROTATION:
		infoFilenamePattern, debugFilenamePattern = handleRotation("%Y-%m-%d-%H")
		rotationTime = time.Hour
	case MIN_ROTATION:
		infoFilenamePattern, debugFilenamePattern = handleRotation("%Y-%m-%d-%H-%M")
		rotationTime = time.Minute
	default:
		infoFilenamePattern, debugFilenamePattern = handleRotation("%Y-%m-%d-%H")
		rotationTime = time.Hour
	}
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(config.EncoderConfig), zapcore.AddSync(newRolling(infoFilenamePattern, rotationTime)), infoLevel),
		zapcore.NewCore(zapcore.NewJSONEncoder(config.EncoderConfig), zapcore.AddSync(newRolling(debugFilenamePattern, rotationTime)), debugLevel),
	)
	return zap.New(core, zap.AddCaller())
}

func handleRotation(timeFormat string) (infoPattern string, debugPattern string) {
	infoDir, infoName := getPatternFromEnv("INFO")
	infoPattern = fmt.Sprintf("%s/%s.%s", infoDir, timeFormat, infoName)

	debugDir, debugName := getPatternFromEnv("DEBUG")
	debugPattern = fmt.Sprintf("%s/%s.%s", debugDir, timeFormat, debugName)

	return infoPattern, debugPattern
}

func getPatternFromEnv(level string) (dirPattern, namePattern string) {
	switch level {
	case "INFO":
		dirPattern = strings.TrimSpace(os.Getenv("LOG_INFO_DIR"))
		if dirPattern == "" {
			dirPattern = DEFAULT_DIR
		}
		namePattern = strings.TrimSpace(os.Getenv("LOG_INFO_NAME"))
		if namePattern == "" {
			namePattern = DEFAULT_INFO_NAME
		}
	case "DEBUG":
		dirPattern = strings.TrimSpace(os.Getenv("LOG_DEBUG_DIR"))
		if dirPattern == "" {
			dirPattern = DEFAULT_DIR
		}
		namePattern = strings.TrimSpace(os.Getenv("LOG_DEBUG_NAME"))
		if namePattern == "" {
			namePattern = DEFAULT_DEBUG_NAME
		}
	default:
		// No need to handle this case right now!
	}
	return dirPattern, namePattern
}

func newRolling(filename string, rotationTime time.Duration) *rolling {
	pattern, err := strftime.New(filename)
	if err != nil {
		panic(err)
	}
	return &rolling{
		filename:     filename,
		pattern:      pattern,
		rotationTime: rotationTime,
		fileWriter:   nil,
	}
}

func (r *rolling) Write(p []byte) (n int, err error) {
	base := time.Now().Truncate(r.rotationTime)
	newFilename := r.pattern.FormatString(base)
	if r.filename != newFilename {
		if r.fileWriter != nil {
			r.fileWriter.Close()
		}

		dirname := filepath.Dir(newFilename)
		if err := os.MkdirAll(dirname, 0755); err != nil {
			return 0, err
		}
		r.fileWriter, err = os.OpenFile(newFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
		r.filename = newFilename
	}

	return r.fileWriter.Write(p)
}

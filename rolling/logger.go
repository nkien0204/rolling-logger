package rolling

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/strftime"
	"github.com/nkien0204/rolling-logger/configuration"
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

type consoleLogger struct{}

type fileLogger struct {
	filename        string
	symlinkFileName string
	dir             string
	pattern         *strftime.Strftime
	rotationTime    time.Duration
	fileWriter      *os.File
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
	cfg := configuration.GetConfigs()
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
	logLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= convertLogLevel(cfg.Log.LogLevelMin) && lvl <= convertLogLevel(cfg.Log.LogLevelMax)
	})

	var core zapcore.Core
	switch cfg.Log.Output {
	case configuration.ConsoleOutputLog:
		core = handleInitConsoleLogger(cfg, config, logLevel)
	case configuration.FileOutputLog:
		core = handleInitFileLogger(cfg, config, logLevel)
	default:
		core = handleInitConsoleLogger(cfg, config, logLevel)
	}
	return zap.New(core, zap.AddCaller())
}

func handleInitFileLogger(cfg *configuration.Cfg, config zap.Config, logLevel zap.LevelEnablerFunc) zapcore.Core {
	var rotationTime time.Duration
	rollingLogger := initRolling()

	switch cfg.Log.RotationTime {
	case DAY_ROTATION:
		rollingLogger.handleRotation("%Y-%m-%d")
		rotationTime = time.Hour * 24
	case HOUR_ROTATION:
		rollingLogger.handleRotation("%Y-%m-%d-%H")
		rotationTime = time.Hour
	case MIN_ROTATION:
		rollingLogger.handleRotation("%Y-%m-%d-%H-%M")
		rotationTime = time.Minute
	default:
		rollingLogger.handleRotation("%Y-%m-%d-%H")
		rotationTime = time.Hour
	}
	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(config.EncoderConfig),
			zapcore.AddSync(rollingLogger.setupRolling(DEFAULT_INFO_NAME, rotationTime)),
			logLevel,
		),
	)
	return core
}

func handleInitConsoleLogger(cfg *configuration.Cfg, config zap.Config, logLevel zap.LevelEnablerFunc) zapcore.Core {
	consoleLogger := initConsoleLogger()
	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(config.EncoderConfig),
			zapcore.AddSync(consoleLogger),
			logLevel,
		),
	)
	return core
}

func convertLogLevel(logLevelStr string) zapcore.Level {
	switch logLevelStr {
	case zapcore.DebugLevel.String():
		return zapcore.DebugLevel
	case zapcore.InfoLevel.String():
		return zapcore.InfoLevel
	case zapcore.WarnLevel.String():
		return zapcore.WarnLevel
	case zapcore.ErrorLevel.String():
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func initConsoleLogger() *consoleLogger {
	return &consoleLogger{}
}

func initRolling() *fileLogger {
	return &fileLogger{
		filename:        "",
		symlinkFileName: "",
		dir:             "",
		pattern:         nil,
		rotationTime:    time.Hour,
		fileWriter:      nil,
	}
}

func (r *fileLogger) handleRotation(timeFormat string) {
	dir, filename := r.getPatternFromEnv()
	filenamePattern := fmt.Sprintf("%s.%s", timeFormat, filename)

	r.dir = dir
	r.filename = filenamePattern
}

func (r *fileLogger) getPatternFromEnv() (dirPattern, namePattern string) {
	configs := configuration.GetConfigs()
	if configs == nil {
		return
	}
	dirPattern = strings.TrimSpace(configs.Log.LogDir)
	if dirPattern == "" {
		dirPattern = DEFAULT_DIR
	}
	namePattern = strings.TrimSpace(configs.Log.LogFileName)
	if namePattern == "" {
		namePattern = DEFAULT_INFO_NAME
	}
	return dirPattern, namePattern
}

func (r *fileLogger) setupRolling(symlinkFileName string, rotationTime time.Duration) *fileLogger {
	pattern, err := strftime.New(r.filename)
	if err != nil {
		panic(err)
	}

	r.symlinkFileName = symlinkFileName
	r.pattern = pattern
	r.rotationTime = rotationTime
	r.fileWriter = nil

	return r
}

func (r *fileLogger) Write(p []byte) (n int, err error) {
	base := time.Now().Truncate(r.rotationTime)
	newFilename := r.pattern.FormatString(base)
	if r.filename != newFilename {
		if r.fileWriter != nil {
			r.fileWriter.Close()
		}

		if err := os.MkdirAll(r.dir, 0755); err != nil {
			return 0, err
		}
		newFileStr := filepath.Join(r.dir, newFilename)
		r.fileWriter, err = os.OpenFile(newFileStr, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
		r.filename = newFilename

		if err := r.createSymlink(); err != nil {
			fmt.Println("error:", err) // no need to return
		}
	}

	return r.fileWriter.Write(p)
}

func (c *consoleLogger) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

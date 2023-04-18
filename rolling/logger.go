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

type rolling struct {
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

	var rotationTime time.Duration
	infoRolling := initRolling()
	debugRolling := initRolling()

	switch os.Getenv("LOG_ROTATION_TIME") {
	case DAY_ROTATION:
		infoRolling.handleRotation("%Y-%m-%d", "INFO")
		debugRolling.handleRotation("%Y-%m-%d", "DEBUG")
		rotationTime = time.Hour * 24
	case HOUR_ROTATION:
		infoRolling.handleRotation("%Y-%m-%d-%H", "INFO")
		debugRolling.handleRotation("%Y-%m-%d-%H", "DEBUG")
		rotationTime = time.Hour
	case MIN_ROTATION:
		infoRolling.handleRotation("%Y-%m-%d-%H-%M", "INFO")
		debugRolling.handleRotation("%Y-%m-%d-%H-%M", "DEBUG")
		rotationTime = time.Minute
	default:
		infoRolling.handleRotation("%Y-%m-%d-%H", "INFO")
		debugRolling.handleRotation("%Y-%m-%d-%H", "DEBUG")
		rotationTime = time.Hour
	}
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(config.EncoderConfig), zapcore.AddSync(infoRolling.setupRolling(DEFAULT_INFO_NAME, rotationTime)), infoLevel),
		zapcore.NewCore(zapcore.NewJSONEncoder(config.EncoderConfig), zapcore.AddSync(debugRolling.setupRolling(DEFAULT_DEBUG_NAME, rotationTime)), debugLevel),
	)
	return zap.New(core, zap.AddCaller())
}

func initRolling() *rolling {
	return &rolling{
		filename:        "",
		symlinkFileName: "",
		dir:             "",
		pattern:         nil,
		rotationTime:    time.Hour,
		fileWriter:      nil,
	}
}

func (r *rolling) handleRotation(timeFormat string, level string) {
	dir, filename := r.getPatternFromEnv(level)
	filenamePattern := fmt.Sprintf("%s.%s", timeFormat, filename)

	r.dir = dir
	r.filename = filenamePattern
}

func (r *rolling) getPatternFromEnv(level string) (dirPattern, namePattern string) {
	configs := configuration.GetConfigs()
	switch level {
	case "INFO":
		dirPattern = strings.TrimSpace(configs.Log.LogInfoDir)
		if dirPattern == "" {
			dirPattern = DEFAULT_DIR
		}
		namePattern = strings.TrimSpace(configs.Log.LogInfoFileName)
		if namePattern == "" {
			namePattern = DEFAULT_INFO_NAME
		}
	case "DEBUG":
		dirPattern = strings.TrimSpace(configs.Log.LogDebugDir)
		if dirPattern == "" {
			dirPattern = DEFAULT_DIR
		}
		namePattern = strings.TrimSpace(configs.Log.LogDebugFileName)
		if namePattern == "" {
			namePattern = DEFAULT_DEBUG_NAME
		}
	default:
		// No need to handle this case right now!
	}
	return dirPattern, namePattern
}

func (r *rolling) setupRolling(symlinkFileName string, rotationTime time.Duration) *rolling {
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

func (r *rolling) Write(p []byte) (n int, err error) {
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

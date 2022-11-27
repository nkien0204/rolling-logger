package rolling

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lestrrat-go/strftime"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const DEFAULT_LOG string = "log/logger.log"
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
		Level:    zap.NewAtomicLevelAt(zap.InfoLevel),
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
	var filenamePattern string
	var rotationTime time.Duration
	switch os.Getenv("LOG_ROTATION_TIME") {
	case DAY_ROTATION:
		filenamePattern = handleRotation("%Y-%m-%d")
		rotationTime = time.Hour * 24
	case HOUR_ROTATION:
		filenamePattern = handleRotation("%Y-%m-%d-%H")
		rotationTime = time.Hour
	case MIN_ROTATION:
		filenamePattern = handleRotation("%Y-%m-%d-%H-%M")
		rotationTime = time.Minute
	default:
		filenamePattern = handleRotation("%Y-%m-%d-%H")
		rotationTime = time.Hour
	}
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(config.EncoderConfig), zapcore.AddSync(newRolling(filenamePattern, rotationTime)), zapcore.InfoLevel),
	)
	return zap.New(core, zap.AddCaller())
}

func handleRotation(timeFormat string) string {
	pattern := os.Getenv("LOG_FILE")
	if pattern == "" {
		return fmt.Sprintf("%s.%s", DEFAULT_LOG, timeFormat)
	}
	return fmt.Sprintf("%s.%s", pattern, timeFormat)
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

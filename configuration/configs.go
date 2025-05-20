package configuration

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

const CONFIG_FILENAME string = "config.yaml"

type OutputType string

const (
	ConsoleOutputLog OutputType = "console"
	FileOutputLog    OutputType = "file"
)

type Cfg struct {
	Log LogConfig
}

type LogConfig struct {
	Output       OutputType `yaml:"output"`
	LogLevelMin  string     `yaml:"log_level_min"`
	LogLevelMax  string     `yaml:"log_level_max"`
	RotationTime string     `yaml:"log_rotation_time"`
	LogDir       string     `yaml:"log_dir"`
	LogFileName  string     `yaml:"log_file_name"`
}

var config *Cfg
var once sync.Once

// Singleton pattern
func GetConfigs() *Cfg {
	once.Do(func() {
		var err error
		if config, err = newConfigs(); err != nil {
			config = &Cfg{
				Log: LogConfig{
					Output:      ConsoleOutputLog,
					LogLevelMin: zapcore.InfoLevel.String(),
					LogLevelMax: zapcore.InfoLevel.String(),
				},
			}
		}
	})
	return config
}

func newConfigs() (*Cfg, error) {
	conf, err := readConf(CONFIG_FILENAME)
	if err != nil {
		return nil, err
	}

	switch conf.Log.Output {
	case ConsoleOutputLog, FileOutputLog:
	default:
		return nil, fmt.Errorf("invalid log output type: %s", conf.Log.Output)
	}

	switch conf.Log.LogLevelMin {
	case zapcore.DebugLevel.String(),
		zapcore.InfoLevel.String(),
		zapcore.WarnLevel.String(),
		zapcore.ErrorLevel.String():
	default:
		return nil, fmt.Errorf("invalid log level: %s", conf.Log.LogLevelMin)
	}

	switch conf.Log.LogLevelMax {
	case zapcore.DebugLevel.String(),
		zapcore.InfoLevel.String(),
		zapcore.WarnLevel.String(),
		zapcore.ErrorLevel.String():
	default:
		return nil, fmt.Errorf("invalid log level: %s", conf.Log.LogLevelMax)
	}
	return conf, nil
}

func readConf(filename string) (*Cfg, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &Cfg{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %w", filename, err)
	}

	return c, err
}

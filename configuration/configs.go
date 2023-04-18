package configuration

import (
	"fmt"
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v2"
)

const CONFIG_FILENAME string = "config.yaml"

type Cfg struct {
	Log LogConfig
}

type LogConfig struct {
	RotationTime     string `yaml:"log_rotation_time"`
	LogInfoDir       string `yaml:"log_info_dir"`
	LogInfoFileName  string `yaml:"log_info_name"`
	LogDebugDir      string `yaml:"log_debug_dir"`
	LogDebugFileName string `yaml:"log_debug_name"`
}

var config *Cfg
var once sync.Once

// Singleton pattern
func GetConfigs() *Cfg {
	once.Do(func() {
		var err error
		if config, err = newConfigs(); err != nil {
			panic(err)
		}
	})
	return config
}

func newConfigs() (*Cfg, error) {
	return readConf(CONFIG_FILENAME)
}

func readConf(filename string) (*Cfg, error) {
	buf, err := ioutil.ReadFile(filename)
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

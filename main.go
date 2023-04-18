package main

import (
	"github.com/nkien0204/rolling-logger/rolling"
)

func main() {
	logger := rolling.New()
	defer logger.Sync()

	logger.Info("hello logger")
	logger.Error("got error")
	logger.Debug("this is debug")
}

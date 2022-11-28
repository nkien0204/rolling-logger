package main

import (
	"fmt"

	"github.com/nkien0204/rolling-logger/rolling"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("load env error: ", err.Error())
		panic(err)
	}
	logger := rolling.New()
	defer logger.Sync()

	logger.Info("hello logger")
	logger.Error("got error")
	logger.Debug("this is debug")
}

package main

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/nkien0204/rolling-logger/rolling"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("load env error: ", err.Error())
		panic(err)
	}
	logger := rolling.New()
	defer logger.Sync()

	go logger.Info("hello logger")
	logger.Error("got error")
	time.Sleep(time.Second)
}

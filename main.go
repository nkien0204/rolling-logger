package main

import (
	"fmt"
	"os"
	"rolling-log/logger"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// logger, err := config.Build()
	// if err != nil {
	// 	log.Fatalf("can't initialize zap logger: %v", err)
	// }
	err := godotenv.Load()
	if err != nil {
		fmt.Println("load env error: ", err.Error())
		os.Exit(1)
	}
	logger := logger.New()
	for {
		logger.Info("hello logger", zap.String("name", "kiennt"))
		logger.Error("got error")
		time.Sleep(5 * time.Second)
	}
	// defer logger.Sync()
}

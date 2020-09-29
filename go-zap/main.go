package main

import (
	"go.uber.org/zap"
	"log"
)

var logger *zap.Logger

func init() {
	logger = zap.NewExample()
	//log, _ = zap.NewDevelopment()
	//log, _ = zap.NewProduction()

	defer logger.Sync() // flush buffers
	
	std := zap.NewStdLog(logger)
	std.Print("standard logger wrapper")
}

func main() {

	log.Println("Hello world!")

	//log.Debug("Debug message")
	//log.Info("Info message")
	//log.Warn("Warning message")
	//log.Error("Error message")

	//log.Panic("Panic message")
	//log.Fatal("Fatal message")

	//var (
	//	index int64 = 10
	//	requestID = "127.0.0.1"
	//	userIP = "43.256.56.1"
	//)
	//
	//log.Error("Failed to send event",
	//	zap.String("request_id", requestID),
	//	zap.String("user_ip", userIP),
	//	zap.Int64("index", index),
	//	zap.Duration("request_time", time.Second),
	//)
	//
	//// 记录层级关系
	//// 方式1
	//log.Info("tracked some metrics",
	//	zap.Namespace("metrics"),
	//	zap.Int64("counter", 1),
	//	zap.String("name", "m1"),
	//)
	//
	//// 方式2
	//logger := log.With(
	//	zap.Namespace("metrics"),
	//	zap.Int("counter", 1),
	//	zap.String("name", "m2"),
	//)
	//logger.Info("tracked some metrics")
}

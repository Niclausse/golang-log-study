package main

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})

	log.SetOutput(os.Stdout)

	log.SetLevel(log.TraceLevel)

	// 报告被调用位置（文件及行号）!!! 开启此功能对性能开销大
	// 考虑通过Hook的方式定制需要此功能的日志级别；比如对ERROR级别及以上的日志报告文件名及行号
	log.SetReportCaller(true)
}


func main() {

	log.Traceln("Trace message")

	log.Debugln("Debug message")

	log.Println("Info message by Print func")
	log.Infoln("Info message by Info func")

	log.Warningln("Warning message")

	log.Errorln("Error message")

	var (
		event = "test event"
		topic = "test topic"
		key = 10

		requestID = "127.0.0.1"
		userIP = "43.256.56.1"
	)

	log.Errorf("Failed to send event %s to topic %s with key %d", event, topic, key)

	log.WithFields(log.Fields{
		"event": event,
		"topic": topic,
		"key": key,
	}).Errorf("Failed to send event")


	logEntry := log.WithFields(log.Fields{"request_id": requestID, "user_ip": userIP})
	logEntry.Infoln("Something happened on that request")
}


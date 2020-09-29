package main

import (
	"flag"
	"github.com/golang/glog"
)

func main() {
	flag.Parse()
	defer glog.Flush() // flush all pending log I/O

	glog.Infoln("This is info message")
	glog.Infof("This is info message: %+v", 12345)

	glog.Warningln("This is warning message")
	glog.Warningf("This is warning message: %+v", 12345)

	glog.Errorln("This is error message")
	glog.Errorf("This is error message: %+v", 12345)

	glog.Fatal("This is fatal error")
}

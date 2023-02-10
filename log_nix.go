//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"log/syslog"
)

var (
	logger *syslog.Writer
)

func configurelogger() {
	if logToSyslog {
		logger, err = syslog.New(syslog.LOG_DAEMON|syslog.LOG_INFO, "go-icmp-status")
		if err != nil {
			fmt.Print(err)
		}
	}
}

func doLogPrintf(format string, v ...interface{}) {
	if logToSyslog {
		errl := logger.Info(format)
		if errl != nil {
			fmt.Print(errl)
		}
	}
}

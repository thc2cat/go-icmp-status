//go:build !windows
// +build !windows

package main

import (
	"log/syslog"
)

var (
	logger *syslog.Writer
)

func configurelogger() {
	if logToSyslog {
		logger, err = syslog.New(syslog.LOG_DAEMON|syslog.LOG_INFO, "go-icmp-status")
	}
}

func doLogPrintf(format string, v ...interface{}) {
	if logToSyslog {
		logger.Info(format)
	}
}

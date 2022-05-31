package main

// Original Source :
// https://github.com/digineo/go-ping/tree/master/cmd/ping-monitor

// History :
//  v0.2 : loosing packets message status, seconds in timestamps
//  v0.3 : using fathi/color
//  v0.4 : packet loss summary
//  v0.5 : add syslog reporting for long term survey
//  v0.6 : -I show resolved IPs, -t allow 1 packet loss tolerance
//  v0.7 : -stopAfter delay option for timed execution
//  v0.8 : moved defered stops before reports
//
// Author of additional code : T.CAILLET.

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/digineo/go-ping"
	"github.com/digineo/go-ping/monitor"
	"github.com/fatih/color"
)

type Stats struct {
	Received int
	Sent     int
}

var (
	pingInterval        = 1 * time.Second
	pingTimeout         = 3 * time.Second
	reportInterval      = 5 * time.Second
	stopAfter           = 365 * 24 * time.Hour
	reportLoss          = false
	logToSyslog         = false
	beTolerant          = false
	showIp              = false
	size           uint = 56
	pinger         *ping.Pinger
	err            error

	targets    []string
	isAlive    = make(map[string]bool)
	displayed  = make(map[string]bool)
	hoststats  = make(map[string]*Stats)
	dateFormat = "2006-01-02 15:04:05"
)

func main() {

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[options] host [host [...]]")
		flag.PrintDefaults()
	}

	flag.DurationVar(&pingInterval, "pingInterval", pingInterval, "interval for ICMP echo requests")
	flag.DurationVar(&pingTimeout, "pingTimeout", pingTimeout, "timeout for ICMP echo request")
	flag.DurationVar(&reportInterval, "reportInterval", reportInterval, "interval for reports")
	flag.UintVar(&size, "size", size, "size of additional payload data")
	flag.BoolVar(&reportLoss, "reportLoss", reportLoss, "report loss summary")
	flag.BoolVar(&logToSyslog, "logToSyslog", logToSyslog, "log events to syslog")
	flag.BoolVar(&beTolerant, "t", beTolerant, "be tolerant, allow 1 packet loss per check")
	flag.BoolVar(&showIp, "showIp", showIp, "show monitored ips resolution")
	flag.DurationVar(&stopAfter, "stopAfter", stopAfter, "stop monitoring after this interval")

	flag.Parse()

	if n := flag.NArg(); n == 0 { // Targets empty?
		flag.Usage()
		os.Exit(1)
	} else if n > 1024 { // Too much icmp may be problematic for some OS
		fmt.Printf("Too many targets : %d > 1024 max\n", n)
		os.Exit(1)
	}

	// Bind to sockets
	if pinger, err = ping.New("0.0.0.0", "::"); err != nil {
		fmt.Printf("Unable to bind: %s\nRunning as root?\n", err)
		os.Exit(2)
	}
	pinger.SetPayloadSize(uint16(size))
	// defer pinger.Close()

	// Create checker
	checker := monitor.New(pinger, pingInterval, pingTimeout)
	// defer checker.Stop()

	// Add targets
	targets = flag.Args()
	for i, target := range targets {
		ipAddr, err := net.ResolveIPAddr("", target)
		if err != nil {
			fmt.Printf("invalid target '%s': %s\n", target, err)
			continue
		}
		if showIp {
			fmt.Printf("ip adress monitored for host %s will be %s\n",
				target, ipAddr.String())
		}
		checker.AddTargetDelayed(string([]byte{byte(i)}), *ipAddr,
			10*time.Millisecond*time.Duration(i))
		isAlive[target] = true // Considers hosts are alive.
		hoststats[target] = new(Stats)
	}

	// Start report routine
	ticker := time.NewTicker(reportInterval)
	// defer ticker.Stop()

	start := time.Now()
	if logToSyslog {
		configurelogger()
	}

	go func() {
		for range ticker.C {
			for i, metrics := range checker.ExportAndClear() {

				host := targets[[]byte(i)[0]]
				hoststats[host].Received += metrics.PacketsSent - metrics.PacketsLost
				hoststats[host].Sent += metrics.PacketsSent

				alive := (metrics.PacketsSent - metrics.PacketsLost) > 0
				loosing := (metrics.PacketsSent - metrics.PacketsLost) != metrics.PacketsSent

				if (!displayed[host]) || (isAlive[host] != alive) || (alive && loosing) {
					stamp := time.Now().Format("2006-02-01 15:04:05")
					percent := float32(hoststats[host].Received) / float32(hoststats[host].Sent) * 100
					switch {

					case alive && metrics.PacketsLost == 0:
						msg := fmt.Sprintf("%s is up", host)
						fmt.Fprintf(color.Output, "%s %s\n", stamp,
							color.GreenString(msg))
						if logToSyslog {
							doLogPrintf(msg)
						}

					case alive && beTolerant && metrics.PacketsLost == 1:

					case alive && metrics.PacketsLost != 0:
						msg := fmt.Sprintf("%s incomplete reply [%d/%d/%.1f%%]",
							host, metrics.PacketsSent-metrics.PacketsLost,
							metrics.PacketsSent, percent)
						fmt.Fprintf(color.Output, "%s %s\n", stamp,
							color.YellowString(msg))
						if logToSyslog {
							doLogPrintf(msg)
						}

					case !alive:
						msg := fmt.Sprintf("%s is down", host)
						fmt.Fprintf(color.Output, "%s %s\n", stamp,
							color.RedString(msg))
						if logToSyslog {
							doLogPrintf(msg)
						}

					}

					isAlive[host], displayed[host] = alive, true
				}
			}
		}
	}()

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ch:
	case <-time.After(stopAfter):
	}

	ticker.Stop()
	checker.Stop()
	pinger.Close()

	if reportLoss {
		end := time.Now()
		fmt.Printf("\ngo-icmp-status summary %s to %s:\n",
			start.Format(dateFormat), end.Format(dateFormat))
		// Summary
		for host := range hoststats {
			if hoststats[host].Sent != 0 && (hoststats[host].Sent-hoststats[host].Received != 0) {
				num := 100. - float32(hoststats[host].Received)/float32(hoststats[host].Sent)*100
				msg := fmt.Sprintf("  received %3d/%3d packets %3.1f %% loss for %s\n",
					hoststats[host].Received, hoststats[host].Sent, num, host)
				switch {
				case num > 5.:
					fmt.Fprintf(color.Output, "%s", color.RedString(msg))
				case num > 0.1:
					fmt.Fprintf(color.Output, "%s", color.YellowString(msg))
				default:
					fmt.Fprintf(color.Output, "%s", msg)
				}
			}
		}
	}

}

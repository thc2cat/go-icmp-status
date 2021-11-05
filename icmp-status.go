package main

// adapted from :
// https://github.com/digineo/go-ping/tree/master/cmd/ping-monitor

// History :
//  v0.2 : loosing packets message status, seconds in timestamps
//  v0.3 : using fathi/color
//  v0.4 : packet loss summary
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
	reportInterval      = 3 * time.Second
	reportSummary       = false
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
	flag.BoolVar(&reportSummary, "S", reportSummary, "report summary")
	flag.Parse()

	if n := flag.NArg(); n == 0 { // Targets empty?
		flag.Usage()
		os.Exit(1)
	} else if n > int(^byte(0)) { // Too many targets?

		fmt.Println("Too many targets")
		os.Exit(1)
	}

	// Bind to sockets
	if pinger, err = ping.New("0.0.0.0", "::"); err != nil {
		fmt.Printf("Unable to bind: %s\nRunning as root?\n", err)
		os.Exit(2)
	}
	pinger.SetPayloadSize(uint16(size))
	defer pinger.Close()

	// Create monitor
	monitor := monitor.New(pinger, pingInterval, pingTimeout)
	defer monitor.Stop()

	// Add targets
	targets = flag.Args()
	for i, target := range targets {
		ipAddr, err := net.ResolveIPAddr("", target)
		if err != nil {
			fmt.Printf("invalid target '%s': %s", target, err)
			continue
		}
		monitor.AddTargetDelayed(string([]byte{byte(i)}), *ipAddr,
			10*time.Millisecond*time.Duration(i))
		isAlive[target] = true

		hoststats[target] = new(Stats)
	}

	// Start report routine
	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()

	start := time.Now()

	go func() {
		for range ticker.C {
			for i, metrics := range monitor.ExportAndClear() {

				host := targets[[]byte(i)[0]]
				hoststats[host].Received += metrics.PacketsSent - metrics.PacketsLost
				hoststats[host].Sent += metrics.PacketsSent

				// tmp := hoststats[host]
				// tmp.Received +=
				// tmp.Sent +=
				// hoststats[host] = tmp

				alive := (metrics.PacketsSent - metrics.PacketsLost) > 0
				loosing := (metrics.PacketsSent - metrics.PacketsLost) != metrics.PacketsSent

				if (!displayed[host]) || (isAlive[host] != alive) || (alive && loosing) {
					stamp := time.Now().Format("2006-02-01 15:04:05")
					percent := int(float32(hoststats[host].Received) / float32(hoststats[host].Sent) * 100)
					switch {

					case alive && metrics.PacketsLost == 0:
						fmt.Fprintf(color.Output, "%s %s", stamp,
							color.GreenString(fmt.Sprintf("%s is up\n", host)))

					case alive && metrics.PacketsLost != 0:
						fmt.Fprintf(color.Output, "%s %s", stamp,
							color.YellowString(fmt.Sprintf("%s incomplete reply [%d/%d/%d%%]\n",
								host, metrics.PacketsSent-metrics.PacketsLost, metrics.PacketsSent,
								percent)))

					case !alive:
						fmt.Fprintf(color.Output, "%s %s", stamp,
							color.RedString(fmt.Sprintf("%s is down\n", host)))

					}

					isAlive[host], displayed[host] = alive, true
				}
			}
		}
	}()

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	// fmt.Println("received", <-ch)
	<-ch

	if reportSummary {
		end := time.Now()
		fmt.Printf("\ngo-icmp-status summary %s to %s:\n", start.Format(dateFormat), end.Format(dateFormat))
		// Summary
		for host := range hoststats {
			if hoststats[host].Sent != 0 {
				fmt.Printf("  received %3d/%3d packets %3d %% loss for %s\n",
					hoststats[host].Received, hoststats[host].Sent,
					100-int(float32(hoststats[host].Received)/float32(hoststats[host].Sent)*100),
					host)
			}
		}
	}

}

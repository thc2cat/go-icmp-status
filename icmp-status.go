package main

// History :
//  v0.2 : loosing packets message status, seconds in timestamps
//  v0.3 using fathi/color

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

var (
	pingInterval        = 1 * time.Second
	pingTimeout         = 3 * time.Second
	reportInterval      = 3 * time.Second
	size           uint = 56
	pinger         *ping.Pinger
	err            error

	targets   []string
	isAlive   = make(map[string]bool)
	displayed = make(map[string]bool)
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
	}

	// Start report routine
	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			for i, metrics := range monitor.ExportAndClear() {

				host := targets[[]byte(i)[0]]
				alive := (metrics.PacketsSent - metrics.PacketsLost) > 0

				if (!displayed[host]) || (isAlive[host] != alive) {
					stamp := time.Now().Format("2006-02-01 15:04:05")

					switch {

					case alive && metrics.PacketsLost == 0:
						fmt.Fprintf(color.Output, "%s %s", stamp,
							color.GreenString(fmt.Sprintf("%s is up [%d/%d]\n",
								host, metrics.PacketsSent-metrics.PacketsLost, metrics.PacketsSent)))

					case alive && metrics.PacketsLost != 0:
						fmt.Fprintf(color.Output, "%s %s", stamp,
							color.YellowString(fmt.Sprintf("%s is up but loosing packets [%d/%d]\n",
								host, metrics.PacketsSent-metrics.PacketsLost, metrics.PacketsSent)))

					case !alive:
						fmt.Fprintf(color.Output, "%s %s", stamp,
							color.RedString(fmt.Sprintf("%s is down [%d/%d]\n",
								host, metrics.PacketsSent-metrics.PacketsLost, metrics.PacketsSent)))

					}

					isAlive[host], displayed[host] = alive, true
				}
			}
		}
	}()

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("received", <-ch)
}

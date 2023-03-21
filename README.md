# go-icmp-status

Very simple tool that keep sending icmp packet to a list of ipv4 or ipv6 hosts, and display flip/flap icmp status changes, and packet loss.

Original code from [github.com/digineo/go-ping/cmd/ping-monitor](https://github.com/digineo/go-ping/tree/master/cmd/ping-monitor)

* Need root rights on linux for sending icmp packets ( or sudo, or chown root `binary` , chmod u+s `binary` after build )

* Go build

```shell
> git clone https://github.com/thc2cat/go-icmp-status 
> cd go-icmp-status 
> go mod tidy 
> go build
```

* Options

```shell
v0.11 $ ./go-icmp-status.exe 
Usage: C:\dev\src\projects\go-icmp-status\go-icmp-status.exe [options] [hosts...]
  -dateFormat string
        log date format (default "2006-01-02 15:04:05")
  -logToSyslog
        log events to syslog
  -noLossReport
        do not report summary
  -pingInterval duration
        interval for ICMP echo requests (default 1s)
  -pingTimeout duration
        timeout for ICMP echo request (default 3s)
  -r string
        read targets from file
  -reportInterval duration
        interval for reports (default 5s)
  -showIp
        show monitored targets name resolution
  -size uint
        size of additional payload data (default 56)
  -stopAfter duration
        test duration (example 10m) (default 8760h0m0s)
  -t    be tolerant, allow 1 packet loss per check

```

* monitoring a list of hosts from a file :

```shell
go-icmp-status -pingInterval 5s -r hosts.txt
```

* Colored output exemple (v7) with multiple names in cli

![ipv6 loss](ipv6-loss.png)

* Green for a host receiving all packets during interval
* Red for a host loosing all packets during interval
* Yellow for a host up but loosing packets during interval, [Received/Sent/Percent] indicate x received packet for y sent packets during interval. Percent indicate percentage of received packets since start.

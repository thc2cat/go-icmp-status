# go-icmp-status

Very simple tool that keep sending icmp packet to a list of ipv4 or ipv6 hosts, and display flip/flap icmp status changes, and packet loss.

Original code from [github.com/digineo/go-ping/cmd/ping-monitor](https://github.com/digineo/go-ping/tree/master/cmd/ping-monitor)

* Need root rights on linux for sending icmp packets ( or sudo, or chown root `binary` , chmod u+s `binary` after build )

* Go build

```shell
> git clone https://github.com/thc2cat/go-icmp-status 
> cd go-icmp-status 
> go mod tidy 
> go build`
```

* fast way of monitoring a list of hosts :

```shell
> cat hosts.txt | xargs go-icmp-status -pingInterval 30s
```

* Colored output ( red/green/yellow ) with timestamp
of continuous monitoring (after `mtr` check and text paste):

![ipv6 loss](ipv6-loss.png)

* Green for a host receiving all packets
* Red for a host loosing all packets
* Yellow for a host up but loosing packets
* [Received/Sent] indicate x received packet for y sent packets during interval.

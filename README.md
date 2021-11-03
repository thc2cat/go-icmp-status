# go-icmp-status

Very simple tool that keep sending icmp packet to a lot of ipv4 or ipv6 hosts, and display flip/flap icmp status changes, and packet loss.

* Need root rights on linux for sendind icmp packets ( or sudo, or chown root `binary` , chmod u+s `binary` )

* Dependencies [github.com/digineo/go-ping](https://github.com/digineo/go-ping) + monitor

* Go build :  `go mod init ; go mod tidy ; go build` 

* Colored output ( red/green/yellow ) with timestamp

```diff
- 2021-11-10 15:55:55 host.domain not responding
+ 2021-11-10 15:55:56 anotherhost.domain is alive
```

* simple way of monitoring a list of hosts :

```shell
cat list | xargs go-icmp-status -pingInterval 30s
```
# Go icmp-status

Very simple tool that keep sending icmp packet to a lot of hosts, and display flip/flap icmp status changes.

* Need root on linux for icmp packets.

* Build with github.com/digineo/go-ping + monitor

* Download dependencies with ``` go mod tidy ```

* Build from go with  ``` go build ```

* simple colored output ( red/green/yellow ) with timestamp

```diff
- 2021-11-10 15:55 host.domain not responding
+ 2021-11-10 15:55 anotherhost.domain is alive
```

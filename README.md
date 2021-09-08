# Go icmp-status

Very simple tool that keep sending icmp packet to a lot of hosts, and display flip/flap icmp status changes.

* Need root on linux for icmp packets.

* Build with github.com/digineo/go-ping + monitor

* Download dependencies with ``` go mod tidy ```

* Build from go with  ``` go build ```

* simple colored output ( red /green )

```text
host.domain not responding
anotherhost.domain up again
```

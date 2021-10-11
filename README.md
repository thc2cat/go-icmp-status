# Go icmp-status

Very simple tool that keep sending icmp packet to a lot of hosts, and display flip/flap icmp status changes.

* Need root on linux for icmp packets.

* Build with github.com/digineo/go-ping + monitor

* Download dependencies with ``` go mod tidy ```

* Build from go with  ``` go build ```

* simple colored output ( red /green ) with timespamp

---

2021-11-10 15:55  <span style="color:red">host.domain not responding</span>  

2021-11-10 15:55  <span style="color:green">anotherhost.domain is alive</span>

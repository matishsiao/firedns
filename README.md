firedns
============
firedns is a simple DNS server which downloads zones over SSDB in JSON format,
parses them, stores them in memory and serves to clients.

* firedns listens on both TCP and UDP port 53.
* It does not support daemonization.
* There is no config file.
* There is no reading from local zone files.
* Malformed zones or records are ignored (with a warning in log).
* Logs to stdout.
* Uses many threads to handle connections (by Go goroutines).
* Exits gracefully on SIGINT or SIGTERM.

Benchmark
============
using Intel(R) Xeon(R) CPU E5-2660 v2 @ 2.20GHz and run 19 process for service udp query
Query Per Second:535000

How to build
============
```bash
go get github.com/matishsiao/firedns
cd GOPATH/github.com/matishsiao/firedns
go build
```

How to run
============
* Parameter ```-c``` defines address of SSDB server. Defaults to 127.0.0.1
* Parameter ```-p``` defines port of SSDB server. Defaults to 8888
* Parameter ```-a``` defines port of SSDB auth password. Defaults to ""
* Parameter ```-d``` defines debug mode. Defaults to false
* Parameter ```-l``` defines the local IP (interface) to listen on. Defaults to all.
* Parameter ```-r``` enable recursive querying of specified servers for answers godnsagent can't provide itself.
* Parameter ```-k``` sets the API key (passed as GET or form value "key" to http notification handlers)

```
./firedns -c 127.0.0.1 -p 8888
```

How it works
============
* It binds to ports 53 on TCP and UDP and serves queries.
* The longest matching zone is chosen.
* All NS records on the zone are returned with an answer as "Authoritative" section.
* If possible, resolutions for NS records are added as "Extra" section.
* If zone file not exist in memory. It will try connect SSDB server to get zone json. If fail it will refused.

Schema of zones file
============
* Class field is optional, defaults to IN
* Fields are case-insensitive
* TTL is optional, defaults to 3600
* Data must hold all information specific to record type (see MX, SOA, SRV, etc.)
* The zone should have SOA record, although godnsagent will not complain
* The zone should have NS records, although godnsagent will not complain
* Zone name (key) should be FQDN or godnsagent will make it FQDN
* Use FQDN whenever possible

```json
{
  "spigu.com.": [
    {"name": "spigu.com.", "type": "A", "tTl": 500, "data": "123.123.123.123"},
    {"name": "b.spigu.com.", "type": "A", "Class": "CH", "Ttl": 300, "data": "123.123.123.124"},
    {"name": "spigu.com", "type": "MX", "Class": "IN", "Ttl": 305, "data": "5 email.spigu.net."},
    {"name": "spigu.com", "type": "NS", "data": "marley.spigu.com."},
    {"name": "spigu.com", "type": "NS", "Class": "IN", "Ttl": 300, "data": "abc.spigu.com."},
    {"name": "spigu.com", "type": "SOA", "TTL": 300, "data": "abc.spigu.com. hostmaster.spigu.com. 1399838297 21600 3600 1814400 300"}
  ],
  "spigu.net.": [
    {"name": "spigu.net.", "type": "A", "tTl": 500, "data": "123.123.123.123"},
    {"name": "b.spigu.net.", "type": "A", "Class": "CH", "Ttl": 300, "data": "123.123.123.125"},
    {"name": "spigu.net", "type": "MX", "Class": "IN", "Ttl": 305, "data": "5 email.spigu.net."},
    {"name": "spigu.net", "type": "NS", "data": "marley.spigu.net."},
    {"name": "spigu.net", "type": "NS", "Class": "IN", "Ttl": 300, "data": "abc.spigu.net."},
    {"name": "abc.spigu.net", "type": "A", "data": "123.123.123.100"},
    {"name": "marley.spigu.net", "type": "A", "data": "123.123.123.101"},
    {"name": "spigu.net", "type": "SOA", "TTL": 300, "data": "marley.spigu.net. hostmaster.spigu.net. 1399838297 21600 3600 1814400 300"}
  ]
}
```

Acknowledgments
============
This software was created thanks to two amazing projects:
  * https://github.com/miekg/dns: DNS library by miekg provided awesome foundations of DNS in Go.
  * https://github.com/DevelopersPL/godnsagent: DNS agent by DevelopersPL provided a great example and reference point for using it.
  * https://github.com/kenshinx/godns: Kenshinx's goDNS provided a great example and reference point for using it.

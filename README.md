# Etherpad prometheus exporter

### Build

```bash
make build
```

### Usage

```bash
etherpad_exporter -h
usage: etherpad_exporter [<flags>]

Flags:
  -h, --help                 Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9301"  
                             Address to listen on for web interface and telemetry.
      --web.telemetry-path="/metrics"  
                             Path under which to expose metrics.
      --etherpad.scrape-uri="http://localhost:9001/stats?"  
                             URI on which to scrape Etherpad.
      --etherpad.ssl-verify  Flag that enables SSL certificate verification for the scrape URI
      --etherpad.timeout=5s  Timeout for trying to get stats from Etherpad.
      --log.level=info       Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt    Output format of log messages. One of: [logfmt, json]

```

### Exposed metrics

```bash
users_total
connects_total
disconnects_total
pending_edits_total
edits_total
failed_change_sets_total
http_requests_total
http_500_total
memory_usage
```

### Docker image

https://hub.docker.com/r/l3opold/etherpad_exporter
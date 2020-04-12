# Etherpad prometheus exporter

[![Docker Repository on Quay](https://quay.io/repository/l3opold/etherpad_exporter/status)][quay]
[![Docker pull](https://img.shields.io/docker/pulls/l3opold/etherpad_exporter)][dockerpull]
[![Go Report Card](https://goreportcard.com/badge/github.com/L3o-pold/etherpad_exporter)][goreportcard]
[![Code Climate](https://codeclimate.com/github/L3o-pold/etherpad_exporter/badges/gpa.svg)][codeclimate]

[quay]: https://quay.io/repository/l3opold/etherpad_exporter
[dockerpull]: https://hub.docker.com/r/l3opold/etherpad_exporter/
[goreportcard]: https://goreportcard.com/report/github.com/L3o-pold/etherpad_exporter
[codeclimate]: https://codeclimate.com/github/L3o-pold/etherpad_exporter


## Build

```bash
make build
```

## Usage

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

## Exposed metrics

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

## Docker image

https://hub.docker.com/repository/docker/l3opold/etherpad_exporter

## Grafana dashboard

https://github.com/L3o-pold/etherpad_exporter/tree/master/grafana/dashboard.json

## License

Apache License 2.0, see [LICENSE](https://github.com/prometheus/haproxy_exporter/blob/master/LICENSE).
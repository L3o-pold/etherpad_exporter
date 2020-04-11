package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	namespace = "etherpad" // For Prometheus metrics.
)

func newServerMetric(metricName string, docString string, constLabels prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, "server", metricName), docString, nil, constLabels)
}

type metrics map[string]*prometheus.Desc

func (m metrics) String() string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ",")
}

var (
	serverMetrics = metrics{
		"totalUsers":       newServerMetric("users_total", "Total users.", nil),
		"connects":         newServerMetric("connects_total", "Connects.", nil),
		"disconnects":      newServerMetric("disconnects_total", "Disconnects.", nil),
		"pendingEdits":     newServerMetric("pending_edits_total", "Pending edits.", nil),
		"edits":            newServerMetric("edits_total", "Edits.", nil),
		"failedChangesets": newServerMetric("failed_change_sets_total", "Failed change sets.", nil),
		"httpRequests":     newServerMetric("http_requests_total", "Http requests.", nil),
		"http500":          newServerMetric("http_500_total", "Http 500.", nil),
		"memoryUsage":      newServerMetric("memory_usage", "Memory usage.", nil),
	}

	etherpadUp = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "up"), "Was the last scrape of etherpad successful.", nil, nil)
)

type EtherPadMetrics struct {
	MemoryUsage, TotalUsers, PendingEdits int
	HttpRequests, Edits                   struct {
		Meter struct {
			Count int
		}
	}
	Connects, Disconnects, Http500, FailedChangesets struct {
		Count int
	}
}

// Exporter collects Etherpad stats from the given URI and exports them using
// the prometheus metrics package.
type Exporter struct {
	URI   string
	mutex sync.RWMutex
	fetch func() (io.ReadCloser, error)

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter
	serverMetrics                   map[string]*prometheus.Desc
	logger                          log.Logger
}

// NewExporter returns an initialized Exporter.
func NewExporter(uri string, sslVerify bool, timeout time.Duration, logger log.Logger) (*Exporter, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	var fetch func() (io.ReadCloser, error)
	switch u.Scheme {
	case "http", "https":
		fetch = fetchHTTP(uri, sslVerify, timeout)
	default:
		return nil, fmt.Errorf("unsupported scheme: %q", u.Scheme)
	}

	return &Exporter{
		URI:   uri,
		fetch: fetch,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of etherpad successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_total_scrapes",
			Help:      "Current total Etherpad scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_json_parse_failures",
			Help:      "Number of errors while parsing json.",
		}),
		serverMetrics: serverMetrics,
		logger:        logger,
	}, nil
}

// Describe describes all the metrics ever exported by the Etherpad exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.serverMetrics {
		ch <- m
	}
	for _, m := range e.serverMetrics {
		ch <- m
	}
	ch <- etherpadUp
	ch <- e.totalScrapes.Desc()
	ch <- e.jsonParseFailures.Desc()
}

// Collect fetches the stats from configured Etherpad location and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	up := e.scrape(ch)

	ch <- prometheus.MustNewConstMetric(etherpadUp, prometheus.GaugeValue, up)
	ch <- e.totalScrapes
	ch <- e.jsonParseFailures
}

func fetchHTTP(uri string, sslVerify bool, timeout time.Duration) func() (io.ReadCloser, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !sslVerify}}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	return func() (io.ReadCloser, error) {
		resp, err := client.Get(uri)
		if err != nil {
			return nil, err
		}
		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
		}
		return resp.Body, nil
	}
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) (up float64) {
	e.totalScrapes.Inc()

	body, err := e.fetch()
	if err != nil {
		level.Error(e.logger).Log("msg", "Can't scrape Etherpad", "err", err)
		return 0
	}
	defer body.Close()

	var result EtherPadMetrics
	err = json.NewDecoder(body).Decode(&result)
	if err == nil {
		e.exportJsonFields(e.serverMetrics, result, ch)
		return 1
	}

	level.Error(e.logger).Log("msg", "Can't read JSON", "err", err)
	e.jsonParseFailures.Inc()
	return 0
}

func (e *Exporter) exportJsonFields(metrics map[string]*prometheus.Desc, result EtherPadMetrics, ch chan<- prometheus.Metric) {

	for fieldIdx, metric := range metrics {
		switch fieldIdx {
		case "totalUsers":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.CounterValue, float64(result.TotalUsers))
		case "connects":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.CounterValue, float64(result.Connects.Count))
		case "disconnects":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.CounterValue, float64(result.Disconnects.Count))
		case "pendingEdits":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.CounterValue, float64(result.PendingEdits))
		case "edits":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.CounterValue, float64(result.Edits.Meter.Count))
		case "failedChangesets":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.CounterValue, float64(result.FailedChangesets.Count))
		case "httpRequests":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.CounterValue, float64(result.HttpRequests.Meter.Count))
		case "http500":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.CounterValue, float64(result.Http500.Count))
		case "memoryUsage":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, float64(result.MemoryUsage))
		}
	}
}

func main() {
	var (
		listenAddress     = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9301").String()
		metricsPath       = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		etherpadScrapeURI = kingpin.Flag("etherpad.scrape-uri", "URI on which to scrape Etherpad.").Default("http://localhost:9001/stats?").String()
		etherpadSSLVerify = kingpin.Flag("etherpad.ssl-verify", "Flag that enables SSL certificate verification for the scrape URI").Default("true").Bool()
		etherpadTimeout   = kingpin.Flag("etherpad.timeout", "Timeout for trying to get stats from Etherpad.").Default("5s").Duration()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting etherpad_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())
	level.Info(logger).Log("msg", "Parse", "url", etherpadScrapeURI)
	exporter, err := NewExporter(*etherpadScrapeURI, *etherpadSSLVerify, *etherpadTimeout, logger)
	if err != nil {
		level.Error(logger).Log("msg", "Error creating an exporter", "err", err)
		os.Exit(1)
	}
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector("etherpad_exporter"))

	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Etherpad Exporter</title></head>
             <body>
             <h1>Etherpad Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}

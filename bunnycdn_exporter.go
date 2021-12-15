// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace = "bunnycdn" // For Prometheus metrics.

	metricOriginResponseTime = "OriginResponseTime"
	metricBandwidthUsed      = "bandwidthUsed"
	metricBandwidthCached    = "bandwidthCached"
	metricRequestsServer     = "requestsServed"
	metricPullRequestsPulled = "pullRequestsPulled"
	metricErr3xx             = "error3xx"
	metricErr4xx             = "error4xx"
	metricErr5xx             = "error5xx"
	metricGeoTrafficDist     = "geoTrafficDistribution"

	metricBalance     = "balance"
	metricStorageUsed = "storageUsed"
)

func newMetric(metricName string, docString string, variableLabels []string, constLabels prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, "", metricName), docString, variableLabels, constLabels)
}

type bunnyPullZone struct {
	ID   int64  `json:"Id"`
	Name string `json:"Name"`
}

type bunnyLocation struct {
	Region   string
	Location string
	Requests float64
}

type bunnyStatistics struct {
	OriginResponseTime     map[string]float64 `json:"OriginResponseTimeChart"`
	BandwidthUsed          map[string]float64 `json:"BandwidthUsedChart"`
	BandwidthCached        map[string]float64 `json:"BandwidthCachedChart"`
	CacheHitRate           map[string]float64 `json:"CacheHitRateChart"`
	RequestsServed         map[string]float64 `json:"RequestsServedChart"`
	PullRequestsPulled     map[string]float64 `json:"PullRequestsPulledChart"`
	UserBalanceHistory     map[string]float64 `json:"UserBalanceHistoryChart"`
	UserStorageUsed        map[string]float64 `json:"UserStorageUsedChart"`
	GeoTrafficDistribution map[string]float64 `json:"GeoTrafficDistribution"`
	Error3Xx               map[string]float64 `json:"Error3xxChart"`
	Error4Xx               map[string]float64 `json:"Error4xxChart"`
	Error5Xx               map[string]float64 `json:"Error5xxChart"`
}

func (s bunnyStatistics) trafficLocations() []bunnyLocation {
	locations := make([]bunnyLocation, 0, len(s.GeoTrafficDistribution))
	for loc, req := range s.GeoTrafficDistribution {
		parts := strings.Split(loc, ":")
		locations = append(
			locations,
			bunnyLocation{
				Region:   parts[0],
				Location: strings.TrimSpace(parts[1]),
				Requests: req,
			})
	}
	return locations
}

func extractFromMap(data map[string]float64) float64 {
	for _, v := range data {
		return v
	}
	return -1
}

type metricsCollection map[string]*prometheus.Desc

func (m metricsCollection) String() string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

var (
	accountMetrics = metricsCollection{
		metricBalance:     newMetric("account_balance", "Current account balance", nil, nil),
		metricStorageUsed: newMetric("storage_used_bytes", "Storage usage in bytes", nil, nil),
	}
	pullZoneMetrics = metricsCollection{
		metricOriginResponseTime: newMetric("origin_response_time_avg", "The average origin response time.", []string{"pull_zone"}, nil),
		metricBandwidthUsed:      newMetric("bandwidth_used_bytes_total", "Total bandwidth used in bytes serving traffic.", []string{"pull_zone"}, nil),
		metricBandwidthCached:    newMetric("bandwidth_cached_bytes_total", "Total bandwidth used in bytes serving cached data.", []string{"pull_zone"}, nil),
		metricRequestsServer:     newMetric("requests_served_total", "Number of requests served.", []string{"pull_zone"}, nil),
		metricPullRequestsPulled: newMetric("pull_requests_pulled", "Number of pull requests from origin.", []string{"pull_zone"}, nil),
		metricErr3xx:             newMetric("request_error_count", "Request error by code.", []string{"pull_zone"}, prometheus.Labels{"code": "3xx"}),
		metricErr4xx:             newMetric("request_error_count", "Request error by code.", []string{"pull_zone"}, prometheus.Labels{"code": "4xx"}),
		metricErr5xx:             newMetric("request_error_count", "Request error by code.", []string{"pull_zone"}, prometheus.Labels{"code": "5xx"}),
		metricGeoTrafficDist:     newMetric("requests_served", "Request by location.", []string{"pull_zone", "region", "location"}, nil),
	}
	bunnyUp = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "up"), "Was the last scrape of bunnyCDN successful.", nil, nil)
)

// Exporter collects BunnyCDN stats from the given URI and exports them using
// the prometheus metrics package.
type Exporter struct {
	URI   string
	mutex sync.RWMutex
	fetch func(path string) (io.ReadCloser, error)

	up                                       prometheus.Gauge
	totalScrapes, totalErrors, totalAPICalls prometheus.Counter
	accountMetrics                           metricsCollection
	pullZoneMetrics                          metricsCollection
}

// NewExporter returns an initialized Exporter.
func NewExporter(uri string, bunnyAPIKey string, sslVerify bool, accountMetrics metricsCollection, pullZoneMetrics metricsCollection, timeout time.Duration) (*Exporter, error) {
	var fetch func(path string) (io.ReadCloser, error)
	fetch = fetchHTTP(uri, bunnyAPIKey, sslVerify, timeout)

	return &Exporter{
		URI:   uri,
		fetch: fetch,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of bunny successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_total_scrapes",
			Help:      "Current total BunnyCDN scrapes.",
		}),
		totalErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_total_errors",
			Help:      "Number of errors while making API calls.",
		}),
		totalAPICalls: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_api_calls_total",
			Help:      "Number of calls made to BunnyCDN API",
		}),
		accountMetrics:  accountMetrics,
		pullZoneMetrics: pullZoneMetrics,
	}, nil
}

// Describe describes all the metrics ever exported by the BunnyCDN exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.accountMetrics {
		ch <- m
	}
	for _, m := range e.pullZoneMetrics {
		ch <- m
	}
	ch <- e.up.Desc()
	ch <- e.totalScrapes.Desc()
	ch <- e.totalErrors.Desc()
	ch <- e.totalAPICalls.Desc()
}

// Collect fetches the stats from BunnyCDN API and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	up := e.scrape(ch)

	e.up.Set(up)
	// ch <- prometheus.MustNewConstMetric(bunnyUp, prometheus.GaugeValue, up)
	ch <- e.up
	ch <- e.totalScrapes
	ch <- e.totalErrors
	ch <- e.totalAPICalls
}

func getStatisticsForPullZone(fetch func(path string) (io.ReadCloser, error), pz bunnyPullZone) (*bunnyStatistics, error) {
	return rawGetStatistics(fetch, map[string]string{"pullZone": fmt.Sprintf("%d", pz.ID)})
}

func getStatistics(fetch func(path string) (io.ReadCloser, error)) (*bunnyStatistics, error) {
	return rawGetStatistics(fetch, nil)
}

func rawGetStatistics(fetch func(path string) (io.ReadCloser, error), extraParams map[string]string) (*bunnyStatistics, error) {
	today := time.Now()

	p := url.Values{}

	p.Add("dateFrom", today.Format("2006-01-02"))
	p.Add("dateTo", today.Format("2006-01-02"))
	p.Add("loadErrors", "true")

	for k, v := range extraParams {
		p.Add(k, v)
	}

	body, err := fetch(
		fmt.Sprintf(
			"/statistics?%s",
			p.Encode(),
		),
	)
	if err != nil {
		return nil, err
	}

	bodyContent, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Printf("Reading failed (read: %d): %s\n", len(bodyContent), err)
		return nil, err
	}

	var stats bunnyStatistics
	err = json.Unmarshal(bodyContent, &stats)

	return &stats, err
}

func listPullZones(fetch func(path string) (io.ReadCloser, error)) ([]bunnyPullZone, error) {
	// body, err := fetchHTTP("https://bunnycdn.com/api/pullzone", true, time.Seconds*5)
	body, err := fetch("/pullzone")
	if err != nil {
		return nil, err
	}

	bodyContent, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Printf("Reading failed (read: %d): %s\n", len(bodyContent), err)
		return nil, err
	}

	var pullZones []bunnyPullZone
	err = json.Unmarshal(bodyContent, &pullZones)

	return pullZones, err
}

func fetchHTTP(uri string, bunnyAPIKey string, sslVerify bool, timeout time.Duration) func(path string) (io.ReadCloser, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !sslVerify}}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	return func(path string) (io.ReadCloser, error) {

		req, err := http.NewRequest("GET", uri+path, nil)
		req.Header.Set("AccessKey", bunnyAPIKey)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP status %d (%s)", resp.StatusCode, resp.Request.URL)
		}
		return resp.Body, nil
	}
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) (up float64) {
	e.totalScrapes.Inc()

	pullZones, err := listPullZones(e.fetch)
	e.totalAPICalls.Inc()

	// body, err := e.fetch("/metrics")
	if err != nil {
		log.Errorf("Unable to list pull zones: %v", err)
		e.totalErrors.Inc()
		return 0
	}

	var aStatsObj *bunnyStatistics

	for _, pullZone := range pullZones {
		stats, err := getStatisticsForPullZone(e.fetch, pullZone)
		e.totalAPICalls.Inc()
		if err != nil {
			log.Errorf("Unable to collect stats: %v", err)
			e.totalErrors.Inc()
		}
		aStatsObj = stats
		for name, metric := range e.pullZoneMetrics {
			switch name {
			case metricOriginResponseTime:
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, extractFromMap(stats.OriginResponseTime), pullZone.Name)
			case metricBandwidthUsed:
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, extractFromMap(stats.BandwidthUsed), pullZone.Name)
			case metricBandwidthCached:
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, extractFromMap(stats.BandwidthUsed), pullZone.Name)
			case metricRequestsServer:
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, extractFromMap(stats.RequestsServed), pullZone.Name)
			case metricPullRequestsPulled:
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, extractFromMap(stats.PullRequestsPulled), pullZone.Name)
			case metricErr3xx:
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, extractFromMap(stats.Error3Xx), pullZone.Name)
			case metricErr4xx:
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, extractFromMap(stats.Error4Xx), pullZone.Name)
			case metricErr5xx:
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, extractFromMap(stats.Error5Xx), pullZone.Name)
			case metricGeoTrafficDist:
				for _, loc := range stats.trafficLocations() {
					ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, loc.Requests, pullZone.Name, loc.Region, loc.Location)
				}
			}
		}
	}

	if aStatsObj == nil {
		aStatsObj, err = getStatistics(e.fetch)
		e.totalAPICalls.Inc()

		if err != nil {
			log.Errorf("Unable to collect global stats (since no pullzone was found): %v", err)
			e.totalErrors.Inc()
		}
	}

	if aStatsObj != nil {
		for name, metric := range e.accountMetrics {
			switch name {
			case metricBalance:
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, extractFromMap(aStatsObj.UserBalanceHistory))
			case metricStorageUsed:
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, extractFromMap(aStatsObj.UserStorageUsed))
			}
		}
	}
	return 1
}

func main() {
	var (
		listenAddress  = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9584").String()
		metricsPath    = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		bunnyAPIURI    = kingpin.Flag("bunnycdn.api-uri", "API URI on which to get stats from.").Default("https://bunnycdn.com/api").String()
		bunnyAPIKey    = kingpin.Flag("bunnycdn.api-key", "API key to connect to bunny.").Default(os.Getenv("BUNNYCDN_API_KEY")).String()
		bunnySSLVerify = kingpin.Flag("bunnycdn.ssl-verify", "Flag that enables SSL certificate verification for the API URI").Default("true").Bool()
		bunnyTimeout   = kingpin.Flag("bunnycdn.timeout", "Timeout for trying to get stats from BunnyCDN.").Default("10s").Duration()
	)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("bunnycdn_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting bunnycdn_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	exporter, err := NewExporter(*bunnyAPIURI, *bunnyAPIKey, *bunnySSLVerify, accountMetrics, pullZoneMetrics, *bunnyTimeout)
	if err != nil {
		log.Fatal(err)
	}
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector("bunnycdn_exporter"))

	log.Infoln("Listening on", *listenAddress)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>BunnyCDN Exporter</title></head>
             <body>
             <h1>BunnyCDN Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

package main

import (
	"crypto/tls"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	prom_strutil "github.com/prometheus/prometheus/util/strutil"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ExporterOptions struct {
	Password            string
	Namespace           string
	ConfigCommandName   string
	CheckSingleKeys     string
	CheckKeys           string
	LuaScript           []byte
	ClientCertificates  []tls.Certificate
	InclSystemMetrics   bool
	SkipTLSVerification bool
	SetClientName       bool // can set redis client name
	IsTile38            bool
	ExportClientList    bool
	ConnectionTimeouts  time.Duration
	MetricsPath         string
	RedisMetricsOnly    bool
	Registry            *prometheus.Registry
}

// Exporter implements the prometheus.Exporter interface, and exports Redis metrics.
/**
@see
*/
type Exporter struct {
	sync.Mutex
	redisAddr string
	namespace string

	totalScrapes              prometheus.Counter
	scrapeDuration            prometheus.Summary
	targetScrapeRequestErrors prometheus.Counter

	metricDescriptions map[string]*prometheus.Desc

	options ExporterOptions

	metricMapCounters map[string]string
	metricMapGauges   map[string]string

	mux *http.ServeMux
}

func NewRedisExporter(redisAddr string, opts ExporterOptions) (*Exporter, error) {
	exporter := &Exporter{
		redisAddr: redisAddr,
		namespace: opts.Namespace,
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: opts.Namespace,
			Name:      "exporter_scrapes_total",
			Help:      "Current total redis scrapes.",
		}),
		scrapeDuration: prometheus.NewSummary(prometheus.SummaryOpts{
			Namespace:   opts.Namespace,
			Subsystem:   "",
			Name:        "exporter_scrape_duration_seconds",
			Help:        "Duration of scrape by the exporter",
			ConstLabels: nil,
			Objectives:  nil,
			MaxAge:      0,
			AgeBuckets:  0,
			BufCap:      0,
		}),
		targetScrapeRequestErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: opts.Namespace,
			Name:      "target_scrape_request_errors_total",
			Help:      "Errors in requests to the exporter",
		}),
		metricDescriptions: nil,
		options:            ExporterOptions{},
		metricMapCounters: map[string]string{
			"total_connections_received": "connections_received_total",
			"total_commands_processed":   "commands_processed_total",

			"rejected_connections":   "rejected_connections_total",
			"total_net_input_bytes":  "net_input_bytes_total",
			"total_net_output_bytes": "net_output_bytes_total",

			"expired_keys":    "expired_keys_total",
			"evicted_keys":    "evicted_keys_total",
			"keyspace_hits":   "keyspace_hits_total",
			"keyspace_misses": "keyspace_misses_total",

			"used_cpu_sys":           "cpu_sys_seconds_total",
			"used_cpu_user":          "cpu_user_seconds_total",
			"used_cpu_sys_children":  "cpu_sys_children_seconds_total",
			"used_cpu_user_children": "cpu_user_children_seconds_total",
		},
		metricMapGauges: map[string]string{
			// # Server
			"uptime_in_seconds": "uptime_in_seconds",
			"process_id":        "process_id",

			// # Clients
			"connected_clients": "connected_clients",
			"blocked_clients":   "blocked_clients",

			"client_recent_max_output_buffer": "client_recent_max_output_buffer_bytes",
			"client_recent_max_input_buffer":  "client_recent_max_input_buffer_bytes",

			// # Memory
			"allocator_active":    "allocator_active_bytes",
			"allocator_allocated": "allocator_allocated_bytes",
			"allocator_resident":  "allocator_resident_bytes",
			"used_memory":         "memory_used_bytes",
			"used_memory_rss":     "memory_used_rss_bytes",
			"used_memory_peak":    "memory_used_peak_bytes",
			"used_memory_lua":     "memory_used_lua_bytes",
			"maxmemory":           "memory_max_bytes",

			// # Persistence
			"rdb_changes_since_last_save":  "rdb_changes_since_last_save",
			"rdb_bgsave_in_progress":       "rdb_bgsave_in_progress",
			"rdb_last_save_time":           "rdb_last_save_timestamp_seconds",
			"rdb_last_bgsave_status":       "rdb_last_bgsave_status",
			"rdb_last_bgsave_time_sec":     "rdb_last_bgsave_duration_sec",
			"rdb_current_bgsave_time_sec":  "rdb_current_bgsave_duration_sec",
			"rdb_last_cow_size":            "rdb_last_cow_size_bytes",
			"aof_enabled":                  "aof_enabled",
			"aof_rewrite_in_progress":      "aof_rewrite_in_progress",
			"aof_rewrite_scheduled":        "aof_rewrite_scheduled",
			"aof_last_rewrite_time_sec":    "aof_last_rewrite_duration_sec",
			"aof_current_rewrite_time_sec": "aof_current_rewrite_duration_sec",
			"aof_last_cow_size":            "aof_last_cow_size_bytes",
			"aof_current_size":             "aof_current_size_bytes",
			"aof_base_size":                "aof_base_size_bytes",
			"aof_pending_rewrite":          "aof_pending_rewrite",
			"aof_buffer_length":            "aof_buffer_length",
			"aof_rewrite_buffer_length":    "aof_rewrite_buffer_length",
			"aof_pending_bio_fsync":        "aof_pending_bio_fsync",
			"aof_delayed_fsync":            "aof_delayed_fsync",
			"aof_last_bgrewrite_status":    "aof_last_bgrewrite_status",
			"aof_last_write_status":        "aof_last_write_status",

			// # Stats
			"pubsub_channels":  "pubsub_channels",
			"pubsub_patterns":  "pubsub_patterns",
			"latest_fork_usec": "latest_fork_usec",

			// # Replication
			"loading":                    "loading_dump_file",
			"connected_slaves":           "connected_slaves",
			"repl_backlog_size":          "replication_backlog_bytes",
			"master_last_io_seconds_ago": "master_last_io_seconds",
			"master_repl_offset":         "master_repl_offset",

			// # Cluster
			"cluster_stats_messages_sent":     "cluster_messages_sent_total",
			"cluster_stats_messages_received": "cluster_messages_received_total",

			// # Tile38
			// based on https://tile38.com/commands/server/
			"tile38_aof_size":        "tile38_aof_size_bytes",
			"tile38_avg_item_size":   "tile38_avg_item_size_bytes",
			"tile38_cpus":            "tile38_cpus_total",
			"tile38_heap_released":   "tile38_heap_released_bytes",
			"tile38_heap_size":       "tile38_heap_size_bytes",
			"tile38_http_transport":  "tile38_http_transport",
			"tile38_in_memory_size":  "tile38_in_memory_size_bytes",
			"tile38_max_heap_size":   "tile38_max_heap_size_bytes",
			"tile38_mem_alloc":       "tile38_mem_alloc_bytes",
			"tile38_num_collections": "tile38_num_collections_total",
			"tile38_num_hooks":       "tile38_num_hooks_total",
			"tile38_num_objects":     "tile38_num_objects_total",
			"tile38_num_points":      "tile38_num_points_total",
			"tile38_pointer_size":    "tile38_pointer_size_bytes",
			"tile38_read_only":       "tile38_read_only",
			"tile38_threads":         "tile38_threads_total",
		},
		mux: nil,
	}

	if exporter.options.ConfigCommandName == "" {
		exporter.options.ConfigCommandName = "CONFIG"
	}
	if keys, err := parseKeyArg(opts.CheckKeys); err != nil {
		return nil, fmt.Errorf("Couldn't parse check-keys: %#v", err)
	} else {
		log.Debugf("keys: %#v", keys)
	}
	if singleKeys, err := parseKeyArg(opts.CheckSingleKeys); err != nil {
		return nil, fmt.Errorf("Couldn't parse check-single-keys: %#v", err)
	} else {
		log.Debugf("singleKeys: %#v", singleKeys)
	}

	if opts.InclSystemMetrics {
		exporter.metricMapGauges["total_system_memory"] = "total_system_memory_bytes"
	}

	exporter.metricDescriptions = map[string]*prometheus.Desc{}
	for key, desc := range map[string]struct {
		txt    string
		labels []string
	}{
		"commands_duration_seconds_total":      {txt: `Total amount of time in seconds spent per command`, labels: []string{"cmd"}},
		"commands_total":                       {txt: `Total number of calls per command`, labels: []string{"cmd"}},
		"connected_slave_lag_seconds":          {txt: "Lag of connected slave", labels: []string{"slave_ip", "slave_port", "slave_state"}},
		"connected_slave_offset_bytes":         {txt: "Offset of connected slave", labels: []string{"slave_ip", "slave_port", "slave_state"}},
		"db_avg_ttl_seconds":                   {txt: "Avg TTL in seconds", labels: []string{"db"}},
		"db_keys":                              {txt: "Total number of keys by DB", labels: []string{"db"}},
		"db_keys_expiring":                     {txt: "Total number of expiring keys by DB", labels: []string{"db"}},
		"exporter_last_scrape_error":           {txt: "The last scrape error status.", labels: []string{"err"}},
		"instance_info":                        {txt: "Information about the Redis instance", labels: []string{"role", "redis_version", "redis_build_id", "redis_mode", "os"}},
		"key_size":                             {txt: `The length or size of "key"`, labels: []string{"db", "key"}},
		"key_value":                            {txt: `The value of "key"`, labels: []string{"db", "key"}},
		"last_slow_execution_duration_seconds": {txt: `The amount of time needed for last slow execution, in seconds`},
		"latency_spike_last":                   {txt: `When the latency spike last occurred`, labels: []string{"event_name"}},
		"latency_spike_duration_seconds":       {txt: `Length of the last latency spike in seconds`, labels: []string{"event_name"}},
		"master_link_up":                       {txt: "Master link status on Redis slave"},
		"script_values":                        {txt: "Values returned by the collect script", labels: []string{"key"}},
		"slave_info":                           {txt: "Information about the Redis slave", labels: []string{"master_host", "master_port", "read_only"}},
		"slowlog_last_id":                      {txt: `Last id of slowlog`},
		"slowlog_length":                       {txt: `Total slowlog`},
		"start_time_seconds":                   {txt: "Start time of the Redis instance since unix epoch in seconds."},
		"up":                                   {txt: "Information about the Redis instance"},
		"connected_clients_details":            {txt: "Details about connected clients", labels: []string{"host", "port", "name", "age", "idle", "flags", "db", "cmd"}},
	} {
		tmp := newMetricDescr(opts.Namespace, key, desc.txt, desc.labels)
		exporter.metricDescriptions[key] = tmp
	}

	if exporter.options.MetricsPath == "" {
		exporter.options.MetricsPath = "/metrics"
	}

	exporter.mux = http.NewServeMux()
	if exporter.options.Registry != nil {
		exporter.options.Registry.MustRegister(exporter)
		exporter.mux.Handle(exporter.options.MetricsPath, promhttp.HandlerFor(
			exporter.options.Registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError},
		))

		if !exporter.options.RedisMetricsOnly {
			buildInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: opts.Namespace,
				Name:      "exporter_build_info",
				Help:      "redis exporter build_info",
			}, []string{"version", "commit_sha", "build_date", "golang_version"})
			buildInfo.WithLabelValues(BuildVersion, BuildCommitSha, BuildDate, runtime.Version()).Set(1)
			exporter.options.Registry.MustRegister(buildInfo)
		}
	}

	exporter.mux.HandleFunc("/scrape", exporter.ScrapeHandler)
	exporter.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`ok`))
	})
	exporter.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
<head><title>Redis Exporter ` + BuildVersion + `</title></head>
<body>
<h1>Redis Exporter ` + BuildVersion + `</h1>
<p><a href='` + opts.MetricsPath + `'>Metrics</a></p>
</body>
</html>
`))
	})

	return exporter, nil
}

func (exporter *Exporter) ScrapeHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "'target' parameter must be specified", 400)
		exporter.targetScrapeRequestErrors.Inc()
		return
	}
	if !strings.Contains(target, "://") {
		target = "redis://" + target
	}
	u, err := url.Parse(target)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid 'target' parameter, parse err: %ck ", err), 400)
		exporter.targetScrapeRequestErrors.Inc()
		return
	}
	u.User = nil
	target = u.String()
	opts := exporter.options

	registry := prometheus.NewRegistry()
	opts.Registry = registry

	_, err = NewRedisExporter(target, opts)
	if err != nil {
		http.Error(w, "NewRedisExporter() err: err", 400)
		exporter.targetScrapeRequestErrors.Inc()
		return
	}
	promhttp.HandlerFor(
		registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError},
	).ServeHTTP(w, r)
}

// Describe outputs Redis metric descriptions.
func (exporter *Exporter) Describe(ch chan<- *prometheus.Desc) {

}

// Collect fetches new metrics from the RedisHost and updates the appropriate metrics.
func (exporter *Exporter) Collect(ch chan<- prometheus.Metric) {
	exporter.Lock()
	defer exporter.Unlock()
	exporter.totalScrapes.Inc()

	if exporter.redisAddr != "" {
		start := time.Now().UnixNano()
		var up float64 = 1
		if err := exporter.scrapeRedisHost(ch); err != nil {
			up = 0
			exporter.registerConstMetricGauge(ch, "exporter_last_scrape_error", 1.0, fmt.Sprintf("%s", err))
		} else {
			exporter.registerConstMetricGauge(ch, "exporter_last_scrape_error", 0, "")
		}

		exporter.registerConstMetricGauge(ch, "up", up)
		exporter.registerConstMetricGauge(ch, "exporter_last_scrape_duration_seconds", float64(time.Now().UnixNano()-start)/1000000000)
	}

	ch <- exporter.totalScrapes
	ch <- exporter.scrapeDuration
	ch <- exporter.targetScrapeRequestErrors
}

func (exporter *Exporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	exporter.mux.ServeHTTP(w, r)
}

func (exporter *Exporter) scrapeRedisHost(ch chan<- prometheus.Metric) error {
	conn, err := exporter.connectToRedis()
	if err != nil {
		return err
	}
	defer conn.Close()

	if exporter.options.SetClientName {
		if _, err := doRedisCmd(conn, "CLIENT", "SETNAME", "redis_exporter"); err != nil {
			log.Errorf("Couldn't set client name, err: %s", err)
		}
	}
	dbCount := 0
	if config, err := redis.Strings(doRedisCmd(conn, exporter.options.ConfigCommandName, "GET", "*")); err == nil {
		dbCount, err = exporter.extractConfigMetrics(ch, config)
		if err != nil {
			log.Errorf("Redis CONFIG err: %s", err)
			return err
		}
	} else {
		log.Debugf("Redis CONFIG err: %s", err)
	}
	infoAll, err := redis.String(doRedisCmd(conn, "INFO", "ALL"))
	if err != nil {
		infoAll, err = redis.String(doRedisCmd(conn, "INFO"))
		if err != nil {
			log.Errorf("Redis INFO err: %s", err)
			return err
		}
	}
	if strings.Contains(infoAll, "cluster_enabled:1") {
		if clusterInfo, err := redis.String(doRedisCmd(conn, "CLUSTER", "INFO")); err == nil {
			exporter.extractClusterInfoMetrics(ch, clusterInfo)
			// in cluster mode Redis only supports one database so no extra padding beyond that needed
			dbCount = 1
		} else {
			log.Errorf("Redis CLUSTER INFO err: %s", err)
		}
	} else {
		// in non-cluster mode, if dbCount is zero then "CONFIG" failed to retrieve a valid
		// number of databases and we use the Redis config default which is 16
		if dbCount == 0 {
			dbCount = 16
		}
	}
	exporter.extractInfoMetrics(ch, infoAll, dbCount)
	exporter.extractLatencyMetrics(ch, conn)
	exporter.extractCheckKeyMetrics(ch, conn)
	exporter.extractSlowLogMetrics(ch, conn)
	if exporter.options.LuaScript != nil && len(exporter.options.LuaScript) > 0 {
		if err := exporter.extractLuaScriptMetrics(ch, conn); err != nil {
			return err
		}
	}
	if exporter.options.ExportClientList {
		exporter.extractConnectedClientMetrics(ch, conn)
	}

	if exporter.options.IsTile38 {
		exporter.extractTile38Metrics(ch, conn)
	}

	log.Debugf("scrapeRedisHost() done")
	return nil
}

func doRedisCmd(conn redis.Conn, cmd string, args ...interface{}) (reply interface{}, err error) {
	log.Debugf("c.Do() - running command: %s %s", cmd, args)
	defer log.Debugf("c.Do() - done")

	reply, err = conn.Do(cmd, args)
	if err != nil {
		log.Debugf("c.Do() err: %s", err)
	}
	return reply, err
}

func (exporter *Exporter) registerConstMetricGauge(ch chan<- prometheus.Metric, metric string, val float64, labels ...string) {
	exporter.registerConstMetric(ch, metric, val, prometheus.GaugeValue, labels...)
}

func (exporter *Exporter) registerConstMetric(ch chan<- prometheus.Metric, metric string, val float64, valType prometheus.ValueType, labelValues ...string) {
	descr := exporter.metricDescriptions[metric]
	if descr == nil {
		descr = newMetricDescr(exporter.options.Namespace, metric, metric + " metric", nil)
	}
	if m, err := prometheus.NewConstMetric(descr, valType, val, labelValues...); err == nil {
		ch <- m
	} else {
		log.Debugf("NewConstMetric() err: %s", err)
	}
}

func (exporter *Exporter) connectToRedis() (redis.Conn, error) {
	options := []redis.DialOption{
		redis.DialConnectTimeout(exporter.options.ConnectionTimeouts),
		redis.DialReadTimeout(exporter.options.ConnectionTimeouts),
		redis.DialWriteTimeout(exporter.options.ConnectionTimeouts),
		redis.DialTLSConfig(&tls.Config{
			InsecureSkipVerify: exporter.options.SkipTLSVerification,
			Certificates:       exporter.options.ClientCertificates,
		}),
	}
	if exporter.options.Password != "" {
		options = append(options, redis.DialPassword(exporter.options.Password))
	}
	uri := exporter.redisAddr
	if !strings.Contains(uri, "://") {
		uri = "redis://" + uri
	}
	log.Debugf("redis DialURL:%s", uri)
	conn, err := redis.DialURL(uri, options...)
	if err != nil {
		log.Debugf("DialURL() failed, err: %s", err)
		if frags := strings.Split(exporter.redisAddr, "://"); len(frags) == 2 {
			log.Debugf("Trying: Dial(): %s %s", frags[0], frags[1])
			conn, err = redis.Dial(frags[0], frags[1], options...)
		} else {
			log.Debugf("Trying: Dial(): tcp %s", exporter.redisAddr)
			conn, err = redis.Dial("tcp", exporter.redisAddr, options...)
		}

	}

	return conn, nil
}

//
func (exporter *Exporter) extractConfigMetrics(ch chan<- prometheus.Metric, config []string) (dbCount int, err error) {
	if len(config)%2 != 0 {
		return 0, fmt.Errorf("invalid config: %#v", config)
	}
	for pos := 0; pos < len(config)/2; pos++ {
		strKey := config[pos*2]
		strVal := config[pos*2+1]

		if strKey == "databases" {
			if dbCount, err = strconv.Atoi(strVal); err != nil {
				return 0, fmt.Errorf("invalid config value for key databases: %#v", strVal)
			}
		}
		if !map[string]bool{
			"maxmemory":  true,
			"maxclients": true,
		}[strKey] {
			continue
		}

		if val, err := strconv.ParseFloat(strVal, 64); err == nil {
			exporter.registerConstMetricGauge(ch, fmt.Sprintf("config_%s", config[pos*2]), val)
		}
	}

	return
}

func (exporter *Exporter) extractClusterInfoMetrics(ch chan<- prometheus.Metric, info string) {
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		split := strings.Split(line, ":")
		if len(split) != 2 {
			continue
		}
		fieldKey := split[0]
		fieldValue := split[1]
		if !exporter.includeMetric(fieldKey) {
			continue
		}
		exporter.parseAndRegisterConstMetric(ch, fieldKey, fieldValue)
	}
}

func (exporter *Exporter) includeMetric(str string) bool {
	if strings.HasPrefix(str, "db") || strings.HasPrefix(str, "cmdstat_") || strings.HasPrefix(str, "cluster_") {
		return true
	}
	if _, ok := exporter.metricMapGauges[str]; ok {
		return true
	}
	_, ok := exporter.metricMapCounters[str]
	return ok
}

func sanitizeMetricName(n string) string {
	return prom_strutil.SanitizeLabelName(n)
}

func (exporter *Exporter) parseAndRegisterConstMetric(ch chan<- prometheus.Metric, fieldKey string, fieldValue string) {
	orgMetricName := sanitizeMetricName(fieldKey)
	metricName := orgMetricName
	if newName, ok := exporter.metricMapGauges[metricName]; ok {
		metricName = newName
	} else {
		if newName, ok := exporter.metricMapCounters[metricName]; ok {
			metricName = newName
		}
	}
	var err error
	var val float64
	switch fieldValue {
	case "ok", "true":
		val = 1
	case "err", "fail", "false":
		val = 0
	default:
		val, err = strconv.ParseFloat(fieldValue, 64)
	}
	if err != nil {
		log.Debugf("couldn't parse %s, err: %s", fieldValue, err)
	}
	t := prometheus.GaugeValue
	if exporter.metricMapCounters[orgMetricName] != "" {
		t = prometheus.CounterValue
	}

	switch metricName {
	case "latest_fork_usec":
		metricName = "latest_fork_seconds"
		val = val / 1e6
	}
	exporter.registerConstMetric(ch, metricName, val, t)
}

func (exporter *Exporter) extractInfoMetrics(ch chan<- prometheus.Metric, all string, count int) {

}

func (exporter *Exporter) extractLatencyMetrics(ch chan<- prometheus.Metric, conn redis.Conn) {

}

func (exporter *Exporter) extractCheckKeyMetrics(ch chan<- prometheus.Metric, conn redis.Conn) {

}

func (exporter *Exporter) extractLuaScriptMetrics(ch chan<- prometheus.Metric, conn redis.Conn) error {

	return nil
}

func (exporter *Exporter) extractSlowLogMetrics(ch chan<- prometheus.Metric, conn redis.Conn) {

}

func (exporter *Exporter) extractConnectedClientMetrics(ch chan<- prometheus.Metric, conn redis.Conn) {

}

func (exporter *Exporter) extractTile38Metrics(ch chan<- prometheus.Metric, conn redis.Conn) {

}

func newMetricDescr(namespace string, metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, "", metricName), docString, labels, nil)
}

type dbKeyPair struct {
	db, key string
}

// splitKeyArgs splits a command-line supplied argument into a slice of dbKeyPairs.
func parseKeyArg(keysArgString string) (keys []dbKeyPair, err error) {
	if keysArgString == "" {
		return keys, err
	}
	for _, k := range strings.Split(keysArgString, ",") {
		db := "0"
		key := ""
		frags := strings.Split(k, "=")
		switch len(frags) {
		case 1:
			db = "0"
			key, err = url.QueryUnescape(strings.TrimSpace(frags[0]))
		case 2:
			db = strings.Replace(strings.TrimSpace(frags[0]), "db", "", -1)
			key, err = url.QueryUnescape(strings.TrimSpace(frags[1]))
		default:
			return keys, fmt.Errorf("invalid key list argument: %s", k)
		}
		if err != nil {
			return keys, fmt.Errorf("couldn't parse db/key string: %s", k)
		}
		keys = append(keys, dbKeyPair{
			db:  db,
			key: key,
		})
	}

	return keys, err
}

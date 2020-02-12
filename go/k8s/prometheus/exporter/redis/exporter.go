package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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

var (
	// BuildVersion, BuildDate, BuildCommitSha are filled in by the build script
	BuildVersion   = "<<< filled in by build >>>"
	BuildDate      = "<<< filled in by build >>>"
	BuildCommitSha = "<<< filled in by build >>>"
)

func main() {
	var (
		redisAddr           = flag.String("redis.addr", getEnv("REDIS_ADDR", "redis://localhost:6379"), "Address of the Redis instance to scrape")
		redisPwd            = flag.String("redis.password", getEnv("REDIS_PASSWORD", ""), "Password of the Redis instance to scrape")
		namespace           = flag.String("namespace", getEnv("REDIS_EXPORTER_NAMESPACE", "redis"), "Namespace for metrics")
		checkKeys           = flag.String("check-keys", getEnv("REDIS_EXPORTER_CHECK_KEYS", ""), "Comma separated list of key-patterns to export value and length/size, searched for with SCAN")
		checkSingleKeys     = flag.String("check-single-keys", getEnv("REDIS_EXPORTER_CHECK_SINGLE_KEYS", ""), "Comma separated list of single keys to export value and length/size")
		scriptPath          = flag.String("script", getEnv("REDIS_EXPORTER_SCRIPT", ""), "Path to Lua Redis script for collecting extra metrics")
		listenAddress       = flag.String("web.listen-address", getEnv("REDIS_EXPORTER_WEB_LISTEN_ADDRESS", ":9121"), "Address to listen on for web interface and telemetry.")
		metricPath          = flag.String("web.telemetry-path", getEnv("REDIS_EXPORTER_WEB_TELEMETRY_PATH", "/metrics"), "Path under which to expose metrics.")
		logFormat           = flag.String("log-format", getEnv("REDIS_EXPORTER_LOG_FORMAT", "txt"), "Log format, valid options are txt and json")
		configCommand       = flag.String("config-command", getEnv("REDIS_EXPORTER_CONFIG_COMMAND", "CONFIG"), "What to use for the CONFIG command")
		connectionTimeout   = flag.String("connection-timeout", getEnv("REDIS_EXPORTER_CONNECTION_TIMEOUT", "15s"), "Timeout for connection to Redis instance")
		tlsClientKeyFile    = flag.String("tls-client-key-file", getEnv("REDIS_EXPORTER_TLS_CLIENT_KEY_FILE", ""), "Name of the client key file (including full path) if the server requires TLS client authentication")
		tlsClientCertFile   = flag.String("tls-client-cert-file", getEnv("REDIS_EXPORTER_TLS_CLIENT_CERT_FILE", ""), "Name of the client certificate file (including full path) if the server requires TLS client authentication")
		isDebug             = flag.Bool("debug", getEnvBool("REDIS_EXPORTER_DEBUG"), "Output verbose debug information")
		isTile38            = flag.Bool("is-tile38", getEnvBool("REDIS_EXPORTER_IS_TILE38"), "Whether to scrape Tile38 specific metrics")
		exportClientList    = flag.Bool("export-client-list", getEnvBool("REDIS_EXPORTER_EXPORT_CLIENT_LIST"), "Whether to scrape Client List specific metrics")
		showVersion         = flag.Bool("version", false, "Show version information and exit")
		redisMetricsOnly    = flag.Bool("redis-only-metrics", getEnvBool("REDIS_EXPORTER_REDIS_ONLY_METRICS"), "Whether to also export go runtime metrics")
		inclSystemMetrics   = flag.Bool("include-system-metrics", getEnvBool("REDIS_EXPORTER_INCL_SYSTEM_METRICS"), "Whether to include system metrics like e.g. redis_total_system_memory_bytes")
		skipTLSVerification = flag.Bool("skip-tls-verification", getEnvBool("REDIS_EXPORTER_SKIP_TLS_VERIFICATION"), "Whether to to skip TLS verification")
	)
	flag.Parse()

	switch *logFormat {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.SetFormatter(&log.TextFormatter{})
	}

	log.Printf("Redis Metrics Exporter %s    build date: %s    sha1: %s    Go: %s    GOOS: %s    GOARCH: %s",
		BuildVersion,
		BuildDate,
		BuildCommitSha,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
	if *isDebug {
		log.SetLevel(log.DebugLevel)
		log.Debugln("Enabling debug output")
	} else {
		log.SetLevel(log.InfoLevel)
	}
	if *showVersion {
		return
	}
	to, err := time.ParseDuration(*connectionTimeout)
	if err != nil {
		log.Fatalf("Couldn't parse connection timeout duration, err: %s", err)
	}
	var tlsClientCertificates []tls.Certificate
	if (*tlsClientKeyFile != "") != (*tlsClientCertFile != "") {
		log.Fatal("TLS client key file and cert file should both be present")
	}
	if *tlsClientKeyFile != "" && *tlsClientCertFile != "" {
		cert, err := tls.LoadX509KeyPair(*tlsClientCertFile, *tlsClientKeyFile)
		if err != nil {
			log.Fatalf("Couldn't load TLS client key pair, err: %s", err)
		}
		tlsClientCertificates = append(tlsClientCertificates, cert)
	}
	var ls []byte
	if *scriptPath != "" {
		if ls, err = ioutil.ReadFile(*scriptPath); err != nil {
			log.Fatalf("Error loading script file %s    err: %s", *scriptPath, err)
		}
	}
	registry := prometheus.NewRegistry()
	if !*redisMetricsOnly {
		registry = prometheus.DefaultRegisterer.(*prometheus.Registry)
	}
	exporter, err := NewRedisExporter(
		*redisAddr,
		ExporterOptions{
			Password:            *redisPwd,
			Namespace:           *namespace,
			ConfigCommandName:   *configCommand,
			CheckKeys:           *checkKeys,
			CheckSingleKeys:     *checkSingleKeys,
			LuaScript:           ls,
			InclSystemMetrics:   *inclSystemMetrics,
			IsTile38:            *isTile38,
			ExportClientList:    *exportClientList,
			SkipTLSVerification: *skipTLSVerification,
			ClientCertificates:  tlsClientCertificates,
			ConnectionTimeouts:  to,
			MetricsPath:         *metricPath,
			RedisMetricsOnly:    *redisMetricsOnly,
			Registry:            registry,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Providing metrics at %s%s", *listenAddress, *metricPath)
	log.Debugf("Configured redis addr: %#v", *redisAddr)
	log.Fatal(http.ListenAndServe(*listenAddress, exporter))
}

func getEnv(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}
	return defaultVal
}

func getEnvBool(key string) (res bool) {
	if envVal, ok := os.LookupEnv(key); ok {
		res, _ = strconv.ParseBool(envVal)
	}
	return res
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
		txt  string
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
	
}

func (exporter *Exporter) registerConstMetricGauge(ch chan<- prometheus.Metric, metric string, val float64, labels ...string) {
	exporter.registerConstMetric(ch, metric, val, prometheus.GaugeValue, labels...)
}

func (exporter *Exporter) registerConstMetric(ch chan<- prometheus.Metric, metric string, val float64, value prometheus.ValueType, labels ...string) {
	
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

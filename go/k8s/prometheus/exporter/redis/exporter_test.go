package main

/*
  to run the tests with redis running on anything but localhost:6379 use
  $ go test   --redis.addr=<host>:<port>
  for html coverage report run
  $ go test -coverprofile=coverage.out  && go tool cover -html=coverage.out
*/

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	TestValue   = 1234.56
	TimeToSleep = 200
)

var (
	keys         []string
	keysExpiring []string
	listKeys     []string
	ts           = int32(time.Now().Unix())

	dbNumStr     = "11"
	altDBNumStr  = "12"
	dbNumStrFull = fmt.Sprintf("db%s", dbNumStr)
)

const (
	TestSetName = "test-set"
)

func init() {
	_ = os.Setenv("TEST_REDIS_URI", "redis://localhost:6379")
	_ = os.Setenv("TEST_REDIS_CLUSTER_SLAVE_URI", "redis://localhost:6379")
	_ = os.Setenv("TEST_PWD_REDIS_URI", "redis://localhost:6380 -a redis-password")

	ll := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if pl, err := log.ParseLevel(ll); err == nil {
		log.Printf("Setting log level to: %s", ll)
		log.SetLevel(pl)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	for _, n := range []string{"john", "paul", "ringo", "george"} {
		key := fmt.Sprintf("key_%s_%d", n, ts)
		keys = append(keys, key)
	}

	listKeys = append(listKeys, "beatles_list")
	for _, n := range []string{"A.J.", "Howie", "Nick", "Kevin", "Brian"} {
		key := fmt.Sprintf("key_exp_%s_%d", n, ts)
		keysExpiring = append(keysExpiring, key)
	}
}

func TestHTTPHTMLPages(t *testing.T) {
	if os.Getenv("TEST_PWD_REDIS_URI") == "" {
		t.Skipf("TEST_PWD_REDIS_URI not set - skipping")
	}

	exporter, _ := NewRedisExporter(os.Getenv("TEST_PWD_REDIS_URI"), ExporterOptions{Namespace: "test", Registry: prometheus.NewRegistry()})
	ts := httptest.NewServer(exporter)
	defer ts.Close()

	for _, tst := range []struct {
		path string
		want string
	}{
		{
			path: "/",
			want: `<head><title>Redis Exporter `,
		},
		{
			path: "/health",
			want: `ok`,
		},
	} {
		t.Run(fmt.Sprintf("path: %s", tst.path), func(t *testing.T) {
			body := downloadURL(t, ts.URL+tst.path)
			if !strings.Contains(body, tst.want) {
				t.Fatalf(`error, expected string "%s" in body, got body: \n\n%s`, tst.want, body)
			}
		})
	}
}

func TestCheckKeys(test *testing.T) {
	for _, tst := range []struct {
		SingleCheckKey string
		CheckKeys      string
		ExpectSuccess  bool
	}{
		{"", "", true},
		{"db1=key3", "", true},
		{"check-key-01", "", true},
		{"", "check-key-02", true},
		{"wrong=wrong=1", "", false},
		{"", "wrong=wrong=2", false},
	} {
		_, err := NewRedisExporter(os.Getenv("TEST_REDIS_URI"), ExporterOptions{Namespace: "test", CheckSingleKeys: tst.SingleCheckKey, CheckKeys: tst.CheckKeys})
		if tst.ExpectSuccess && err != nil {
			test.Errorf("Expected success for test: %#v, got err: %s", tst, err)
			return
		}

		if !tst.ExpectSuccess && err == nil {
			test.Errorf("Expected failure for test: %#v, got no err", tst)
			return
		}
	}
}

func TestClusterSlave(t *testing.T) {
	if os.Getenv("TEST_REDIS_CLUSTER_SLAVE_URI") == "" {
		t.Skipf("TEST_REDIS_CLUSTER_SLAVE_URI not set - skipping")
	}

	addr := os.Getenv("TEST_REDIS_CLUSTER_SLAVE_URI")
	exporter, _ := NewRedisExporter(addr, ExporterOptions{Namespace: "test", Registry: prometheus.NewRegistry()})
	ts := httptest.NewServer(exporter)
	defer ts.Close()

	chM := make(chan prometheus.Metric, 10000)
	go func() {
		exporter.Collect(chM)
		close(chM)
	}()

	body := downloadURL(t, ts.URL+"/metrics")
	log.Debugf("slave - body: %s", body)
	for _, want := range []string{
		"test_instance_info",
		"test_master_last_io_seconds",
		"test_slave_info",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("Did not find key [%s] \nbody: %s", want, body)
		}
	}
}

func TestPasswordInvalid(t *testing.T) {
	if os.Getenv("TEST_PWD_REDIS_URI") == "" {
		t.Skipf("TEST_PWD_REDIS_URI not set - skipping")
	}
	testPwd := "redis-password"
	uri := strings.Replace(os.Getenv("TEST_PWD_REDIS_URI"), testPwd, "wrong-pwd", -1)

	exporter, _ := NewRedisExporter(uri, ExporterOptions{Namespace: "test", Registry: prometheus.NewRegistry()})
	ts := httptest.NewServer(exporter)
	defer ts.Close()

	chM := make(chan prometheus.Metric, 10000)
	go func() {
		exporter.Collect(chM)
		close(chM)
	}()

	want := `test_exporter_last_scrape_error{err="dial redis: unknown network redis"} 1`
	body := downloadURL(t, ts.URL+"/metrics")
	if !strings.Contains(body, want) {
		t.Errorf(`error, expected string "%s" in body, got body: \n\n%s`, want, body)
	}
}

func TestHTTPScrapeMetricsEndpoints(test *testing.T) {
	_ = setupDBKeys(test, os.Getenv("TEST_REDIS_URI"))
	defer deleteKeysFromDB(test, os.Getenv("TEST_REDIS_URI"))
	_ = setupDBKeys(test, os.Getenv("TEST_REDIS_URI"))
	defer deleteKeysFromDB(test, os.Getenv("TEST_REDIS_URI"))

	csk := dbNumStrFull + "=" + url.QueryEscape(keys[0])
	testRedisIPAddress := ""
	testRedisHostname := ""
	for _, tst := range []struct {
		addr   string
		ck     string
		csk    string
		pwd    string
		target string
	}{
		{addr: os.Getenv("TEST_REDIS_URI"), csk: csk},
		{addr: testRedisIPAddress, csk: csk},
		{addr: testRedisHostname, csk: csk},
		{addr: os.Getenv("TEST_REDIS_URI"), ck: csk},
		{pwd: "", target: os.Getenv("TEST_REDIS_URI"), ck: csk},
		{pwd: "", target: os.Getenv("TEST_REDIS_URI"), csk: csk},
		{pwd: "redis-password", target: os.Getenv("TEST_PWD_REDIS_URI"), csk: csk},
	}{
		name := fmt.Sprintf("addr:[%s]___target:[%s]___pwd:[%s]", tst.addr, tst.target, tst.pwd)
		test.Run(name, func(test *testing.T) {
			options := ExporterOptions{
				Namespace: "test",
				Password:  tst.pwd,
				LuaScript: []byte(`return {"a", "11", "b", "12", "c", "13"}`),
				Registry:  prometheus.NewRegistry(),
			}
			if tst.target == "" {
				options.CheckSingleKeys = tst.csk
				options.CheckKeys = tst.ck
			}

			exporter, _ := NewRedisExporter(tst.addr, options)
			ts := httptest.NewServer(exporter)
			uri := ts.URL
			if tst.target != "" {
				uri += "/scrape"
				v := url.Values{}
				v.Add("target", tst.target)
				v.Add("check-single-keys", tst.csk)
				v.Add("check-keys", tst.ck)

				up, _ := url.Parse(uri)
				up.RawQuery = v.Encode()
				uri = up.String()
			} else {
				uri += "/metrics"
			}
			wants := []string{
				// metrics
				`test_connected_clients`,
				`test_commands_processed_total`,
				`test_instance_info`,

				"db_keys",
				"db_avg_ttl_seconds",
				"cpu_sys_seconds_total",
				"loading_dump_file", // testing renames
				"config_maxmemory",  // testing config extraction
				"config_maxclients", // testing config extraction
				"slowlog_length",
				"slowlog_last_id",
				"start_time_seconds",
				"uptime_in_seconds",

				// labels and label values
				`redis_mode`,
				`standalone`,
				`cmd="config`,
				`test_script_value`, // lua script
				`test_key_size{db="db11",key="` + keys[0] + `"} 7`,
				`test_key_value{db="db11",key="` + keys[0] + `"} 1234.56`,

				`test_db_keys{db="db11"} `,
				`test_db_keys_expiring{db="db11"} `,
			}
			body := downloadURL(test, uri)
			for _, want := range wants {
				if !strings.Contains(body, want) {
					test.Errorf("url: %s    want metrics to include %q, have:\n%s", uri, want, body)
					break
				}
			}
			ts.Close()
		})
	}
}

func deleteKeysFromDB(t *testing.T, addr string) error {
	c, err := redis.DialURL(addr)
	if err != nil {
		t.Errorf("couldn't setup redis, err: %s ", err)
		return err
	}
	defer c.Close()

	if _, err := c.Do("SELECT", dbNumStr); err != nil {
		log.Printf("deleteKeysFromDB() - couldn't setup redis, err: %s ", err)
		// not failing on this one - cluster doesn't allow for SELECT so we log and ignore the error
	}

	for _, key := range keys {
		_, _ = c.Do("DEL", key)
	}

	for _, key := range keysExpiring {
		_, _ = c.Do("DEL", key)
	}

	for _, key := range listKeys {
		_, _ = c.Do("DEL", key)
	}

	_, _ = c.Do("DEL", TestSetName)
	return nil
}

func setupDBKeys(t *testing.T, uri string) error {
	c, err := redis.DialURL(uri)
	if err != nil {
		t.Errorf("couldn't setup redis for uri %s, err: %s ", uri, err)
		return err
	}
	defer c.Close()

	if _, err := c.Do("SELECT", dbNumStr); err != nil {
		log.Printf("setupDBKeys() - couldn't setup redis, err: %s ", err)
		// not failing on this one - cluster doesn't allow for SELECT so we log and ignore the error
	}
	for _, key := range keys {
		_, err = c.Do("SET", key, TestValue)
		if err != nil {
			t.Errorf("couldn't setup redis, err: %s ", err)
			return err
		}
	}
	// setting to expire in 300 seconds, should be plenty for a test run
	for _, key := range keysExpiring {
		_, err = c.Do("SETEX", key, "300", TestValue)
		if err != nil {
			t.Errorf("couldn't setup redis, err: %s ", err)
			return err
		}
	}
	for _, key := range listKeys {
		for _, val := range keys {
			_, err = c.Do("LPUSH", key, val)
			if err != nil {
				t.Errorf("couldn't setup redis, err: %s ", err)
				return err
			}
		}
	}

	_, _ = c.Do("SADD", TestSetName, "test-val-1")
	_, _ = c.Do("SADD", TestSetName, "test-val-2")

	time.Sleep(time.Millisecond * 50)

	return nil
}

func downloadURL(t *testing.T, url string) string {
	log.Debugf("downloadURL() %s", url)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

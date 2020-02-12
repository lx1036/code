package main

/*
  to run the tests with redis running on anything but localhost:6379 use
  $ go test   --redis.addr=<host>:<port>
  for html coverage report run
  $ go test -coverprofile=coverage.out  && go tool cover -html=coverage.out
*/

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
	_ = os.Setenv("TEST_PWD_REDIS_URI", "redis://localhost:6379")
	_ = os.Setenv("TEST_REDIS_CLUSTER_SLAVE_URI", "redis://localhost:6379")

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

func TestCheckKeys(test *testing.T)  {
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

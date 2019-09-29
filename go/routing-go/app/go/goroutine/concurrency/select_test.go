package concurrency

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

var tenSecondTimeout = 10 * time.Second

func TestRacer(test *testing.T) {
    /*slowURL := "http://www.baisu.com"
    fastURL := "http://www.so.com"

    want := fastURL
    got := Racer(slowURL, fastURL)

    if got != want {
        test.Errorf("got '%s', want '%s'", got, want)
    }*/

    /*slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       time.Sleep(20 * time.Millisecond)
       w.WriteHeader(http.StatusOK)
    }))
    fastServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       //time.Sleep(40 * time.Millisecond)
       w.WriteHeader(http.StatusOK)
    }))
    slowURL := slowServer.URL
    fastURL := fastServer.URL
    want := fastURL
    got := Racer(slowURL, fastURL)
    if got != want {
       test.Errorf("got '%s', want '%s'", got, want)
    }
    slowServer.Close()
    fastServer.Close()*/

    /*slowServer := makeDelayedServer(20 * time.Millisecond)
    fastServer := makeDelayedServer(0 * time.Millisecond)
    defer slowServer.Close()
    defer fastServer.Close()
    slowURL := slowServer.URL
    fastURL := fastServer.URL
    want := fastURL
    got := Racer(slowURL, fastURL)
    if got != want {
        test.Errorf("got '%s', want '%s'", got, want)
    }*/



    test.Run("returns an error if a server doesn't respond within 10s", func(test *testing.T) {
        serverA := makeDelayedServer(11 * time.Second)
        serverB := makeDelayedServer(12 * time.Second)

        defer serverA.Close()
        defer serverB.Close()

        _, err := Racer(serverA.URL, serverB.URL, tenSecondTimeout)

        if err == nil {
            test.Error("expected an error but didn't get one")
        }
    })
}

func makeDelayedServer(delay time.Duration) *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(delay)
        w.WriteHeader(http.StatusOK)
    }))
}

package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"sync"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// install prometheus
	prometheus.Init(prometheus.Options{
		AppName: "meetup-2020-04-18",
		Idc:     "dev",
		WatchPath: map[string]struct{}{
			"/api/v1/hello": {},
		},
		HistogramBuckets: []float64{
			10, 13, 17, 22, 28, 36, 46, 59, 76, 98, 127, 164, 212, 275, 356, 461, 597, 773, 1000, 1001,
		},
	})
	router.Use(prometheus.MiddlewarePrometheusAccessLogger())
	prometheus.MetricsServerStart("/metrics", 18081)

	routerV1 := router.Group("/api/v1")
	routerV1.GET("/hello", func(context *gin.Context) {
		var wg sync.WaitGroup
		wg.Add(1)
		var body1, body2 string
		go func() {
			defer wg.Done()
			var err error
			var body []byte
			body, err = httplib.Get("https://api.github.com/repos/lx1036/code/commits?per_page=3&sha=master").Bytes()
			type Commit struct {
				Url string `json:"url"`
			}
			var commit []Commit
			err = json.Unmarshal(body, &commit)
			if err != nil {
				body1 = "https://www.default.com"
				return
			}
			body1 = commit[0].Url
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			var body []byte
			body, err = httplib.Get("https://api.github.com/users/lx1036").Bytes()
			type User struct {
				Login string `json:"login"`
			}
			var user User
			err = json.Unmarshal(body, &user)
			if err != nil {
				body2 = "default"
				return
			}
			body2 = user.Login
		}()
		wg.Wait()

		body := body1 + ":" + body2
		//_, _ = context.Writer.Write([]byte(body))
		context.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": body,
		})
	})

	fmt.Println(router.Run())
}

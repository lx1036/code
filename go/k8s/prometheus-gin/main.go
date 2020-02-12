package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"k8s-lx1036/k8s/prometheus-gin/prometheus"
	"net/http"
	"net/url"
	"sort"
	"time"
)

func ginDemo() {
	app := gin.Default()

	prometheus.Init(prometheus.Options{
		AppName: "Prometheus-Gin",
		Idc:     "beijing",
		WatchPath: map[string]struct{}{
			"/hello": {},
		},
		HistogramBuckets: []float64{0.001, 0.05, 0.1, 1},
	})
	prometheus.MetricsServerStart("/metrics", 18081)
	app.Use(prometheus.MiddlewarePrometheusAccessLogger())

	app.POST("/person", func(context *gin.Context) {
		id := context.PostForm("id")
		fmt.Println(id)
	})

	iRoutes := app.GET("/ping", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": Person{
				Name: "lx1036",
				Age:  29,
			},
		})
	})

	iRoutes.GET("/hello", func(context *gin.Context) {
		time.Sleep(time.Millisecond)
		context.JSON(http.StatusOK, gin.H{
			"data": "world",
		})
	})

	routes := app.Routes()
	for _, route := range routes {
		fmt.Printf("Method: %s Path: %s Handler: %s HandlerFunc: %T \n", route.Method, route.Path, route.Handler, route.HandlerFunc)
	}

	err := app.Run(":8080")
	if err != nil {
		fmt.Printf("uncaught error: %v", err)
	}
}

func pkgNetHttpDemo() {
	http.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		io.WriteString(writer, fmt.Sprintf("world from %s %s", request.Method, request.URL))
	})

	http.ListenAndServe(":9090", nil)
}

/**
https://www.jianshu.com/p/b38b1719636e
https://github.com/qcrao/Go-Questions/blob/master/interface/Go%20%E6%8E%A5%E5%8F%A3%E4%B8%8E%20C%2B%2B%20%E6%8E%A5%E5%8F%A3%E6%9C%89%E4%BD%95%E5%BC%82%E5%90%8C.md
https://github.com/developer-learning/night-reading-go/issues/393
*/
type Person struct {
	Name string
	Age  int
}

func (p Person) String() string {
	return fmt.Sprintf("%s: %d", p.Name, p.Age)
}

// ByAge implements sort.Interface for []Person based on
// the Age field.
type ByAge []Person //自定义

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }

/*
https://golang.org/pkg/net/http/
*/
func main() {
	people := []Person{
		{"Bob", 31},
		{"John", 42},
		{"Michael", 17},
		{"Jenny", 26},
	}

	fmt.Println(people)
	sort.Sort(ByAge(people))
	fmt.Println(people)

	ginDemo()

	_, _ = http.PostForm("localhost:9090/person", url.Values{
		"id": {"1"},
	})

	//pkgNetHttpDemo()
}

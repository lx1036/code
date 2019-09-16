package main

import (
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

/*
Gin Global Error Handler
 */

func logGin()  {
	gin.DisableConsoleColor()
	file, _ := os.Create("gin.log")
	stat, _ := file.Stat()
	fmt.Println("Mode: ", stat.Mode())
	gin.DefaultWriter = io.MultiWriter(file)
	router := gin.Default()
	router.GET("/", func(context *gin.Context) {
		context.String(200, "OK")
	})
	router.GET("ping", func(context *gin.Context) {
		context.String(200, "Pong")
	})
	router.Run(":8080")
}

func pkgLog()  {
	err := os.Chmod("./test.txt", 0777)
	if err != nil {
		fmt.Println("os chmod file failed")
		log.Fatal(err)
	}

	file, err := os.Open("./test.txt")
	if err != nil {
		fmt.Println("open file failed")
		log.Fatal(err)
	}

	err = file.Chmod(0777)
	if err != nil {
		fmt.Println("chmod failed")
		log.Fatal(err)
	}

	_, err = file.Write([]byte("abc"))
	if err != nil {
		fmt.Println("write bytes failed")
		log.Fatal(err, file.Fd())
	}
	/*
		write bytes failed
		2019/09/04 12:33:31 write ./test.txt: bad file descriptor 3
		exit status 1
	*/

	fmt.Println(file.Name())
}

func pkgErrors()  {
	err := errors.New("Error/Exception")
	fmt.Println(err)
}

/*
SENTRY_DSN, SENTRY_RELEASE, SENTRY_ENVIRONMENT
 */
func getSentryDsn() string  {
	viper.AddConfigPath(".")
	viper.SetConfigFile("env.json")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Read config file failed:%v\n", err))
	}
	sentryDsn := viper.Get("sentry_dsn")

	return cast.ToString(sentryDsn)
}

func sentryGo()  {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              getSentryDsn(),
		Debug:            true,
		AttachStacktrace: true,
		SampleRate:       0,
		IgnoreErrors:     nil,
		BeforeSend:       nil,
		BeforeBreadcrumb: nil,
		Integrations:     nil,
		DebugWriter:      nil,
		Transport:        nil,
		ServerName:       "",
		Release:          "",
		Dist:             "",
		Environment:      "",
		MaxBreadcrumbs:   0,
		HTTPTransport:    nil,
		HTTPProxy:        "",
		HTTPSProxy:       "",
		CaCerts:          nil,
	})
	if err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	_, err = os.Open("./sentry2.txt")
	if err != nil {
		fmt.Printf("Open file failed: %v\n", err)
		/*
		https://sentry.io/organizations/leftcapital/issues/?project=1550933&query=is%3Aunresolved&statsPeriod=14d
		 */
		sentry.CaptureException(err)
		sentry.Flush(time.Second * 5)
	}
}

func sentryGin()  {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              getSentryDsn(),
		Debug:            true,
		AttachStacktrace: true,
		SampleRate:       0,
		IgnoreErrors:     nil,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			if hint.Context != nil {
				if request, ok :=hint.Context.Value(sentry.RequestContextKey).(*http.Request); ok {
					fmt.Println("request: ", request)
				}
			}

			fmt.Println("event: ", event)

			return event
		},
		BeforeBreadcrumb: nil,
		Integrations:     nil,
		DebugWriter:      nil,
		Transport:        nil,
		ServerName:       "",
		Release:          "",
		Dist:             "",
		Environment:      "",
		MaxBreadcrumbs:   0,
		HTTPTransport:    nil,
		HTTPProxy:        "",
		HTTPSProxy:       "",
		CaCerts:          nil,
	})

	if err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	app := gin.Default()
	app.Use(sentrygin.New(sentrygin.Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         0,
	}))
	app.Use(func(context *gin.Context) {
		if hub := sentrygin.GetHubFromContext(context); hub != nil {
			hub.Scope().SetTag("someTag", "TagValue")
		}

		context.Next()
	})
	app.GET("/", func(context *gin.Context) {
		if hub := sentrygin.GetHubFromContext(context); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetExtra("someScope", "someScopeValue")
				hub.CaptureMessage("test test test")
			})
		}

		context.Status(http.StatusOK)
	})

	app.GET("/foo", func(context *gin.Context) {
		panic("foo bar") // sentrygin handler will catch it
	})

	app.Run(":8080")
}

func debugHelloWorld()  {
	fmt.Println("Debug")
}

func main() {
	//pkgLog()

	logGin()

	//pkgErrors()

	//sentryGo()

	//sentryGin()

	//debugHelloWorld()
}

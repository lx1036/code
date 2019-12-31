package cors

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

var allHeaders = []string{
	"Vary",
	"Access-Control-Allow-Origin",
	"Access-Control-Allow-Methods",
	"Access-Control-Allow-Headers",
	"Access-Control-Allow-Credentials",
	"Access-Control-Max-Age",
	"Access-Control-Expose-Headers",
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func TestDemo(test *testing.T) {
	router := gin.Default()

	router.Use(CorsMiddleware())
	router.GET("/", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"hello": "world"})
	})

	//router.Run(":8080")
}

func TestAllowAllNotNil(test *testing.T) {
	handler := WrapperAllowAll()
	if handler == nil {
		test.Error("Want not nil handler, got nil")
	}
}

func TestAbortsWhenPreflightRequest(test *testing.T) {
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request, _ = http.NewRequest(http.MethodOptions, "http://localhost:8080/foo", nil)
	context.Request.Header.Add("Origin", "http://localhost:8080")
	context.Request.Header.Add("Access-Control-Request-Method", http.MethodPost)
	context.Status(http.StatusAccepted)
	response.Code = http.StatusAccepted

	handler := corsWrapper{
		Cors: New(Options{}),
	}.build()

	handler(context)

	if !context.IsAborted() {
		test.Error("Should abort preflight request")
	}

	logrus.WithFields(logrus.Fields{
		"code": response.Code,
	}).Info("preflight-request response code: ")

	if response.Code != http.StatusOK {
		test.Error("Should abort preflight request with status code " + strconv.Itoa(http.StatusOK))
	}
}

func TestNotAbortsWhenPassthrough(test *testing.T) {
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request, _ = http.NewRequest(http.MethodOptions, "http://localhost:8080/foo", nil)
	context.Request.Header.Add("Origin", "http://localhost:8080")
	context.Request.Header.Add("Access-Control-Request-Method", http.MethodPost)

	handler := corsWrapper{
		Cors: New(Options{
			OptionsPassthrough: true,
		}),
		OptionPassthrough: true,
	}.build()

	handler(context)

	if context.IsAborted() {
		test.Error("Should not abort preflight request when OPTIONS passthrough is enabled")
	}
}

func TestPreflightRequestInvalidOrigin(test *testing.T) {
	cors := New(Options{
		AllowedOrigins: []string{"http://foo.com"},
	})
	response := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodOptions, "http://bar.com/foo", nil)
	request.Header.Add("origin", "http://bar.com")

	cors.HandlePreflightRequest(response, request)

	exposedHeaders := map[string]string{
		"Vary": "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
	}

	assertHeaders(test, response.Header(), exposedHeaders)
}

func TestPreflightNoOptionsAbortion(test *testing.T) {
	cors := Default()
	response := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodGet, "http://bar.com/foo", nil)

	cors.HandlePreflightRequest(response, request)

	assertHeaders(test, response.Header(), map[string]string{})
}

func assertHeaders(test *testing.T, headers http.Header, exposedHeaders map[string]string) {
	for _, name := range allHeaders {
		want := exposedHeaders[name]
		got := strings.Join(headers[name], ", ")
		if want != got {
			test.Errorf("Header %s: got %s, want %s", name, got, want)
		}
	}
}

package mock

import (
	"bou.ke/monkey"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func SetupGin() *gin.Engine {
	engine := gin.New()
	engine.GET("/hello", func(context *gin.Context) {
		body1, _ := httplib.Get("https://api.github.com/repos/lx1036/code/commits?per_page=3&sha=master").Bytes()
		body2, _ := httplib.Get("https://api.github.com/users/lx1036").Bytes()

		body := string(body1) + ":" + string(body2)
		_, _ = context.Writer.Write([]byte(body))
	})

	return engine
}

func TestGin(test *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := SetupGin()
	request := httptest.NewRequest("GET", "/hello", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	response := recorder.Result()
	body, _ := ioutil.ReadAll(response.Body)
	var commits []struct {
		Sha    string `json:"sha"`
		Url    string `json:"url"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
	}
	_ = json.Unmarshal(body, &commits)

	assert.Equal(test, 3, len(commits))
	assert.Equal(test, "lx1036", commits[0].Author.Login)
}

func TestMockery(test *testing.T) {
	var guard *monkey.PatchGuard
	guard = monkey.PatchInstanceMethod(reflect.TypeOf(http.DefaultClient), "Get", func(c *http.Client, url string) (*http.Response, error) {
		guard.Unpatch()
		defer guard.Restore()

		if !strings.HasPrefix(url, "https://") {
			return nil, fmt.Errorf("only https requests allowed")
		}

		recorder := httptest.NewRecorder()
		_, _ = recorder.Write([]byte("hello"))
		response := recorder.Result()

		return response, nil
	})

	_, err := http.Get("http://google.com")
	fmt.Println(err) // only https requests allowed
	response, err := http.Get("https://google.com")
	body, _ := ioutil.ReadAll(response.Body)
	assert.Equal(test, "hello", string(body))
	assert.Equal(test, 200, response.StatusCode)
}

// SUCCESS!!!
// https://github.com/bouk/monkey/blob/master/README.md
func TestDoRequest(test *testing.T) {
	/*var guard *monkey.PatchGuard
	guard = monkey.PatchInstanceMethod(reflect.TypeOf(http.DefaultClient), "Do", func(c *http.Client, request *http.Request) (*http.Response, error) {
		guard.Unpatch()
		defer guard.Restore()

		url := request.URL.String()

		fmt.Println(url)

		if !strings.HasPrefix(url, "https://") {
			return nil, fmt.Errorf("only https requests allowed")
		}

		recorder := httptest.NewRecorder()
		_, _ = recorder.Write([]byte("hello"))
		response := recorder.Result()

		return response, nil
	})*/

	var req *httplib.BeegoHTTPRequest
	monkey.PatchInstanceMethod(reflect.TypeOf(req), "Bytes", func(request *httplib.BeegoHTTPRequest) ([]byte, error) {
		url := request.GetRequest().URL.String()
		var body string
		if strings.Contains(url, "https://api.github.com/repos") {
			body = "hello"
		}

		if strings.Contains(url, "https://api.github.com/users") {
			body = "world"
		}

		return []byte(body), nil
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(req), "Bytes")

	gin.SetMode(gin.TestMode)
	engine := SetupGin()
	request := httptest.NewRequest("GET", "/hello", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	response := recorder.Result()
	body, _ := ioutil.ReadAll(response.Body)
	assert.Equal(test, "hello:world", string(body))

	//url := "https://www.so.com"
	//req, err := http.NewRequest("GET", url, nil)
	//if err != nil {
	//}
	//
	//response, err := http.DefaultClient.Do(req)
	//fmt.Println(err) // only https requests allowed
	////response, err := http.Get("https://google.com")
	//body, _ := ioutil.ReadAll(response.Body)
	//assert.Equal(test, 200, response.StatusCode)
}

func TestMockGin(test *testing.T) {
	gin.SetMode(gin.TestMode)
	//req := GetRequest().GetRequest()

	var guard *monkey.PatchGuard
	guard = monkey.PatchInstanceMethod(reflect.TypeOf(http.DefaultClient), "Do", func(c *http.Client, request *http.Request) (*http.Response, error) {
		guard.Unpatch()
		defer guard.Restore()
		recorder := httptest.NewRecorder()
		var response *http.Response
		url := request.URL.String()
		if !strings.HasPrefix(url, "https://") {
			return nil, fmt.Errorf("only https requests allowed")
		}

		fmt.Println(url, strings.Contains(url, "https://api.github.com"))

		if strings.Contains(url, "https://api.github.com") {
			_, _ = recorder.Write([]byte("hello"))
			response = recorder.Result()
		}

		return response, nil
	})

	engine := SetupGin()
	request := httptest.NewRequest("GET", "/hello", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	response := recorder.Result()
	body, _ := ioutil.ReadAll(response.Body)
	assert.Equal(test, "hello", string(body))
}

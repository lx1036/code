package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	_ "k8s-lx1036/k8s-ui/backend/controllers/auth"
	"k8s-lx1036/k8s-ui/backend/models"
	routers_gin "k8s-lx1036/k8s-ui/backend/routers-gin"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

type AuthSuite struct {
	suite.Suite
	routers *gin.Engine
}

func (suite *AuthSuite) SetupTest() {
	suite.routers = routers_gin.SetupRouter()
}

func (suite *AuthSuite) TeardownTest() {

}

func (suite *AuthSuite) TestCors() {
	//routers := routers_gin.SetupRouter()
	data := url.Values{}
	data.Add("username", "admin")
	data.Add("password", "password")
	request := httptest.NewRequest("OPTIONS", "/me", strings.NewReader(data.Encode()))
	request.Header.Add("Access-Control-Request-Method", http.MethodGet)
	request.Header.Add("Access-Control-Request-Headers", "authorization,content-type")
	//request.Header.Add("Access-Control-Request-Headers", "content-type")
	request.Header.Add("Origin", "http://localhost:4200")
	request.Header.Add("Host", "localhost:8080")
	recorder := httptest.NewRecorder()
	suite.routers.ServeHTTP(recorder, request)
	//response := recorder.Result()
	headers := recorder.Header()
	for key, value := range headers {
		fmt.Println(key, value)
	}

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	assert.Equal(suite.T(), "*", recorder.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(suite.T(), http.MethodGet, recorder.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(suite.T(), "Authorization, Content-Type", recorder.Header().Get("Access-Control-Allow-Headers"))
}

func (suite *AuthSuite) TestLogin() {
	body := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: "admin",
		Password: "password",
	}
	requestBody, _ := json.Marshal(body)
	request := httptest.NewRequest("POST", "/login/db", bytes.NewBuffer(requestBody))
	request.Header.Set("content-type", "application/json")
	recorder := httptest.NewRecorder()
	suite.routers.ServeHTTP(recorder, request)
	response := recorder.Result()
	responseBody, _ := ioutil.ReadAll(response.Body)

	var token struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	_ = json.Unmarshal(responseBody, &token)

	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *AuthSuite) TestCurrentUser() {
	request := httptest.NewRequest("GET", "/me", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", Token))
	recorder := httptest.NewRecorder()
	suite.routers.ServeHTTP(recorder, request)
	response := recorder.Result()
	body, _ := ioutil.ReadAll(response.Body)

	var me struct {
		Errno  int    `json:"errno"`
		Errmsg string `json:"errmsg"`
		Data   struct {
			ID int `json:"ID"`
		} `json:"data"`
	}
	_ = json.Unmarshal(body, &me)

	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
	assert.Equal(suite.T(), 1, me.Data.ID)
}

func (suite *AuthSuite) TestNotificationSubscribe() {
	request := httptest.NewRequest("GET", "/api/v1/notifications/subscribe", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", Token))
	recorder := httptest.NewRecorder()
	suite.routers.ServeHTTP(recorder, request)
	response := recorder.Result()
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	type NotificationLogs struct {
		Errno  int                      `json:"errno"`
		Errmsg string                   `json:"errmsg"`
		Data   []models.NotificationLog `json:"data"`
	}
	var notificationLogs NotificationLogs
	_ = json.Unmarshal(body, &notificationLogs)

	//Assert(response.StatusCode)
	//Assert(response.Header)
	//Assert(notificationLogs.Data)

	var slices []interface{}
	slices = append(slices, response.StatusCode)
	slices = append(slices, response.Header)
	slices = append(slices, notificationLogs)

	path := getBaselineDataFile()
	if !assert.FileExists(suite.T(), path) || rebase { // create baseline use first actual as next expected
		body2, _ := json.Marshal(slices)
		var buffer = new(bytes.Buffer)
		_ = json.Indent(buffer, body2, "", "  ")
		_ = ioutil.WriteFile(path, buffer.Bytes(), 0666)
	} else if rebase { // new actual override baseline as next expected

	} else {
		var decoded []json.RawMessage
		baseline, _ := ioutil.ReadFile(path)
		err := json.Unmarshal(baseline, &decoded)
		if err != nil {
			panic(err)
		}

		var code int
		_ = json.Unmarshal(decoded[0], &code)
		assert.EqualValues(suite.T(), code, response.StatusCode)

		var header http.Header
		_ = json.Unmarshal(decoded[1], &header)
		assert.EqualValues(suite.T(), header, response.Header)

		assert.EqualValues(suite.T(), strings.ReplaceAll(strings.ReplaceAll(string(decoded[2]), "\n", ""), " ", ""), strings.ReplaceAll(strings.ReplaceAll(string(body), "\n", ""), " ", ""))

		var content NotificationLogs
		_ = json.Unmarshal(decoded[2], &content)
		assert.EqualValues(suite.T(), content, notificationLogs)
	}
}

var rebase bool = false

func Assert(actual interface{}) {
	var baselines = map[string]map[int]interface{}{}
	pc, _, _, _ := runtime.Caller(1)
	pkgFunctionName := runtime.FuncForPC(pc).Name() // e.g. k8s-lx1036/k8s-ui/backend/tests/controllers.(*AuthSuite).TestNotificationSubscribe
	functionNames := strings.Split(pkgFunctionName, ".")
	signature := SnakeCase(functionNames[len(functionNames)-1]) // TestNotificationSubscribe

	path := getBaselineDataFile()
	_, ok := baselines[signature]
	if !ok {
		if _, err := os.Stat(path); os.IsExist(err) {
			baseline, _ := ioutil.ReadFile(path)
			fmt.Println(string(baseline))
		} else {
			baselines[signature] = map[int]interface{}{}
		}
	}

}

func (suite *AuthSuite) TestNotificationList() {
	request := httptest.NewRequest("GET", "/api/v1/notifications", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", Token))
	recorder := httptest.NewRecorder()
	suite.routers.ServeHTTP(recorder, request)
	response := recorder.Result()
	body, _ := ioutil.ReadAll(response.Body)

	fmt.Println(string(body))
}

func TestAuthSuite(test *testing.T) {
	suite.Run(test, new(AuthSuite))
}

func getBaselineDataFile() string {
	pc, file, _, _ := runtime.Caller(1)
	baselineDir := filepath.Dir(file)
	pkgFunctionName := runtime.FuncForPC(pc).Name() // e.g. k8s-lx1036/k8s-ui/backend/tests/controllers.(*AuthSuite).TestNotificationSubscribe
	functionNames := strings.Split(pkgFunctionName, ".")
	functionName := strings.TrimPrefix(functionNames[len(functionNames)-1], "Test")
	baselineDir = getBaselinePath(baselineDir)
	if _, err := os.Stat(baselineDir); os.IsNotExist(err) {
		err = os.MkdirAll(baselineDir, 0755)
		if err != nil {
			panic(err)
		}
	}

	return fmt.Sprintf("%s/%s.json", baselineDir, functionName)
}

func SnakeCase(str string) string {
	snake := regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}

const DataSet = "simple"

func getBaselinePath(path string) string {
	return fmt.Sprintf("%s/%s/%s", path, "_baseline", DataSet)
}

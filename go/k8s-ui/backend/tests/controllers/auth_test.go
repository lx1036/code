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
	"strings"
	"testing"
)

type AuthSuite struct {
	suite.Suite
	//Token   string
	routers *gin.Engine
}

func (suite *AuthSuite) SetupTest() {
	//initial.InitDb()
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

	var notificationLogs struct{
		Data []models.NotificationLog `json:"data"`
	}

	_ = json.Unmarshal(body, &notificationLogs)

	var slices []interface{}
	slices = append(slices, response.StatusCode)
	slices = append(slices, response.Header)
	slices = append(slices, notificationLogs.Data)

	body2, _ := json.Marshal(slices)

	path, err := filepath.Abs("./baseline/auth.json")
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {

		var buffer = new(bytes.Buffer)
		err = json.Indent(buffer, body2, "", "  ")
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(path, buffer.Bytes(), 0666)
		if err != nil {
			panic(err)
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

func getBaselineDataFile(function string) string  {
	baselineDir := ""

	if _, err := os.Stat(baselineDir); os.IsNotExist(err) {
		err = os.MkdirAll(baselineDir, 0755)
		if err != nil {
			panic(err)
		}
	}

	return fmt.Sprintf("%s/%s.json", baselineDir, function)
}

func Assert(any interface{}) {
	//fmt.Println("StatusCode", response.StatusCode)
	//
	//body2, _ := ioutil.ReadAll(response.Body)
	//
	//
	header := http.Header{}
	header.Add("Auth", "123")

	//slices := make([]interface{})

	var slices []interface{}
	//slices[0] = struct {
	//	Code int `json:"code"`
	//}{
	//	Code: 200,
	//}
	//slices[1] = struct {
	//	Header http.Header `json:"header"`
	//}{
	//	Header: header,
	//}


	slices = append(slices, 200)
	slices = append(slices, header)

	var body2 struct{
		Name string `json:"name"`
	}
	body2.Name = "lx1036"
	slices = append(slices, body2)

	//slices[2] = struct {
	//	Content string `json:"content"`
	//}{
	//	Content: string(body2),
	//}




	//signature := fmt.Sprintf("%s:%s", )

	//code := response.StatusCode
	//
	//var content struct{
	//	Code int `json:"code"`
	//	Header http.Header `json:"header"`
	//
	//}

	body, _ := json.Marshal(slices)

	path, err := filepath.Abs("./baseline/auth.json")
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		var buffer = new(bytes.Buffer)
		err = json.Indent(buffer, body, "", "  ")
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(path, buffer.Bytes(), 0666)
		if err != nil {
			panic(err)
		}
	}
}

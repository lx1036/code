package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	_ "k8s-lx1036/k8s-ui/backend/controllers/auth"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	routers_gin "k8s-lx1036/k8s-ui/backend/routers-gin"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type AuthSuite struct {
	suite.Suite
	Token string
	routers *gin.Engine
}

func (suite *AuthSuite) SetupTest() {
	//initial.InitDb()
	//suite.routers = routers_gin.SetupRouter()
}

func (suite *AuthSuite) TeardownTest() {

}

func (suite *AuthSuite) TestCors()  {
	routers := routers_gin.SetupRouter()
	data := url.Values{}
	data.Set("username", "admin")
	data.Set("password", "password")
	request := httptest.NewRequest("OPTIONS", "/login/db", strings.NewReader(data.Encode()))
	request.Header.Add("Access-Control-Request-Method", http.MethodPost)
	request.Header.Add("Access-Control-Request-Headers", "authorization")
	request.Header.Add("Access-Control-Request-Headers", "content-type")
	request.Header.Add("Origin", "http://localhost:4200")
	recorder := httptest.NewRecorder()
	routers.ServeHTTP(recorder, request)
	//response := recorder.Result()
	headers := recorder.Header()
	for key, value := range headers {
		fmt.Println(key, value)
	}
	var token struct{
		Data struct{
			Token string `json:"token"`
		} `json:"data"`
	}

	_ =json.Unmarshal(recorder.Body.Bytes(), &token)

	suite.Token = token.Data.Token

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	assert.Equal(suite.T(), "*", recorder.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(suite.T(), http.MethodPost, recorder.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(suite.T(), "Authorization", recorder.Header().Get("Access-Control-Allow-Headers"))
}

func (suite *AuthSuite) TestLogin() {
	routers := routers_gin.SetupRouter()
	data := url.Values{}
	data.Set("username", "admin")
	data.Set("password", "password")
	request := httptest.NewRequest("POST", "/login/db", strings.NewReader(data.Encode()))
	request.Header.Set("content-type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	routers.ServeHTTP(recorder, request)
	response := recorder.Result()
	body, _ := ioutil.ReadAll(response.Body)

	var token base.JsonResponse
	_ = json.Unmarshal(body, &token)

	fmt.Println(response.Header)
	fmt.Println(token)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func (suite *AuthSuite) TestCurrentUser()  {
	request := httptest.NewRequest("GET", "/me", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", suite.Token))
	recorder := httptest.NewRecorder()
	suite.routers.ServeHTTP(recorder, request)
	response := recorder.Result()
	body, _ := ioutil.ReadAll(response.Body)

	var me struct{
		Errno int `json:"errno"`
		Errmsg string `json:"errmsg"`
		Data struct{
			ID int `json:"ID"`
		} `json:"data"`
	}
	_ = json.Unmarshal(body, &me)

	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
	assert.Equal(suite.T(), 1, me.Data.ID)
}

func TestAuthSuite(test *testing.T) {
	suite.Run(test, new(AuthSuite))
}

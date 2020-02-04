package controllers

import (
	"encoding/json"
	"fmt"
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
}

func (suite *AuthSuite) SetupTest() {
	//initial.InitDb()
}

func (suite *AuthSuite) TeardownTest() {

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

	fmt.Println(string(body))
	var token base.JsonResponse
	_ = json.Unmarshal(body, &token)

	fmt.Println(response.Header)
	fmt.Println(token)
	assert.Equal(suite.T(), http.StatusOK, response.StatusCode)
}

func TestAuthSuite(test *testing.T) {
	suite.Run(test, new(AuthSuite))
}

package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s-lx1036/k8s-ui/backend/models"
	_ "k8s-lx1036/k8s-ui/backend/routers"
	"net/http"
	"net/http/httptest"
	"testing"
)

type AppSuite struct {
	suite.Suite
}

func (appSuite *AppSuite) SetupTest() {

}

func (appSuite *AppSuite) TeardownTest() {

}

func TestExampleTestSuite(test *testing.T) {
	suite.Run(test, new(AppSuite))
}

// https://github.com/goinggo/beego-mgo/blob/master/test/endpointTests/buoyEndpoints_test.go
func (appSuite *AppSuite) TestDemo(test *testing.T) {
	//beego.TestBeegoInit(beego.AppPath)
	request, _ := http.NewRequest("PUT", "/api/v1/namespaces/1/apps/1", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, request)
	var app struct {
		Data models.App `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &app)

	assert.Equal(test, 200, w.Code)
	assert.Equal(test, "admin", app.Data.User)
}

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/deployment"
	"k8s-lx1036/k8s-ui/dashboard/model"
	"k8s-lx1036/k8s-ui/dashboard/router"
	v1 "k8s.io/api/apps/v1"
	api "k8s.io/api/core/v1"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

type DeploymentSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *DeploymentSuite) SetupTest() {
	model.SetMode(model.TestMode)
	suite.router = router.SetupRouter()
}

func (suite *DeploymentSuite) TeardownTest() {

}

func (suite *DeploymentSuite) TestDeploymentNameValidity() {
	payload := deployment.AppNameValiditySpec{
		Name:      "nginx-demo",
		Namespace: "default",
	}
	requestBody, _ := json.Marshal(payload)
	request := httptest.NewRequest("POST", "/api/v1/appdeployment/validate/name", bytes.NewBuffer(requestBody))
	request.Header.Set("content-type", "application/json")
	writer := httptest.NewRecorder()
	suite.router.ServeHTTP(writer, request)
	response := writer.Result()
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	type Response struct {
		Errno  int    `json:"errno"`
		Errmsg string `json:"errmsg"`
		Data   struct {
			Valid bool `json:"valid"`
		} `json:"data"`
	}
	var deploymentResponse Response
	_ = json.Unmarshal(body, &deploymentResponse)

	// Assert
	var slices []interface{}
	slices = append(slices, response.StatusCode)
	slices = append(slices, response.Header)
	slices = append(slices, deploymentResponse)

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

		//assert.EqualValues(suite.T(), strings.ReplaceAll(strings.ReplaceAll(string(decoded[2]), "\n", ""), " ", ""), strings.ReplaceAll(strings.ReplaceAll(string(body), "\n", ""), " ", ""))

		var content Response
		_ = json.Unmarshal(decoded[2], &content)
		assert.EqualValues(suite.T(), content, deploymentResponse)
	}
}

func (suite *DeploymentSuite) TestDeployment() {
	payload := deployment.DeploymentSpec{
		Name:      "nginx-demo",
		Namespace: "test-namespace",
		Labels: []deployment.Label{
			{Key: "app", Value: "nginx-demo"},
		},
		Description:     "create nginx deployment and service",
		Replicas:        3,
		ContainerImage:  "nginx:1.17.8",
		RunAsPrivileged: true,
		PortMappings: []deployment.PortMapping{
			{Port: 8088, TargetPort: 80, Protocol: "TCP"},
		},
		IsExternal: true,
	}
	requestBody, _ := json.Marshal(payload)
	request := httptest.NewRequest("POST", "/api/v1/appdeployment", bytes.NewBuffer(requestBody))
	request.Header.Set("content-type", "application/json")
	writer := httptest.NewRecorder()
	suite.router.ServeHTTP(writer, request)
	response := writer.Result()
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	type Response struct {
		Errno  int    `json:"errno"`
		Errmsg string `json:"errmsg"`
		Data   struct {
			Deployment *v1.Deployment `json:"deployment"`
			Service    *api.Service   `json:"service"`
		} `json:"data"`
	}
	var deploymentResponse Response
	_ = json.Unmarshal(body, &deploymentResponse)

	// Assert
	var slices []interface{}
	slices = append(slices, response.StatusCode)
	slices = append(slices, response.Header)
	slices = append(slices, deploymentResponse)

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

		//assert.EqualValues(suite.T(), strings.ReplaceAll(strings.ReplaceAll(string(decoded[2]), "\n", ""), " ", ""), strings.ReplaceAll(strings.ReplaceAll(string(body), "\n", ""), " ", ""))

		var content Response
		_ = json.Unmarshal(decoded[2], &content)
		assert.EqualValues(suite.T(), content, deploymentResponse)
	}
}

func TestDeploymentSuite(test *testing.T) {
	suite.Run(test, new(DeploymentSuite))
}

var rebase bool = false

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

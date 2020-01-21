package deployment

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/google/go-querystring/query"
)

const (
	K8sCdApi = "https://localhost:8080/openapi/v1/gateway/action/upgrade_deployment"
)

type UpgradeDeploymentResponse struct {
	Code   int      `json:"code"`
	Errors []string `json:"errors"`
}

type UpgradeDeploymentConfig struct {
	DeploymentName string `url:"deployment_name"`
	Namespace      string `url:"namespace"`
	Cluster        string `url:"cluster"`
	TemplateId     int    `url:"template_id"` // 32 bits
	Publish        bool   `url:"publish"`
	Description    string `url:"description"`
	Images         string `url:"images"`
	Environments   string `url:"environments"`
}

func UpgradeDeployment() {
	queryStr, _ := query.Values(UpgradeDeploymentConfig{
		DeploymentName: "",
		Namespace:      "",
		Cluster:        "",
		TemplateId:     0,
		Publish:        false,
		Description:    "",
		Images:         "",
		Environments:   "",
	})
	api := fmt.Sprintf("%s?%s", K8sCdApi, queryStr.Encode())
	response, err := httplib.Get(api).String()
	if err != nil {

	}

	var resp UpgradeDeploymentResponse
	if err := json.Unmarshal([]byte(response), &resp); err != nil {

	}
}

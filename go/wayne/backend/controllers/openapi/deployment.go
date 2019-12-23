package openapi

import (
	"encoding/json"
	"fmt"
	"k8s-lx1036/wayne/backend/client"
	"k8s-lx1036/wayne/backend/models"
	resdeployment "k8s-lx1036/wayne/backend/resources/deployment"
	"k8s.io/api/apps/v1beta1"
	"strings"
	"time"
)

// swagger:parameters UpgradeDeploymentParam
type UpgradeDeploymentParam struct {
	// in: query
	// Required: true
	Deployment string `json:"deployment"`
	// Required: true
	Namespace string `json:"namespace"`
	// 支持同时填写多个 Cluster，只需要在 cluster 之间使用英文半角的逗号分隔即可
	// Required: true
	Cluster  string `json:"cluster"`
	clusters []string
	// Required: false
	TemplateId int `json:"template_id"`
	// 该字段为 true 的时候，会自动使用新生成的配置模板上线，否则会只创建对应的模板，并且将模板 ID 返回（用于敏感的需要手动操作的上线环境）
	// Required: false
	Publish bool `json:"publish"`
	// 升级的描述
	// Required: false
	Description string `json:"description"`
	// 该字段为扁平化为字符串的 key-value 字典，填写格式为 容器名1=镜像名1,容器名2=镜像名2 (即:多个容器之间使用英文半角的逗号分隔）
	// Required: false
	Images   string `json:"images"`
	imageMap map[string]string
	// 该字段为扁平化为字符串的 key-value 字典，填写格式为 环境变量1=值1,环境变量2=值2 (即:多个环境变量之间使用英文半角的逗号分隔）
	// Required: false
	Environments string `json:"environments"`
	envMap       map[string]string
}

// swagger:route GET /upgrade_deployment deploy UpgradeDeploymentParam
//
// 用于 CI/CD 中的集成升级部署
//
// 该接口只能使用 app 级别的 apikey，这样做的目的主要是防止 apikey 的滥用。
// 目前用户可以选择两种用法，第一种是默认的，会根据请求的 images 和 environments 对特定部署线上模板进行修改并创建新模板，然后使用新模板进行升级；
// 需要说明的是，environments 列表会对 deployment 内所有容器中包含指定环境变量 key 的环境变量进行更新，如不包含，则不更新。
// 第二种是通过指定 publish=false 来关掉直接上线，这种条件下会根据 images 和 environments 字段创建新的模板，并返回新模板id，用户可以选择去平台上手动上线或者通过本接口指定template_id参数上线。
// cluster 字段可以选择单个机房也可以选择多个机房，对于创建模板并上线的用法，会根据指定的机房之前的模板进行分类（如果机房 a 和机房 b 使用同一个模板，那么调用以后仍然共用一个新模板）
// 而对于指定 template_id 来上线的形式，则会忽略掉所有检查，直接使用特定模板上线到所有机房。
//
//     Responses:
//       200: responseSuccess
//       400: responseState
//       401: responseState
//       500: responseState
// @router /upgrade_deployment [get]
func (controller *OpenAPIController)UpgradeDeployment()  {
	param := UpgradeDeploymentParam{
		Deployment:   controller.GetString("deployment"),
		Namespace:    controller.GetString("namespace"),
		Cluster:      controller.GetString("cluster"),
		clusters:     nil,
		TemplateId:   0,
		Publish:      false,
		Description:  controller.GetString("description"),
		Images:       controller.GetString("images"),
		imageMap:     nil,
		Environments: controller.GetString("environments"),
		envMap:       nil,
	}
	
	if !controller.CheckoutRoutePermission(UpgradeDeploymentAction) || !controller.CheckDeploymentPermission(param.Deployment) || !controller.CheckNamespacePermission(param.Namespace) {
		return
	}
	
	param.clusters = strings.Split(param.Cluster, ",")
	var err error
	param.Publish, err = controller.GetBool("publish", true)
	if err != nil {
	
	}
	param.TemplateId, err = controller.GetInt("template_id", 0)
	if err != nil {
	
	}
	if param.TemplateId != 0 && param.Publish {
	
	}
	
	images := strings.Split(param.Images, ",")
	for _, image := range images {
		arr := strings.Split(image, "=")
		if len(arr) == 2 && arr[1] != "" {
			param.imageMap[arr[0]] = arr[1]
		}
	}
	
	deployInfoMap := make(map[int64]([]*DeploymentInfo))
	for _, cluster := range param.clusters {
		deployInfo, err := getOnlineDeploymenetInfo(param.Deployment, param.Namespace, cluster, 0)
		if err != nil {
		
		}
		
		tmplId := deployInfo.DeploymentTemplete.Id
		
		deployInfoMap[tmplId] = append(deployInfoMap[tmplId], deployInfo)
	}
	
	
	for _, deployInfos := range deployInfoMap {
		deployInfo := deployInfos[0]
		newTpl, err := json.Marshal(deployInfo.DeploymentObject)
		if err != nil {
			
			continue
		}
		deployInfo.DeploymentTemplete.Template = string(newTpl)
		//更新deploymentTpl中的CreateTime和UpdateTime,数据库中不会自动更新
		deployInfo.DeploymentTemplete.CreateTime = time.Now()
		deployInfo.DeploymentTemplete.UpdateTime = time.Now()
		
		for _, _ = range deployInfos {
		
		}
	}
	
	for _, deployInfos := range deployInfoMap {
		for _, deployInfo := range deployInfos {
			err := publishDeployment(deployInfo, controller.APIKey.String())
			if err != nil {
			
			}
		}
	}
	
	controller.HandleResponse(nil)
}

type DeploymentInfo struct {
	Deployment         *models.Deployment
	DeploymentTemplete *models.DeploymentTemplate
	DeploymentObject   *v1beta1.Deployment
	Cluster            *models.Cluster
	Namespace          *models.Namespace
}

// 主要用于从数据库中查找、拼凑出用于更新的模板资源，资源主要用于 k8s 数据更新和 数据库存储更新记录等
func getOnlineDeploymenetInfo(deployment, namespace, cluster string, templateId int64) (deployInfo *DeploymentInfo, err error) {
	deployInfo = new(DeploymentInfo)
	
	return deployInfo, nil
	
}

// 通过给定模板资源把业务发布到k8s集群中，并在数据库中更新发布记录
func publishDeployment(deployInfo *DeploymentInfo, username string) error {
	// 操作 kubernetes api，实现升级部署
	cli, err := client.Client(deployInfo.Cluster.Name)
	if err == nil {
		
		_, err = resdeployment.CreateOrUpdateDeployment(cli, deployInfo.DeploymentObject)
		
		return nil
	} else {
		return fmt.Errorf("Failed to get k8s client(cluster: %s): %v", deployInfo.Cluster.Name, err)
	}
}

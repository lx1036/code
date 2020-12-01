package openapi

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/validation"
	"k8s-lx1036/k8s-ui/backend/controllers/common"
	"k8s-lx1036/k8s-ui/backend/models"
	"k8s.io/api/apps/v1beta1"
	"strings"
)

// swagger:parameters UpgradeDeploymentParam
type UpgradeDeploymentParam struct {
	// in: query
	// Required: true
	DeploymentName string `valid:"Required"`
	// Required: true
	Namespace string `valid:"Required"`
	// 支持同时填写多个 Cluster，只需要在 cluster 之间使用英文半角的逗号分隔即可
	// Required: true
	Cluster  string `valid:"Required"`
	clusters []string
	// Required: false
	TemplateId int
	// 该字段为 true 的时候，会自动使用新生成的配置模板上线，否则会只创建对应的模板，并且将模板 ID 返回（用于敏感的需要手动操作的上线环境）
	// Required: false
	Publish bool
	// 升级的描述
	// Required: false
	Description string
	// 该字段为扁平化为字符串的 key-value 字典，填写格式为 容器名1=镜像名1,容器名2=镜像名2 (即:多个容器之间使用英文半角的逗号分隔）
	// Required: false
	Images   string
	imageMap map[string]string
	// 该字段为扁平化为字符串的 key-value 字典，填写格式为 环境变量1=值1,环境变量2=值2 (即:多个环境变量之间使用英文半角的逗号分隔）
	// Required: false
	Environments string
	envMap       map[string]string
}

type DeploymentInfo struct {
	Deployment         *models.Deployment
	DeploymentTemplate *models.DeploymentTemplate
	DeploymentObject   *v1beta1.Deployment
	Cluster            *models.Cluster
	Namespace          *models.Namespace
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
func (controller *OpenAPIController) UpgradeDeployment() {
	var err error
	param := UpgradeDeploymentParam{
		DeploymentName: controller.GetString("deployment_name"), // deployment name, e.g. "fanyi-so-com-stage"
		Namespace:      controller.GetString("namespace"),
		Cluster:        controller.GetString("cluster"), // "SHBT,ZZZC"
		Description:    controller.GetString("description"),
		Images:         controller.GetString("images"),
		Environments:   controller.GetString("environments"),
	}
	valid := validation.Validation{}
	b, err := valid.Valid(param)
	if err != nil { // Validate request params
		controller.HandleResponse(nil)
		return
	}
	if !b {
		controller.HandleResponse(nil)
		return
	}

	if !controller.CheckoutRoutePermission(UpgradeDeploymentAction) || !controller.CheckDeploymentPermission(param.DeploymentName) || !controller.CheckNamespacePermission(param.Namespace) {
		return
	}

	param.clusters = strings.Split(param.Cluster, ",")
	param.Publish, err = controller.GetBool("publish", true)
	if err != nil {

	}
	param.TemplateId, err = controller.GetInt("template_id", 0)
	if err != nil {

	}
	// publish 表示立即发布 deployment 模板
	if param.TemplateId != 0 && param.Publish {
		for _, cluster := range param.clusters {
			deploymentInfo, err := getOnlineDeploymenetInfo(param.DeploymentName, param.Namespace, cluster, int64(param.TemplateId))
			if err != nil {

				continue
			}

			common.DeploymentPreDeploy(deploymentInfo.DeploymentObject, deploymentInfo.Deployment, deploymentInfo.Cluster, deploymentInfo.Namespace)
			err = publishDeployment(deploymentInfo, controller.APIKey.String())
			if err != nil {

			}
		}

		controller.HandleResponse(nil)
		return
	}

	// merge images
	// "golang=aliyun.cloud/golang:1.0.0,openresty=aliyun.cloud/openresty:1.0.0"
	images := strings.Split(param.Images, ",")
	for _, image := range images {
		items := strings.Split(image, "=") // golang=aliyun.cloud/golang
		if len(items) == 2 && items[1] != "" {
			param.imageMap[items[0]] = items[1]
		}
	}

	// merge envMap
	if param.Environments != "" { // environments e.g. "key1=value1,key2=value2"
		environments := strings.Split(param.Environments, ",")
		for _, environment := range environments {
			items := strings.Split(environment, "=")
			if len(items) == 2 && items[1] != "" {
				param.envMap[items[0]] = items[1]
			}
		}
	}

	deploymentInfoMap := make(map[int64][]*DeploymentInfo)
	for _, cluster := range param.clusters {
		deployInfo, err := getOnlineDeploymenetInfo(param.DeploymentName, param.Namespace, cluster, 0)
		if err != nil {

		}

		// 率先把强制指定的环境变量，如和系统环境变量冲突，后面会覆盖
		for key, container := range deployInfo.DeploymentObject.Spec.Template.Spec.Containers {
			for index, value := range container.Env {
				if param.envMap[value.Name] != "" {
					deployInfo.DeploymentObject.Spec.Template.Spec.Containers[key].Env[index].Value = param.envMap[value.Name]
				}
			}
		}

		common.DeploymentPreDeploy(deployInfo.DeploymentObject, deployInfo.Deployment, deployInfo.Cluster, deployInfo.Namespace)

		//tmplId := deployInfo.DeploymentTemplate.ID
		//deployInfo.DeploymentTemplate.Id = 0
		//deployInfo.DeploymentTemplate.User = controller.APIKey.String()
		deployInfo.DeploymentTemplate.Description = "[APIKey] " + controller.GetString("description")

		for k, v := range deployInfo.DeploymentObject.Spec.Template.Spec.Containers {
			if param.imageMap[v.Name] != "" {
				deployInfo.DeploymentObject.Spec.Template.Spec.Containers[k].Image = param.imageMap[v.Name]
			}
		}

		//deploymentInfoMap[tmplId] = append(deploymentInfoMap[tmplId], deployInfo)
	}

	//for id, deploymentInfos := range deploymentInfoMap {
	//	deploymentInfo := deploymentInfos[0]
	//	newTpl, err := json.Marshal(deploymentInfo.DeploymentObject)
	//	if err != nil {
	//
	//		continue
	//	}
	//	deploymentInfo.DeploymentTemplate.Template = string(newTpl)
	//	//更新deploymentTpl中的CreateTime和UpdateTime,数据库中不会自动更新
	//	//deploymentInfo.DeploymentTemplate.CreateTime = time.Now()
	//	//deploymentInfo.DeploymentTemplate.UpdateTime = time.Now()
	//	//newTplId, err := models.DeploymentTplModel.Add(deploymentInfo.DeploymentTemplate)
	//	//if err != nil {
	//	//	continue
	//	//}
	//	//for k, info := range deploymentInfos {
	//	//	err := models.DeploymentModel.UpdateById(info.Deployment)
	//	//	if err != nil {
	//	//		continue
	//	//	}
	//	//	//deploymentInfoMap[id][k].DeploymentTemplate.Id = newTplId
	//	//}
	//}

	for _, deployInfos := range deploymentInfoMap {
		for _, deployInfo := range deployInfos {
			err := publishDeployment(deployInfo, controller.APIKey.String())
			if err != nil {

			}
		}
	}

	controller.HandleResponse(nil)
}

// 主要用于从数据库中查找、拼凑出用于更新的模板资源，资源主要用于 k8s 数据更新和数据库存储更新记录等
func getOnlineDeploymenetInfo(name, namespace, cluster string, templateId int64) (deployInfo *DeploymentInfo, err error) {
	deployment, err := models.DeploymentModel.GetByName(name)
	if err != nil {

	}

	deployInfo = &DeploymentInfo{}

	if templateId != 0 {

	} else {
		// 从 mysql 中获取线上模板数据
		//status, err := models.PublishStatusModel.GetByCluster(models.PublishTypeDeployment, deployment.Id, cluster)
		//if err != nil {
		//
		//}
		//deployInfo.DeploymentTemplate, err = models.DeploymentTplModel.GetById(status.TemplateId)
		//if err != nil {
		//
		//}
	}

	deployObj := v1beta1.Deployment{}
	err = json.Unmarshal(hack.Slice(deployInfo.DeploymentTemplate.Template), &deployObj)
	if err != nil {

	}

	//app, _ := models.AppModel.GetById(deployment.AppId)
	//err = json.Unmarshal([]byte(app.Namespace.MetaData), &app.Namespace.MetaDataObj)
	//if err != nil {
	//
	//}

	//deployObj.Namespace = app.Namespace.KubeNamespace
	//err = json.Unmarshal([]byte(deployment.MetaData), &deployment.MetaDataObj)
	//if err != nil {
	//
	//}
	//
	//rp := deployment.MetaDataObj.Replicas[cluster]
	//deployObj.Spec.Replicas = &rp
	deployInfo.DeploymentObject = &deployObj
	deployInfo.Deployment = deployment
	//deployInfo.Namespace = app.Namespace

	deployInfo.Cluster, err = models.ClusterModel.GetParsedMetaDataByName(cluster)
	if err != nil {

	}

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

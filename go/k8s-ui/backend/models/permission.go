package models

import (
	"errors"
	"strings"
	"time"
)

const (
	TableNamePermission = "permission"

	PermissionCreate = "CREATE"
	PermissionUpdate = "UPDATE"
	PermissionRead   = "READ"
	PermissionDelete = "DELETE"

	PermissionTypeCronjob = "CRONJOB"
	PermissionBlank       = "_"
)

type Permission struct {
	Id      int64  `gorm:"auto" json:"id,omitempty"`
	Name    string `gorm:"index;size(200)" json:"name,omitempty"`
	Comment string `gorm:"type(text)" json:"comment,omitempty"`

	CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`

	Groups []*Group `gorm:"reverse(many)" json:"groups,omitempty"`
}

type ActionPermission struct {
	PermissionRead   bool `json:"read" mapstructure:"READ"`
	PermissionCreate bool `json:"create" mapstructure:"CREATE"`
	PermissionUpdate bool `json:"update" mapstructure:"UPDATE"`
	PermissionDelete bool `json:"delete" mapstructure:"DELETE"`
}

type TypePermission struct {
	PermissionTypeApp                   ActionPermission `json:"app" mapstructure:"APP"`
	PermissionTypeAppUser               ActionPermission `json:"appUser" mapstructure:"APPUSER"`
	PermissionTypeNamespace             ActionPermission `json:"namespace" mapstructure:"NAMESPACE"`
	PermissionTypeNamespaceUser         ActionPermission `json:"namespaceUser" mapstructure:"NAMESPACEUSER"`
	PermissionTypeDeployment            ActionPermission `json:"deployment" mapstructure:"DEPLOYMENT"`
	PermissionTypeSecret                ActionPermission `json:"secret" mapstructure:"SECRET"`
	PermissionTypeService               ActionPermission `json:"service" mapstructure:"SERVICE"`
	PermissionTypeConfigMap             ActionPermission `json:"configmap" mapstructure:"CONFIGMAP"`
	PermissionTypeCronjob               ActionPermission `json:"cronjob" mapstructure:"CRONJOB"`
	PermissionTypePersistentVolumeClaim ActionPermission `json:"pvc" mapstructure:"PVC"`
	PermissionTypeWebHook               ActionPermission `json:"webHook" mapstructure:"WEBHOOK"`
	PermissionTypeApiKey                ActionPermission `json:"apiKey" mapstructure:"APIKEY"`
	PermissionTypeStatefulset           ActionPermission `json:"statefulset" mapstructure:"STATEFULSET"`
	PermissionTypeDaemonSet             ActionPermission `json:"daemonSet" mapstructure:"DAEMONSET"`
	PermissionTypeBILL                  ActionPermission `json:"bill" mapstructure:"BILL"`
	PermissionIngress                   ActionPermission `json:"ingress" mapstructure:"INGRESS"`
	PermissionHPA                       ActionPermission `json:"hpa" mapstructure:"HPA"`

	// Kubernetes resource permission
	PermissionTypeKubeConfigMap               ActionPermission `json:"kubeConfigMap" mapstructure:"KUBECONFIGMAP"`
	PermissionTypeKubeDaemonSet               ActionPermission `json:"kubeDaemonSet" mapstructure:"KUBEDAEMONSET"`
	PermissionTypeKubeDeployment              ActionPermission `json:"kubeDeployment" mapstructure:"KUBEDEPLOYMENT"`
	PermissionTypeKubeEvent                   ActionPermission `json:"kubeEvent" mapstructure:"KUBEEVENT"`
	PermissionTypeKubeHorizontalPodAutoscaler ActionPermission `json:"kubeHorizontalPodAutoscaler" mapstructure:"KUBEHORIZONTALPODAUTOSCALER"`
	PermissionTypeKubeIngress                 ActionPermission `json:"kubeIngress" mapstructure:"KUBEINGRESS"`
	PermissionTypeKubeJob                     ActionPermission `json:"kubeJob" mapstructure:"KUBEJOB"`
	PermissionTypeKubeCronJob                 ActionPermission `json:"kubeCronJob" mapstructure:"KUBECRONJOB"`
	PermissionTypeKubeNamespace               ActionPermission `json:"kubeNamespace" mapstructure:"KUBENAMESPACE"`
	PermissionTypeKubeNode                    ActionPermission `json:"kubeNode" mapstructure:"KUBENODE"`
	PermissionTypeKubePersistentVolumeClaim   ActionPermission `json:"kubePersistentVolumeClaim" mapstructure:"KUBEPERSISTENTVOLUMECLAIM"`
	PermissionTypeKubePersistentVolume        ActionPermission `json:"kubePersistentVolume" mapstructure:"KUBEPERSISTENTVOLUME"`
	PermissionTypeKubePod                     ActionPermission `json:"kubePod" mapstructure:"KUBEPOD"`
	PermissionTypeKubeReplicaSet              ActionPermission `json:"kubeReplicaSet" mapstructure:"KUBEREPLICASET"`
	PermissionTypeKubeSecret                  ActionPermission `json:"kubeSecret" mapstructure:"KUBESECRET"`
	PermissionTypeKubeService                 ActionPermission `json:"kubeService" mapstructure:"KUBESERVICE"`
	PermissionTypeKubeStatefulSet             ActionPermission `json:"kubeStatefulSet" mapstructure:"KUBESTATEFULSET"`
	PermissionTypeKubeEndpoint                ActionPermission `json:"kubeEndpoints" mapstructure:"KUBEENDPOINTS"`
	PermissionTypeKubeStorageClass            ActionPermission `json:"kubeStorageClass" mapstructure:"KUBESTORAGECLASS"`
	PermissionTypeKubeRole                    ActionPermission `json:"kubeRole" mapstructure:"KUBEROLE"`
	PermissionTypeKubeRoleBinding             ActionPermission `json:"kubeRoleBinding" mapstructure:"KUBEROLEBINDING"`
	PermissionTypeKubeClusterRole             ActionPermission `json:"kubeClusterRole" mapstructure:"KUBECLUSTERROLE"`
	PermissionTypeKubeClusterRoleBinding      ActionPermission `json:"kubeClusterRoleBinding" mapstructure:"KUBECLUSTERROLEBINDING"`
	PermissionTypeKubeServiceAccount          ActionPermission `json:"kubeServiceAccount" mapstructure:"KUBESERVICEACCOUNT"`
}

type permissionModel struct{}

/*
 * 合并permission的type和action
 */
func (*permissionModel) MergeName(perType string, perAction string) (perName string) {
	perName = perType + PermissionBlank + perAction
	return perName
}

func (*permissionModel) SplitName(name string) (paction string, ptype string, err error) {
	stringSlice := strings.Split(name, PermissionBlank)
	if len(stringSlice) < 2 {
		err = errors.New("Permission name split fail")
		return "", "", err
	}
	ptype = stringSlice[0]
	paction = stringSlice[1]
	return
}

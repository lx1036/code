package models

import (
	"github.com/astaxie/beego/orm"
	"sync"
)

var (
	globalOrm orm.Ormer
	once      sync.Once

	PermissionModel    *permissionModel
	CronjobModel       *cronjobModel
	UserModel          *userModel
	DeploymentModel    *deploymentModel
	PublishStatusModel *publishStatusModel
	DeploymentTplModel *deploymentTplModel
	AppModel           *appModel
	ClusterModel       *clusterModel
	NamespaceModel     *namespaceModel
)

func init() {
	orm.RegisterModel(
		new(User),
		new(Cluster),
		new(User),
		new(App),
		new(AppStarred),
		new(AppUser),
		new(NamespaceUser),
		new(Cluster),
		new(Namespace),
		new(Deployment),
		new(DeploymentTemplate),
		new(Group),
		new(Permission),
		new(Cronjob),
		new(PublishStatus),
		new(APIKey),
	)

	// init models
	UserModel = &userModel{}
	AppModel = &appModel{}
	ClusterModel = &clusterModel{}
	NamespaceModel = &namespaceModel{}
	DeploymentModel = &deploymentModel{}
	DeploymentTplModel = &deploymentTplModel{}
	CronjobModel = &cronjobModel{}
	PublishStatusModel = &publishStatusModel{}
}

// singleton init ormer ,only use for normal db operation
// if you begin transaction，please use orm.NewOrm()
func Ormer() orm.Ormer {
	once.Do(func() {
		globalOrm = orm.NewOrm()
	})
	return globalOrm
}

package models

import (
	"github.com/astaxie/beego/orm"
	"sync"
)

var (
	globalOrm orm.Ormer
	once      sync.Once

	ApiKeyModel                   *apiKeyModel
	AppUserModel                  *appUserModel
	AppStarredModel               *appStarredModel
	AppModel                      *appModel
	AuditLogModel                 *auditLogModel
	ChargeModel                   *chargeModel
	ClusterModel                  *clusterModel
	ConfigModel                   *configModel
	ConfigMapModel                *configMapModel
	ConfigMapTplModel             *configMapTplModel
	CronjobModel                  *cronjobModel
	CronjobTplModel               *cronjobTplModel
	CustomLinkModel               *customLinkModel
	DaemonSetModel                *daemonSetModel
	DaemonSetTplModel             *daemonSetTplModel
	DeploymentModel               *deploymentModel
	DeploymentTplModel            *deploymentTplModel
	GroupModel                    *groupModel
	HPAModel                      *hpaModel
	HPATemplateModel              *hpaTemplateModel
	IngressModel                  *ingressModel
	IngressTemplateModel          *ingressTemplateModel
	InvoiceModel                  *invoiceModel
	LinkTypeModel                 *linkTypeModel
	NamespaceModel                *namespaceModel
	NamespaceUserModel            *namespaceUserModel
	NotificationModel             *notificationModel
	NotificationLogModel          *notificationLogModel
	PermissionModel               *permissionModel
	PersistentVolumeClaimModel    *persistentVolumeClaimModel
	PersistentVolumeClaimTplModel *persistentVolumeClaimTplModel
	PublishStatusModel            *publishStatusModel
	PublishHistoryModel           *publishHistoryModel
	SecretModel                   *secretModel
	SecretTplModel                *secretTplModel
	StatefulsetModel              *statefulsetModel
	StatefulsetTplModel           *statefulsetTplModel
	UserModel                     *userModel
	WebHookModel                  *webHookModel
)

func init() {
	/*orm.RegisterModel( // 41 tables
		new(App),
		new(APIKey),
		new(AppStarred),
		new(AppUser),
		new(AuditLog),
		new(Cluster),
		new(Charge),
		new(Config),
		new(ConfigMap),
		new(ConfigMapTemplate),
		new(Cronjob),
		new(CronjobTemplate),
		new(CustomLink),
		new(DaemonSet),
		new(DaemonSetTemplate),
		new(Deployment),
		new(DeploymentTemplate),
		new(Group),
		new(HPA),
		new(HPATemplate),
		new(Ingress),
		new(IngressTemplate),
		new(Invoice),
		new(LinkType),
		new(Namespace),
		new(NamespaceUser),
		new(Notification),
		new(NotificationLog),
		new(Permission),
		new(PersistentVolumeClaim),
		new(PersistentVolumeClaimTemplate),
		new(PublishHistory),
		new(PublishStatus),
		new(Secret),
		new(SecretTemplate),
		new(Service),
		new(ServiceTemplate),
		new(Statefulset),
		new(StatefulsetTemplate),
		new(User),
		new(WebHook),
	)*/

	// init models
	ApiKeyModel = &apiKeyModel{}
	AppModel = &appModel{}
	AppUserModel = &appUserModel{}
	AppStarredModel = &appStarredModel{}
	AuditLogModel = &auditLogModel{}
	ChargeModel = &chargeModel{}
	ClusterModel = &clusterModel{}
	ConfigMapModel = &configMapModel{}
	ConfigMapTplModel = &configMapTplModel{}
	ConfigModel = &configModel{}
	CronjobModel = &cronjobModel{}
	CronjobTplModel = &cronjobTplModel{}
	CustomLinkModel = &customLinkModel{}
	DaemonSetModel = &daemonSetModel{}
	DaemonSetTplModel = &daemonSetTplModel{}
	DeploymentModel = &deploymentModel{}
	DeploymentTplModel = &deploymentTplModel{}
	GroupModel = &groupModel{}
	HPAModel = &hpaModel{}
	HPATemplateModel = &hpaTemplateModel{}
	IngressModel = &ingressModel{}
	IngressTemplateModel = &ingressTemplateModel{}
	InvoiceModel = &invoiceModel{}
	LinkTypeModel = &linkTypeModel{}
	NamespaceModel = &namespaceModel{}
	NamespaceUserModel = &namespaceUserModel{}
	NotificationModel = &notificationModel{}
	NotificationLogModel = &notificationLogModel{}
	PermissionModel = &permissionModel{}
	PersistentVolumeClaimModel = &persistentVolumeClaimModel{}
	PersistentVolumeClaimTplModel = &persistentVolumeClaimTplModel{}
	PublishHistoryModel = &publishHistoryModel{}
	PublishStatusModel = &publishStatusModel{}
	SecretModel = &secretModel{}
	SecretTplModel = &secretTplModel{}
	StatefulsetModel = &statefulsetModel{}
	StatefulsetTplModel = &statefulsetTplModel{}
	UserModel = &userModel{}
	WebHookModel = &webHookModel{}
}

// singleton init ormer ,only use for normal db operation
// if you begin transaction，please use orm.NewOrm()
func Ormer() orm.Ormer {
	once.Do(func() {
		globalOrm = orm.NewOrm()
	})
	return globalOrm
}

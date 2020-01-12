package models

const (
	PublishTypeDeployment PublishType = iota
)

type PublishType int32

// 记录已发布模版信息
type PublishStatus struct {
	Id         int64       `orm:"auto" json:"id,omitempty"`
	Type       PublishType `orm:"index;type(integer)" json:"type,omitempty"`
	ResourceId int64       `orm:"index;column(resource_id)" json:"resourceId,omitempty"`
	TemplateId int64       `orm:"index;column(template_id);" json:"templateId,omitempty"`
	Cluster    string      `orm:"size(128);column(cluster)" json:"cluster,omitempty"`
}

type publishStatusModel struct{}

func (*publishStatusModel) GetByCluster(publishType PublishType, resourceId int64, cluster string) (publishStatus PublishStatus, err error) {
	err = Ormer().
		QueryTable(new(PublishStatus)).
		Filter("ResourceId", resourceId).
		Filter("Type", publishType).
		Filter("Cluster", cluster).
		One(&publishStatus)
	return
}

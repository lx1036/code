package models

import "time"

type DeploymentTemplate struct {
	ID           uint      `gorm:"column:id;primary_key;"`
	Name         string    `gorm:"column:name;size:128;not null;default:'';"`
	Template     string    `gorm:"column:template;type:longtext;not null;"`
	DeploymentId uint      `gorm:"column:deployment_id"`
	MetaData     string    `gorm:"column:meta_data;type:longtext;not null;"`
	Description  string    `gorm:"column:description;size:512;not null;default:'';"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt    time.Time `gorm:"column:deleted_at;default:null;"`

	//Id          int64       `gorm:"auto" json:"id,omitempty"`
	//Name        string      `gorm:"size(128)" json:"name,omitempty"`
	//Template    string      `gorm:"type(text)" json:"template,omitempty"`
	//Deployment  *Deployment `gorm:"index;rel(fk);column(deployment_id)" json:"deployment,omitempty"`
	//Description string      `gorm:"size(512)" json:"description,omitempty"`
	//
	//// TODO
	//// 如果使用指针类型auto_now_add和auto_now可以自动生效,但是orm QueryRows无法对指针类型的time正常赋值，
	//// 不使用指针类型创建时需要手动把创建时间设置为当前时间,更新时也需要处理创建时间
	//CreateTime time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	//UpdateTime time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	//User       string    `gorm:"size(128)" json:"user,omitempty"`
	//Deleted    bool      `gorm:"default(false)" json:"deleted,omitempty"`
	//
	//DeploymentId int64            `gorm:"-" json:"deploymentId,omitempty"`
	//Status       []*PublishStatus `gorm:"-" json:"status,omitempty"`
}

type deploymentTplModel struct{}

func (*deploymentTplModel) GetById(id int64) (deploymentTpl *DeploymentTemplate, err error) {
	//deploymentTpl = &DeploymentTemplate{Id: id}
	//if err = Ormer().Read(deploymentTpl); err == nil {
	//	_, err = Ormer().LoadRelated(deploymentTpl, "Deployment")
	//	if err == nil {
	//		deploymentTpl.DeploymentId = deploymentTpl.Deployment.Id
	//		return deploymentTpl, nil
	//	}
	//}

	return nil, err
}

func (*deploymentTplModel) Add(template *DeploymentTemplate) (id int64, err error) {
	//template.Deployment = &Deployment{Id: template.DeploymentId}
	//now := time.Now()
	//template.CreateTime = now
	//template.UpdateTime = now
	//id, err = Ormer().Insert(template)
	return
}

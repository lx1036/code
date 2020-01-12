package models

import "time"

type DeploymentTemplate struct {
	Id          int64       `orm:"auto" json:"id,omitempty"`
	Name        string      `orm:"size(128)" json:"name,omitempty"`
	Template    string      `orm:"type(text)" json:"template,omitempty"`
	Deployment  *Deployment `orm:"index;rel(fk);column(deployment_id)" json:"deployment,omitempty"`
	Description string      `orm:"size(512)" json:"description,omitempty"`

	// TODO
	// 如果使用指针类型auto_now_add和auto_now可以自动生效,但是orm QueryRows无法对指针类型的time正常赋值，
	// 不使用指针类型创建时需要手动把创建时间设置为当前时间,更新时也需要处理创建时间
	CreateTime time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `orm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `orm:"default(false)" json:"deleted,omitempty"`

	DeploymentId int64            `orm:"-" json:"deploymentId,omitempty"`
	Status       []*PublishStatus `orm:"-" json:"status,omitempty"`
}

type deploymentTplModel struct{}

func (*deploymentTplModel) GetById(id int64) (deploymentTpl *DeploymentTemplate, err error) {
	deploymentTpl = &DeploymentTemplate{Id: id}
	if err = Ormer().Read(deploymentTpl); err == nil {
		_, err = Ormer().LoadRelated(deploymentTpl, "Deployment")
		if err == nil {
			deploymentTpl.DeploymentId = deploymentTpl.Deployment.Id
			return deploymentTpl, nil
		}
	}

	return nil, err
}

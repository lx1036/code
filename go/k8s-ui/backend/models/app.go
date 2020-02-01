package models

import (
	"k8s-lx1036/k8s-ui/backend/common"
	"time"
)

type App struct {
	Id        int64      `gorm:"auto" json:"id,omitempty"`
	Name      string     `gorm:"index;size(128)" json:"name,omitempty"`
	Namespace *Namespace `gorm:"index;column(namespace_id);rel(fk)" json:"namespace"`
	/*
		{
		    "mode": "beta",
		    "system.api-name-generate-rule":"none" // refers to models.Config ConfigKeyApiNameGenerateRule
		}
	*/
	MetaData    string `gorm:"type(text)" json:"metaData,omitempty"`
	Description string `gorm:"null;size(512)" json:"description,omitempty"`

	CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string     `gorm:"size(128)" json:"user,omitempty"`
	Deleted    bool       `gorm:"default(false)" json:"deleted,omitempty"`

	// 用于权限的关联查询
	AppUsers []*AppUser `gorm:"reverse(many)" json:"-"`

	// 关注的关联查询
	AppStars []*AppStarred `gorm:"reverse(many)" json:"-"`
}

type AppStar struct {
	App

	CreateTime    time.Time `json:"createTime"`
	NamespaceId   int64     `json:"namespaceId"`
	NamespaceName string    `json:"namespaceName"`
	Starred       bool      `json:"starred"`
}

type AppStatistics struct {
	Total   int64              `json:"total,omitempty"`
	Details *[]NamespaceDetail `json:"details,omitempty"`
}
type NamespaceDetail struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type appModel struct{}

func (model *appModel) GetById(id int64) (v *App, err error) {
	v = &App{Id: id}

	if err = Ormer().Read(v); err != nil {
		return nil, err
	}
	_, err = Ormer().LoadRelated(v, "namespace")
	if err == nil {
		return v, nil
	}
	return nil, err
}

func (*appModel) UpdateById(m *App) (err error) {
	v := App{Id: m.Id}
	if err = Ormer().Read(&v); err == nil {
		_, err = Ormer().Update(m)
		return err
	}
	return
}

func (model *appModel) GetAppCountGroupByNamespace() (*[]NamespaceDetail, error) {
	sql := `SELECT namespace.name as name, count(*) as count FROM
			app inner join namespace on app.namespace_id=namespace.id
             group by app.namespace_id;`
	var details []NamespaceDetail
	_, err := Ormer().Raw(sql).QueryRows(&details)

	return &details, err
}

func (model *appModel) Count(param *common.QueryParam, b bool, i int64) (total int64, err error) {
	return 0, nil
}

func (model *appModel) List(param *common.QueryParam, b bool, i int64) (apps []AppStar, err error) {
	return nil, nil
}

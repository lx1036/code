package models

import (
	"k8s-lx1036/k8s-ui/backend/common"
	"time"
)

const (
	TableNameApp = "apps"
)

type App struct {
	ID          uint      `gorm:"column:id;primary_key;"`
	Name        string    `gorm:"column:name;size:128;not null;index:app_name;default:'';"`
	NamespaceID uint      `gorm:"column:namespace_id;"`
	MetaData    string    `gorm:"column:meta_data;type:longtext;not null;"`
	Description string    `gorm:"column:description;size:512;default:null;"`
	UserID      uint      `gorm:"column:user_id;"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt   time.Time `gorm:"column:deleted_at;default:null;"`

	User      User      `gorm:"foreignkey:UserID;association_foreignkey:ID;"`
	Namespace Namespace `gorm:"foreignkey:NamespaceID;association_foreignkey:ID;"`
	//Namespace *Namespace `gorm:"index;column(namespace_id);rel(fk)" json:"namespace"`
	/*
		{
		    "mode": "beta",
		    "system.api-name-generate-rule":"none" // refers to models.Config ConfigKeyApiNameGenerateRule
		}
	*/

	//CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	//UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	//User       string     `gorm:"size(128)" json:"user,omitempty"`
	//Deleted    bool       `gorm:"default(false)" json:"deleted,omitempty"`

	// 用于权限的关联查询
	//AppUsers []*AppUser `gorm:"reverse(many)" json:"-"`

	// 关注的关联查询
	//AppStars []*AppStarred `gorm:"reverse(many)" json:"-"`
}

func (App) TableName() string {
	return TableNameApp
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
	v = &App{ID: uint(id)}

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
	v := App{ID: uint(m.ID)}
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

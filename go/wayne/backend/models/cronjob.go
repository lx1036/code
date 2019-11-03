package models

import (
    kapi "k8s.io/api/core/v1"
    "time"
)

type cronjobModel struct{}

type CronjobMetaData struct {
    Replicas map[string]int32 `json:"replicas"`
    Suspends map[string]bool  `json:"suspends"`
    Affinity *kapi.Affinity   `json:"affinity,omitempty"`
    // 是否允许用户使用特权模式，默认不允许,key 为容器名称
    Privileged map[string]*bool `json:"privileged"`
}

type Cronjob struct {
    Id   int64  `orm:"auto" json:"id,omitempty"`
    Name string `orm:"unique;size(128)" json:"name,omitempty"`
    // 存储模版可上线机房，已挂起的机房
    /*
    	{
    	  "replicas": {
    	    "K8S": 1
    	  },
    	}
    */
    MetaData    string          `orm:"type(text)" json:"metaData,omitempty"`
    MetaDataObj CronjobMetaData `orm:"-" json:"-"`
    App         *App            `orm:"index;rel(fk)" json:"app,omitempty"`
    Description string          `orm:"null;size(512)" json:"description,omitempty"`
    OrderId     int64           `orm:"index;default(0)" json:"order"`

    CreateTime *time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
    UpdateTime *time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
    User       string     `orm:"size(128)" json:"user,omitempty"`
    Deleted    bool       `orm:"default(false)" json:"deleted,omitempty"`

    AppId int64 `orm:"-" json:"appId,omitempty"`
}

func (*cronjobModel) GetById(id int64) (v *Cronjob, err error) {
    v = &Cronjob{Id: id}
    if err = Ormer().Read(v); err == nil {
        v.AppId = v.App.Id
        return v, nil
    }
    return nil, err
}

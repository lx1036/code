package models

import (
	kapi "k8s.io/api/core/v1"
	"time"
)

const (
	TableNameStatefulset = "statefulset"
)

type statefulsetModel struct{}

/* 存储元数据
{
  "replicas": {
    "K8S": 1
  },
  "privileged":{"nginx",true},
  "affinity": {
    "podAntiAffinity": {
      "requiredDuringSchedulingIgnoredDuringExecution": [
        {
          "labelSelector": {
            "matchExpressions": [
              {
                "operator": "In",
                "values": [
                  "xxx"
                ],
                "key": "app"
              }
            ]
          },
          "topologyKey": "kubernetes.io/hostname"
        }
      ]
    }
  },
  "resources":{
		"cpuRequestLimitPercent": "50%", // cpu request和limit百分比，默认50%
		"memoryRequestLimitPercent": "100%", // memory request和limit百分比，默认100%
		"cpuLimit":"12",  // cpu限制，默认12个核
		"memoryLimit":"64" // 内存限制，默认64G
		"replicaLimit":"32" // 份数限制，默认32份
  }
}
*/
type StatefulsetMetaData struct {
	Replicas  map[string]int32  `json:"replicas"`
	Resources map[string]string `json:"resources,omitempty"`
	Affinity  *kapi.Affinity    `json:"affinity,omitempty"`
	// 是否允许用户使用特权模式，默认不允许,key 为容器名称
	Privileged map[string]*bool `json:"privileged"`
}

type Statefulset struct {
	Id   int64  `orm:"auto" json:"id,omitempty"`
	Name string `orm:"unique;index;size(128)" json:"name,omitempty"`

	MetaData    string              `orm:"type(text)" json:"metaData,omitempty"`
	MetaDataObj StatefulsetMetaData `orm:"-" json:"-"`
	App         *App                `orm:"index;rel(fk)" json:"app,omitempty"`
	Description string              `orm:"null;size(512)" json:"description,omitempty"`
	OrderId     int64               `orm:"index;default(0)" json:"order"`

	CreateTime time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `orm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `orm:"default(false)" json:"deleted,omitempty"`

	AppId int64 `orm:"-" json:"appId,omitempty"`
}

func (*Statefulset) TableName() string {
	return TableNameStatefulset
}

package models

import (
	kapi "k8s.io/api/core/v1"
	"time"
)

const (
	TableNameStatefulSet = "stateful_sets"
)

type statefulsetModel struct{}

/*
	存储元数据

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

type StatefulSet struct {
	ID          uint       `gorm:"column:id;primary_key;"`
	Name        string     `gorm:"column:name;size:128;not null;unique;default:'';"`
	MetaData    string     `gorm:"column:meta_data;type:longtext;not null;"`
	AppId       uint       `gorm:"column:app_id;size:20;not null;"`
	Description string     `gorm:"column:description;size:512;default:null;"`
	OrderId     uint       `gorm:"column:order_id;size:20;"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;default:null;"`

	//Id   int64  `gorm:"auto" json:"id,omitempty"`
	//Name string `gorm:"unique;index;size(128)" json:"name,omitempty"`
	//
	//MetaData    string              `gorm:"type(text)" json:"metaData,omitempty"`
	//MetaDataObj StatefulsetMetaData `gorm:"-" json:"-"`
	//App         *App                `gorm:"index;rel(fk)" json:"app,omitempty"`
	//Description string              `gorm:"null;size(512)" json:"description,omitempty"`
	//OrderId     int64               `gorm:"index;default(0)" json:"order"`
	//
	//CreateTime time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	//UpdateTime time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	//User       string    `gorm:"size(128)" json:"user,omitempty"`
	//Deleted    bool      `gorm:"default(false)" json:"deleted,omitempty"`
	//
	//AppId int64 `gorm:"-" json:"appId,omitempty"`
}

func (StatefulSet) TableName() string {
	return TableNameStatefulSet
}

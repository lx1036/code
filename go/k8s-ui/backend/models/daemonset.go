package models

import (
	kapi "k8s.io/api/core/v1"
	"time"
)

const (
	TableNameDaemonSet = "daemon_sets"
)

/*
	存储元数据

	{
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
type DaemonSetMetaData struct {
	Resources map[string]string `json:"resources,omitempty"`
	Affinity  *kapi.Affinity    `json:"affinity,omitempty"`
	// 是否允许用户使用特权模式，默认不允许,key 为容器名称
	Privileged map[string]*bool `json:"privileged"`
}

type DaemonSet struct {
	ID          uint       `gorm:"column:id;primary_key;"`
	Name        string     `gorm:"column:name;size:128;not null;unique;default:'';"`
	MetaData    string     `gorm:"column:meta_data;type:longtext;not null;"`
	AppId       uint       `gorm:"column:app_id;size:20;not null;"`
	Description string     `gorm:"column:description;size:512;default:null;"`
	OrderId     uint       `gorm:"column:order_id;size:20;"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;default:null;"`
}

func (DaemonSet) TableName() string {
	return TableNameDaemonSet
}

type daemonSetModel struct{}

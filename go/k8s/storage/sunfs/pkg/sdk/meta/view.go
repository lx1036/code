package meta

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"k8s-lx1036/k8s/storage/sunfs/pkg/proto"

	"k8s.io/klog/v2"
)

type VolStatInfo struct {
	Name      string
	TotalSize uint64
	UsedSize  uint64
}

// POST http://{master_ip}:9500/client/volStat?name={pv-name}
/*
{
    "code": 0,
    "msg": "success",
    "data": {
        "Name": "pvc-liuxiang",
        "TotalSize": 1073741824,
        "UsedSize": 6
    }
}
*/
// 获取 volume stats 数据
func (mw *MetaWrapper) updateVolStatInfo() error {
	params := make(map[string]string)
	params["name"] = mw.volname
	body, err := mw.master.Request(http.MethodPost, proto.ClientVolStat, params, nil)
	if err != nil {
		klog.Errorf("updateVolStatInfo request: err(%v)", err)
		return err
	}

	info := new(VolStatInfo)
	if err = json.Unmarshal(body, info); err != nil {
		klog.Errorf("updateVolStatInfo unmarshal: err(%v)", err)
		return err
	}
	atomic.StoreUint64(&mw.totalSize, info.TotalSize)
	atomic.StoreUint64(&mw.usedSize, info.UsedSize)

	klog.Infof(fmt.Sprintf("VolStatInfo: %+v", *info))
	return nil
}

// POST http://{master_ip}:9500/admin/getIp
/*
{
    "code": 0,
    "msg": "success",
    "data": {
        "Cluster": "test-sunfs",
        "Ip": "101.20.30.40"
    }
}
*/
func (mw *MetaWrapper) updateClusterInfo() error {
	body, err := mw.master.Request(http.MethodPost, proto.AdminGetIP, nil, nil)
	if err != nil {
		klog.Error(err)
		return err
	}

	info := new(proto.ClusterInfo)
	if err = json.Unmarshal(body, info); err != nil {
		klog.Errorf("updateClusterInfo unmarshal: err(%v)", err)
		return err
	}

	klog.V(5).Infof("ClusterInfo: %v", *info)
	mw.cluster = info.Cluster
	mw.localIP = info.Ip

	klog.Infof(fmt.Sprintf("ClusterInfo: %+v", *info))
	return nil
}

// POST http://{master_ip}:9500/admin/getVol
/*
{
    "code": 0,
    "msg": "success",
    "data": {
        "ID": 5,
        "Name": "pvc-liuxiang",
        "Owner": "sunfs",
        "MpReplicaNum": 3,
        "Status": 0,
        "Capacity": 1,
        "MpCnt": 3,
        "S3Endpoint": "http://test.s3.cn",
        "BucketDeleted": false
    }
}
*/
// 获取 volume 基本数据
func (mw *MetaWrapper) updateVolSimpleInfo() error {
	params := make(map[string]string)
	params["name"] = mw.volname
	body, err := mw.master.Request(http.MethodPost, proto.AdminGetVol, params, nil)
	if err != nil {
		klog.Errorf("updateVolSimpleInfo request: err(%v)", err)
		return err
	}

	info := new(proto.SimpleVolView)
	if err = json.Unmarshal(body, info); err != nil {
		klog.Errorf("updateVolSimpleInfo body unmarshal: err(%v)", err)
		return err
	}

	mw.S3Endpoint = info.S3Endpoint
	klog.Infof(fmt.Sprintf("SimpleVolView: %+v", *info))
	return nil
}

func (mw *MetaWrapper) updateMetaPartitions() error {
	view, err := mw.fetchVolumeView()
	if err != nil {
		return err
	}

	metaPartitions := make([]*MetaPartition, 0)
	for _, metaPartition := range view.MetaPartitions {
		mw.replaceOrInsertPartition(metaPartition)
		if metaPartition.Status == proto.ReadWrite {
			metaPartitions = append(metaPartitions, metaPartition)
		}
	}

	if len(metaPartitions) == 0 {
		klog.Infof("updateMetaPartition: no read-write meta partitions")
		return nil
	}

	mw.Lock()
	defer mw.Unlock()
	mw.rwPartitions = metaPartitions
	return nil
}

type VolumeView struct {
	VolName        string           `json:"Name"`
	MetaPartitions []*MetaPartition `json:"MetaPartitions"`
}

// POST http://{master_ip}:9500/client/vol
/*
{
    "code": 0,
    "msg": "success",
    "data": {
        "Name": "pvc-liuxiang",
        "Status": 0,
        "MetaPartitions": [
            {
                "PartitionID": 3,
                "Start": 33554433,
                "End": 9223372036854775807,
                "Members": [
                    "101.206.77.175:9021",
                    "101.206.77.176:9021",
                    "101.206.77.177:9021"
                ],
                "LeaderAddr": "101.206.77.176:9021",
                "Status": 2
            },
            {
                "PartitionID": 1,
                "Start": 0,
                "End": 16777216,
                "Members": [
                    "101.206.77.175:9021",
                    "101.206.77.176:9021",
                    "101.206.77.177:9021"
                ],
                "LeaderAddr": "101.206.77.175:9021",
                "Status": 2
            },
            {
                "PartitionID": 2,
                "Start": 16777217,
                "End": 33554432,
                "Members": [
                    "101.206.77.175:9021",
                    "101.206.77.176:9021",
                    "101.206.77.177:9021"
                ],
                "LeaderAddr": "101.206.77.177:9021",
                "Status": 2
            }
        ]
    }
}
*/
func (mw *MetaWrapper) fetchVolumeView() (*VolumeView, error) {
	params := make(map[string]string)
	params["name"] = mw.volname
	body, err := mw.master.Request(http.MethodPost, proto.ClientVol, params, nil)
	if err != nil {
		klog.Errorf("fetchVolumeView request: err(%v)", err)
		return nil, err
	}

	view := new(VolumeView)
	if err = json.Unmarshal(body, view); err != nil {
		klog.Errorf("fetchVolumeView unmarshal: err(%v) body(%v)", err, string(body))
		return nil, err
	}

	klog.Infof(fmt.Sprintf("VolumeView: %+v", *view))
	return view, nil
}

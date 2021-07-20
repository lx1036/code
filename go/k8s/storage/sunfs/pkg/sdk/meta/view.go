package meta

import (
	"encoding/json"
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

	klog.Infof("VolStatInfo: info(%v)", *info)
	return nil
}

// POST http://{master_ip}:9500/admin/getIp
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
	return nil
}

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
	klog.V(5).Infof("VolSimpleInfo: %+v", *info)
	return nil
}

func (mw *MetaWrapper) updateMetaPartitions() error {
	view, err := mw.fetchVolumeView()
	if err != nil {
		return err
	}

	rwPartitions := make([]*MetaPartition, 0)
	for _, mp := range view.MetaPartitions {
		mw.replaceOrInsertPartition(mp)
		klog.Infof("updateMetaPartition: mp(%v)", mp)
		if mp.Status == proto.ReadWrite {
			rwPartitions = append(rwPartitions, mp)
		}
	}

	if len(rwPartitions) == 0 {
		klog.Infof("updateMetaPartition: no rw partitions")
		return nil
	}

	mw.Lock()
	defer mw.Unlock()
	mw.rwPartitions = rwPartitions
	return nil
}

type VolumeView struct {
	VolName        string           `json:"Name"`
	MetaPartitions []*MetaPartition `json:"MetaPartitions"`
}

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

	return view, nil
}

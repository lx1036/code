package meta

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"k8s-lx1036/k8s/storage/sunfs/pkg/util/proto"

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

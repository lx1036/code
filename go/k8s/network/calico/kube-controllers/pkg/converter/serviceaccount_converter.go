package converter

import (
	"fmt"

	apisv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/k8s/conversion"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

type serviceAccountConverter struct {
}

func (s serviceAccountConverter) Convert(k8sObj interface{}) (interface{}, error) {
	converter := conversion.NewConverter()
	serviceAccount, ok := k8sObj.(*corev1.ServiceAccount)
	if !ok {
		tombstone, ok := k8sObj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, fmt.Errorf("couldn't get object from tombstone %+v", k8sObj)
		}
		serviceAccount, ok = tombstone.Obj.(*corev1.ServiceAccount)
		if !ok {
			return nil, fmt.Errorf("tombstone contained object that is not a Serviceaccount %+v", k8sObj)
		}
	}

	kvPair, err := converter.ServiceAccountToProfile(serviceAccount)
	if err != nil {
		return nil, err
	}

	profile := kvPair.Value.(*apisv3.Profile)
	// 只关心Name字段忽略其他字段，如ResourceVersion, CreationTimestamp等，可以避免不必要的更新
	profile.ObjectMeta = metav1.ObjectMeta{Name: profile.Name}

	return *profile, nil
}

func (s serviceAccountConverter) GetKey(obj interface{}) string {
	return obj.(apisv3.Profile).Name
}

func (s serviceAccountConverter) DeleteArgsFromKey(key string) (string, string) {
	// Not serviceaccount, so just return the key, which is the profile name.
	return "", key
}

func NewServiceAccountConverter() Converter {
	return &serviceAccountConverter{}
}

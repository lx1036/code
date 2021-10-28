package aliyun

import (
	"fmt"

	"k8s.io/client-go/util/flowcontrol"
)

type OpenAPI struct {
	ClientSet *ClientSet

	ReadOnlyRateLimiter flowcontrol.RateLimiter
	MutatingRateLimiter flowcontrol.RateLimiter
}

func NewAliyun(ak, sk, regionID, credentialPath string) (*OpenAPI, error) {
	if regionID == "" {
		return nil, fmt.Errorf("regionID unset")
	}
	clientSet, err := NewClientSet(ak, sk, credentialPath, regionID)
	if err != nil {
		return nil, fmt.Errorf("error get clientset, %w", err)
	}
	return &OpenAPI{
		ClientSet:           clientSet,
		ReadOnlyRateLimiter: flowcontrol.NewTokenBucketRateLimiter(8, 10),
		MutatingRateLimiter: flowcontrol.NewTokenBucketRateLimiter(4, 5),
	}, nil
}

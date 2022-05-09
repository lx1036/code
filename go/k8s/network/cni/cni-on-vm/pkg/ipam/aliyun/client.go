package aliyun

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"

	"k8s.io/klog/v2"
)

var (
	kubernetesAlicloudIdentity = "Kubernetes.Alicloud"

	tokenReSyncPeriod = 5 * time.Minute
)

func clientCfg() *sdk.Config {
	return &sdk.Config{
		Timeout:   20 * time.Second,
		Transport: http.DefaultTransport,
		UserAgent: kubernetesAlicloudIdentity,
		Scheme:    "HTTPS",
	}
}

// ClientSet manager of aliyun openapi clientset
type ClientSet struct {
	sync.RWMutex

	regionID string

	auth Interface

	expireAt time.Time
	updateAt time.Time

	ecs *ecs.Client
	vpc *vpc.Client
}

func NewClientSet(key, secret, credentialPath, regionID string) (*ClientSet, error) {
	clientSet := &ClientSet{
		regionID: regionID,
	}
	providers := []Interface{
		NewAKPairProvider(key, secret),
		NewEncryptedCredentialProvider(credentialPath),
		NewMetadataProvider(),
	}
	for _, p := range providers {
		c, err := p.Resolve()
		if err != nil {
			return nil, err
		}
		if c == nil {
			continue
		}
		clientSet.auth = p
		klog.Infof(fmt.Sprintf("using %s provider", clientSet.auth.Name()))
		break
	}
	if clientSet.auth == nil {
		return nil, fmt.Errorf("unable to found a valid credential provider")
	}

	return clientSet, nil
}

func (c *ClientSet) VPC() *vpc.Client {
	c.Lock()
	defer c.Unlock()
	ok, err := c.refreshToken()
	if err != nil {
		klog.Errorf(fmt.Sprintf("create vpc client err:%v", err))
	}
	if ok {
		klog.Infof(fmt.Sprintf("vpc credential update updateAt:%s expireAt:%s", c.updateAt.String(), c.expireAt.String()))
	}

	return c.vpc
}

func (c *ClientSet) ECS() *ecs.Client {
	c.Lock()
	defer c.Unlock()
	ok, err := c.refreshToken()
	if err != nil {
		klog.Errorf(fmt.Sprintf("create ecs client err:%v", err))
	}
	if ok {
		klog.Infof(fmt.Sprintf("ecs credential update updateAt:%s expireAt:%s", c.updateAt.String(), c.expireAt.String()))
	}
	return c.ecs
}

func (c *ClientSet) refreshToken() (bool, error) {
	if c.updateAt.IsZero() || c.expireAt.Before(time.Now()) || time.Since(c.updateAt) > tokenReSyncPeriod {
		var err error
		defer func() {
			if err == nil {
				c.updateAt = time.Now()
			}
		}()

		cc, err := c.auth.Resolve()
		if err != nil {
			return false, err
		}

		c.ecs, err = ecs.NewClientWithOptions(c.regionID, clientCfg(), cc.Credential)
		if err != nil {
			return false, err
		}
		// 使用默认地址
		//c.ecs.SetEndpointRules(c.ecs.EndpointMap, "regional", "vpc")

		c.vpc, err = vpc.NewClientWithOptions(c.regionID, clientCfg(), cc.Credential)
		if err != nil {
			return false, err
		}
		// 使用默认地址
		//c.vpc.SetEndpointRules(c.vpc.EndpointMap, "regional", "vpc")

		c.expireAt = cc.Expiration
		return true, nil
	}

	return false, nil
}

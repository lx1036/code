package aliyun

import (
	"fmt"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"k8s.io/klog/v2"
)

func TestEIP(test *testing.T) {
	clientSet, err := NewAliyun("ak", "sk", "cn-beijing", "~/.aliyun/config.json")
	if err != nil {
		klog.Fatalf(fmt.Sprintf("create aliyun clientset err:%v", err))
	}

	request := ecs.CreateDescribeInstancesRequest()
	response, err := clientSet.ClientSet.ECS().DescribeInstances(request)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("DescribeInstances err:%v", err))
	}

	klog.Infof(response.String())

}

func TestName(test *testing.T) {
	client, err := sdk.NewClientWithAccessKey("cn-hangzhou", "ak", "sk")
	if err != nil {
		// Handle exceptions
		panic(err)
	}

	client.GetHttpProxy()
}

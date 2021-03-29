package scraper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type KubeletClient struct {
	client *http.Client
	scheme string

	buffers sync.Pool
}

func NewKubeletClient(config *rest.Config) (*KubeletClient, error) {
	transport, err := rest.TransportFor(config)
	if err != nil {
		return nil, fmt.Errorf("unable to construct transport: %v", err)
	}

	return &KubeletClient{
		scheme: "https",
		client: &http.Client{
			Transport: transport,
		},
		buffers: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}, nil
}

func (kubeletClient *KubeletClient) getBuffer() *bytes.Buffer {
	return kubeletClient.buffers.Get().(*bytes.Buffer)
}

// 存储数据
func (kubeletClient *KubeletClient) returnBuffer(b *bytes.Buffer) {
	b.Reset()
	kubeletClient.buffers.Put(b)
}

func (kubeletClient *KubeletClient) GetSummary(ctx context.Context, node *corev1.Node) (*Summary, error) {
	nodeStatusPort := node.Status.DaemonEndpoints.KubeletEndpoint.Port
	u := url.URL{
		Scheme:   kubeletClient.scheme,
		Host:     net.JoinHostPort(kubeletClient.NodeAddress(node, corev1.NodeExternalIP), strconv.Itoa(int(nodeStatusPort))),
		Path:     "/stats/summary",
		RawQuery: "only_cpu_and_memory=true",
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	response, err := kubeletClient.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// 把数据放入 sync.Pool 来存储数据
	b := kubeletClient.getBuffer()
	defer kubeletClient.returnBuffer(b)
	_, err = io.Copy(b, response.Body)
	if err != nil {
		return nil, err
	}
	body := b.Bytes()
	if response.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found url %s", req.URL.String())
	} else if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed status %s", response.Status)
	}

	summary := &Summary{}
	err = json.Unmarshal(body, summary)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

func (kubeletClient *KubeletClient) NodeAddress(node *corev1.Node, nodeAddressType corev1.NodeAddressType) string {
	for _, addresses := range node.Status.Addresses {
		if addresses.Type == nodeAddressType {
			return addresses.Address
		}
	}

	return ""
}

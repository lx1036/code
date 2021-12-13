package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

/*
apiVersion: crihookserver.k9s.io/v1
kind: HookConfiguration
timeout: 10
listenAddress: unix:///var/run/cri-hook-server.sock
webhooks:
- name: lighthouse.io
  endpoint: unix://@lighthouse-hook
  failurePolicy: Fail
  stages:
  - urlPattern: /containers/create
	method: post
	type: PreHook
  - urlPattern: /containers/create
	method: post
	type: PostHook
*/

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true

type HookConfiguration struct {
	metav1.TypeMeta `json:",inline"`
	// Timeout tell hook manager how long to wait each requests
	// +optional
	Timeout time.Duration `json:"timeout,omitempty"`
	// ListenAddress tell hook manager listen address for requests
	// +optional
	ListenAddress string `json:"listenAddress,omitempty"`
	// RemoteEndpoint tell hook manager the remote address to forward requests.
	// +optional
	RemoteEndpoint string `json:"remoteEndpoint,omitempty"`
	// WebHooks tell hook manager the configuration for webhook rules
	// +optional
	WebHooks WebHooks `json:"webhooks,omitempty"`
}

type WebHooks []WebHook

// WebHook describe hook policy
type WebHook struct {
	// Name represents the name of webhook
	Name string `json:"name,omitempty"`
	// Endpoint represents the backend endpoint address to receive the requests
	Endpoint string `json:"endpoint,omitempty"`
	// FailurePolicy tells the default policy when requests failed
	FailurePolicy FailurePolicyType `json:"failurePolicy,omitempty"`
	// Stages tells the hook manager which stage the requests sent to the backend
	Stages HookStageList `json:"stages,omitempty"`
}

type FailurePolicyType string

const (
	// PolicyFail returns error when got an error
	PolicyFail FailurePolicyType = "Fail"
	// PolicyIgnore returns nothing when got an error
	PolicyIgnore FailurePolicyType = "Ignore"
)

type HookStageList []HookStage

// HookStage describe hook stage
type HookStage struct {
	// Method tell hook manager which http method it will accept
	Method string `json:"method,omitempty"`
	// URLPattern tell hook manager the url pattern for the request to be accepted
	URLPattern string `json:"urlPattern,omitempty"`
	// Type tell hook manager when to send the request to the backend
	Type HookType `json:"type,omitempty"`
}

type HookType string

const (
	// PreHookType means the request need to be handled before send to original backend
	PreHookType HookType = "PreHook"
	// PostHookType means the response should be handled after receiving from the original backend
	PostHookType HookType = "PostHook"
)

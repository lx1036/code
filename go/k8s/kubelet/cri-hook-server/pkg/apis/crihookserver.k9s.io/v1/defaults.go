package v1

import "net/http"

func SetDefaults_HookConfiguration(obj *HookConfiguration)  {
	if obj.Timeout == 0 {
		obj.Timeout = 5
	}
	
	if obj.RemoteEndpoint == "" {
		obj.RemoteEndpoint = "unix:///var/run/docker.sock"
	}
}

func SetDefaults_WebHook(obj *WebHook) {
	if obj.FailurePolicy == "" {
		obj.FailurePolicy = PolicyFail
	}
}

func SetDefaults_HookStage(obj *HookStage) {
	if obj.Method == "" {
		obj.Method = http.MethodPost
	}
}

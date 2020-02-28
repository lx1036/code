package api

import (
	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/authorization/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// ClientManager is responsible for initializing and creating clients to communicate with
// kubernetes apiserver on demand.
type ClientManager interface {
	Client(req *restful.Request) (kubernetes.Interface, error)
	InsecureClient() kubernetes.Interface
	APIExtensionsClient(req *restful.Request) (apiextensionsclientset.Interface, error)
	PluginClient(req *restful.Request) (pluginclientset.Interface, error)
	InsecureAPIExtensionsClient() apiextensionsclientset.Interface
	InsecurePluginClient() pluginclientset.Interface
	CanI(req *restful.Request, ssar *v1.SelfSubjectAccessReview) bool
	Config(req *restful.Request) (*rest.Config, error)
	ClientCmdConfig(req *restful.Request) (clientcmd.ClientConfig, error)
	CSRFKey() string
	HasAccess(authInfo api.AuthInfo) error
	VerberClient(req *restful.Request, config *rest.Config) (ResourceVerber, error)
	SetTokenManager(manager authApi.TokenManager)
}

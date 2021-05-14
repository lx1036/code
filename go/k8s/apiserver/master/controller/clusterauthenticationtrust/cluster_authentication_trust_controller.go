package clusterauthenticationtrust

import (
	"k8s.io/apiserver/pkg/authentication/request/headerrequest"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
)

// ClusterAuthenticationInfo holds the information that will included in public configmap.
type ClusterAuthenticationInfo struct {
	// ClientCA is the CA that can be used to verify the identity of normal clients
	ClientCA dynamiccertificates.CAContentProvider

	// RequestHeaderUsernameHeaders are the headers used by this kube-apiserver to determine username
	RequestHeaderUsernameHeaders headerrequest.StringSliceProvider
	// RequestHeaderGroupHeaders are the headers used by this kube-apiserver to determine groups
	RequestHeaderGroupHeaders headerrequest.StringSliceProvider
	// RequestHeaderExtraHeaderPrefixes are the headers used by this kube-apiserver to determine user.extra
	RequestHeaderExtraHeaderPrefixes headerrequest.StringSliceProvider
	// RequestHeaderAllowedNames are the sujbects allowed to act as a front proxy
	RequestHeaderAllowedNames headerrequest.StringSliceProvider
	// RequestHeaderCA is the CA that can be used to verify the front proxy
	RequestHeaderCA dynamiccertificates.CAContentProvider
}

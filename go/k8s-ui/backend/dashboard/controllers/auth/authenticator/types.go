package authenticator

type LoginSpec struct {
	// Basic authentication
	Username string `json:"username,omitempty" form:"username"`
	Password string `json:"password,omitempty" form:"password"`
	// Token authentication
	Token string `json:"token,omitempty" form:"token"`
	// .kubeconfig file authentication
	KubeConfig string `json:"kubeconfig,omitempty" form:"kubeconfig"`
}

type AuthResponse struct {
	JweToken string `json:"jweToken"`
}

type LoginModesResponse struct {
	Modes []AuthenticationMode `json:"modes"`
}

type AuthenticationMode string

const (
	Token AuthenticationMode = "token"
	Basic AuthenticationMode = "basic"
)

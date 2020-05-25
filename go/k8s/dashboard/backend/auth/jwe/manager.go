package jwe

import (
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/tools/clientcmd/api"
	"time"
)

// Implements TokenManager interface
type jweTokenManager struct {
	keyHolder KeyHolder
	tokenTTL  time.Duration
}

func (self *jweTokenManager) Generate(authInfo api.AuthInfo) (string, error) {
	marshalledAuthInfo, err := json.Marshal(authInfo)
	if err != nil {
		return "", err
	}

}

func (self *jweTokenManager) Decrypt(string) (*api.AuthInfo, error) {
	panic("implement me")
}

func (self *jweTokenManager) Refresh(string) (string, error) {
	panic("implement me")
}

func (self *jweTokenManager) SetTokenTTL(time.Duration) {
	panic("implement me")
}

// Creates and returns default JWE token manager instance.
func NewJWETokenManager(holder KeyHolder) authApi.TokenManager {
	manager := &jweTokenManager{keyHolder: holder, tokenTTL: authApi.DefaultTokenTTL * time.Second}
	return manager
}

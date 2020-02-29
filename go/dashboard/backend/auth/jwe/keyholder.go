package jwe

import (
	"crypto/rsa"
	"gopkg.in/square/go-jose.v2"
	syncApi "k8s-lx1036/dashboard/backend/sync/api"
	"sync"
)

// Entries held by resource used to synchronize encryption key data.
const (
	holderMapKeyEntry  = "priv"
	holderMapCertEntry = "pub"
)

// KeyHolder is responsible for generating, storing and synchronizing encryption key used for token
// generation/decryption.
type KeyHolder interface {
	// Returns encrypter instance that can be used to encrypt data.
	Encrypter() jose.Encrypter
	// Returns encryption key that can be used to decrypt data.
	Key() *rsa.PrivateKey
	// Forces refresh of encryption key synchronized with kubernetes resource (secret).
	Refresh()
}

// Implements KeyHolder interface
type rsaKeyHolder struct {
	// 256-byte random RSA key pair. Synced with a key saved in a secret.
	key          *rsa.PrivateKey
	synchronizer syncApi.Synchronizer
	mux          sync.Mutex
}

func (self *rsaKeyHolder) Encrypter() jose.Encrypter {
	panic("implement me")
}

func (self *rsaKeyHolder) Key() *rsa.PrivateKey {
	panic("implement me")
}

func (self *rsaKeyHolder) Refresh() {
	panic("implement me")
}

func (self *rsaKeyHolder) init() {

}

// NewRSAKeyHolder creates new KeyHolder instance.
func NewRSAKeyHolder(synchronizer syncApi.Synchronizer) KeyHolder {
	holder := &rsaKeyHolder{
		synchronizer: synchronizer,
	}

	holder.init()
	return holder
}

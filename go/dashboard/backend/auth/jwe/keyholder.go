package jwe

import (
	"crypto/rand"
	"crypto/rsa"
	"gopkg.in/square/go-jose.v2"
	syncApi "k8s-lx1036/dashboard/backend/sync/api"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"sync"
	"log"
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

// Encrypter implements key holder interface. See KeyHolder for more information.
// Used encryption algorithms:
//    - Content encryption: AES-GCM (256)
//    - Key management: RSA-OAEP-SHA256
func (self *rsaKeyHolder) Encrypter() jose.Encrypter {
	publicKey := &self.Key().PublicKey
	encrypter, err := jose.NewEncrypter(jose.A256GCM, jose.Recipient{Algorithm: jose.RSA_OAEP_256, Key: publicKey}, nil)
	if err != nil {
		panic(err)
	}

	return encrypter
}

// Key implements key holder interface. See KeyHolder for more information.
func (self *rsaKeyHolder) Key() *rsa.PrivateKey {
	self.mux.Lock()
	defer self.mux.Unlock()
	return self.key
}

func (self *rsaKeyHolder) Refresh() {
	panic("implement me")
}

func (self *rsaKeyHolder) init() {
	self.initEncryptionKey()

	// Register event handlers
	self.synchronizer.RegisterActionHandler(self.update, watch.Added, watch.Modified)
	self.synchronizer.RegisterActionHandler(self.recreate, watch.Deleted)
}

// Handler function executed by synchronizer used to store encryption key. It is called whenever watched object
// is created or updated.
func (self *rsaKeyHolder) update(obj runtime.Object) {

}
// Handler function executed by synchronizer used to store encryption key. It is called whenever watched object
// gets deleted. It is then recreated based on local key.
func (self *rsaKeyHolder) recreate(obj runtime.Object) {
	secret := obj.(*v1.Secret)
	log.Printf("Synchronized secret %s has been deleted. Recreating.", secret.Name)
	if err := self.synchronizer.Create(self.getEncryptionKeyHolder()); err != nil {
		panic(err)
	}
}

func (self *rsaKeyHolder) getEncryptionKeyHolder() runtime.Object {
	priv, pub := ExportRSAKeyOrDie(self.Key())


}

// Generates encryption key used to encrypt token payload.
func (self *rsaKeyHolder) initEncryptionKey() {
	log.Print("Generating JWE encryption key")
	self.mux.Lock()
	defer self.mux.Unlock()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	self.key = privateKey
}

// NewRSAKeyHolder creates new KeyHolder instance.
func NewRSAKeyHolder(synchronizer syncApi.Synchronizer) KeyHolder {
	holder := &rsaKeyHolder{
		synchronizer: synchronizer,
	}

	holder.init()
	return holder
}

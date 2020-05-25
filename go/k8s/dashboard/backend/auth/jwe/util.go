package jwe

import "crypto/rsa"

// Credits to David W. https://stackoverflow.com/a/44688503

// ExportRSAKeyOrDie exports rsa key object to a private/public strings. In case of fail panic is called.
func ExportRSAKeyOrDie(privKey *rsa.PrivateKey) (priv, pub string) {

}

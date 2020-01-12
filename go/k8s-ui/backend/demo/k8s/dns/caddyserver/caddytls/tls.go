package caddytls

import "github.com/mholt/certmagic"

// Revoke revokes the certificate fro host via the ACME protocol.
// It assumes the certificate was obtained from certmagic.CA.
func Revoke(domainName string) error {
	return certmagic.NewDefault().RevokeCert(domainName, true)
}

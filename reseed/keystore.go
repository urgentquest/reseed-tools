package reseed

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
)

// KeyStore manages certificate and key storage for the reseed service.
// Moved from: utils.go
type KeyStore struct {
	Path string
}

// NewKeyStore creates a new KeyStore instance with the specified path.
// Moved from: utils.go
func NewKeyStore(path string) *KeyStore {
	return &KeyStore{
		Path: path,
	}
}

// ReseederCertificate loads a reseed certificate for the given signer.
// Moved from: utils.go
func (ks *KeyStore) ReseederCertificate(signer []byte) (*x509.Certificate, error) {
	return ks.reseederCertificate("reseed", signer)
}

// DirReseederCertificate loads a reseed certificate from a specific directory.
// Moved from: utils.go
func (ks *KeyStore) DirReseederCertificate(dir string, signer []byte) (*x509.Certificate, error) {
	return ks.reseederCertificate(dir, signer)
}

// reseederCertificate is a helper method to load certificates from the keystore.
// Moved from: utils.go
func (ks *KeyStore) reseederCertificate(dir string, signer []byte) (*x509.Certificate, error) {
	certFile := filepath.Base(SignerFilename(string(signer)))
	certString, err := os.ReadFile(filepath.Join(ks.Path, dir, certFile))
	if nil != err {
		return nil, err
	}

	certPem, _ := pem.Decode(certString)
	return x509.ParseCertificate(certPem.Bytes)
}

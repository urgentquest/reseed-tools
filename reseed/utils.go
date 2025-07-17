package reseed

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"strings"
	"time"
)

// KeyStore struct and methods moved to keystore.go

// SignerFilename creates a certificate filename from signer ID.
// Uses SignerFilenameFromID for consistency.
// Moved from: multiple files
func SignerFilename(signer string) string {
	return SignerFilenameFromID(signer) + ".crt"
}

func NewTLSCertificate(host string, priv *ecdsa.PrivateKey) ([]byte, error) {
	return NewTLSCertificateAltNames(priv, host)
}

func NewTLSCertificateAltNames(priv *ecdsa.PrivateKey, hosts ...string) ([]byte, error) {
	notBefore := time.Now()
	notAfter := notBefore.Add(5 * 365 * 24 * time.Hour)
	host := ""
	if len(hosts) > 0 {
		host = hosts[0]
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"I2P Anonymous Network"},
			OrganizationalUnit: []string{"I2P"},
			Locality:           []string{"XX"},
			StreetAddress:      []string{"XX"},
			Country:            []string{"XX"},
			CommonName:         host,
		},
		NotBefore:          notBefore,
		NotAfter:           notAfter,
		SignatureAlgorithm: x509.ECDSAWithSHA512,

		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              hosts[1:],
	}

	hosts = strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	return derBytes, nil
}

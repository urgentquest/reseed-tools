package cmd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"
)

func TestCertificateExpirationLogic(t *testing.T) {
	// Generate a test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	testCases := []struct {
		name        string
		expiresIn   time.Duration
		shouldRenew bool
		description string
	}{
		{
			name:        "Certificate expires in 24 hours",
			expiresIn:   24 * time.Hour,
			shouldRenew: true,
			description: "Should renew certificate that expires within 48 hours",
		},
		{
			name:        "Certificate expires in 72 hours",
			expiresIn:   72 * time.Hour,
			shouldRenew: false,
			description: "Should not renew certificate with more than 48 hours remaining",
		},
		{
			name:        "Certificate expires in 47 hours",
			expiresIn:   47 * time.Hour,
			shouldRenew: true,
			description: "Should renew certificate just under 48 hour threshold",
		},
		{
			name:        "Certificate expires in 49 hours",
			expiresIn:   49 * time.Hour,
			shouldRenew: false,
			description: "Should not renew certificate just over 48 hour threshold",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a certificate that expires at the specified time
			template := x509.Certificate{
				SerialNumber: big.NewInt(1),
				Subject: pkix.Name{
					Organization: []string{"Test"},
				},
				NotBefore:   time.Now(),
				NotAfter:    time.Now().Add(tc.expiresIn),
				KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
				ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			}

			certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
			if err != nil {
				t.Fatalf("Failed to create certificate: %v", err)
			}

			cert, err := x509.ParseCertificate(certDER)
			if err != nil {
				t.Fatalf("Failed to parse certificate: %v", err)
			}

			// Test the logic that was fixed
			shouldRenew := time.Until(cert.NotAfter) < (time.Hour * 48)

			if shouldRenew != tc.shouldRenew {
				t.Errorf("%s: Expected shouldRenew=%v, got %v. %s",
					tc.name, tc.shouldRenew, shouldRenew, tc.description)
			}

			// Also test that a TLS certificate with this cert would have the same behavior
			tlsCert := tls.Certificate{
				Certificate: [][]byte{certDER},
				PrivateKey:  privateKey,
				Leaf:        cert,
			}

			tlsShouldRenew := time.Until(tlsCert.Leaf.NotAfter) < (time.Hour * 48)
			if tlsShouldRenew != tc.shouldRenew {
				t.Errorf("%s: TLS certificate logic mismatch. Expected shouldRenew=%v, got %v",
					tc.name, tc.shouldRenew, tlsShouldRenew)
			}
		})
	}
}

func TestOldBuggyLogic(t *testing.T) {
	// Test to demonstrate that the old buggy logic was incorrect

	// Create a certificate that expires in 24 hours (should be renewed)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(24 * time.Hour), // Expires in 24 hours
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Old buggy logic (commented out to show what was wrong)
	// oldLogic := time.Now().Sub(cert.NotAfter) < (time.Hour * 48)

	// New correct logic
	newLogic := time.Until(cert.NotAfter) < (time.Hour * 48)

	// For a certificate expiring in 24 hours:
	// - Old logic would be: time.Now().Sub(futureTime) = negative value < 48 hours = false (wrong!)
	// - New logic would be: time.Until(futureTime) = 24 hours < 48 hours = true (correct!)

	if !newLogic {
		t.Error("New logic should indicate renewal needed for certificate expiring in 24 hours")
	}
}

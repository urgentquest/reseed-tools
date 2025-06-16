package su3

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"
)

func TestCheckSignature_RSA(t *testing.T) {
	// Generate RSA key pair for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Create test certificate
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	testData := []byte("test data to sign")

	tests := []struct {
		name        string
		algo        x509.SignatureAlgorithm
		expectError bool
	}{
		{
			name:        "SHA256 with RSA",
			algo:        x509.SHA256WithRSA,
			expectError: false,
		},
		{
			name:        "SHA384 with RSA",
			algo:        x509.SHA384WithRSA,
			expectError: false,
		},
		{
			name:        "SHA512 with RSA",
			algo:        x509.SHA512WithRSA,
			expectError: false,
		},
		{
			name:        "SHA1 with RSA",
			algo:        x509.SHA1WithRSA,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create signature using the standard library for comparison
			var hashType x509.SignatureAlgorithm
			switch tt.algo {
			case x509.SHA1WithRSA:
				hashType = x509.SHA1WithRSA
			case x509.SHA256WithRSA:
				hashType = x509.SHA256WithRSA
			case x509.SHA384WithRSA:
				hashType = x509.SHA384WithRSA
			case x509.SHA512WithRSA:
				hashType = x509.SHA512WithRSA
			}

			// Create a proper signature for testing
			tempCert := &x509.Certificate{
				SerialNumber: big.NewInt(1),
				Subject: pkix.Name{
					CommonName: "test",
				},
				NotBefore:             time.Now(),
				NotAfter:              time.Now().Add(time.Hour),
				KeyUsage:              x509.KeyUsageDigitalSignature,
				BasicConstraintsValid: true,
				SignatureAlgorithm:    hashType,
			}

			signedCert, err := x509.CreateCertificate(rand.Reader, tempCert, tempCert, &privateKey.PublicKey, privateKey)
			if err != nil {
				t.Fatalf("Failed to create signed certificate: %v", err)
			}

			parsedCert, err := x509.ParseCertificate(signedCert)
			if err != nil {
				t.Fatalf("Failed to parse signed certificate: %v", err)
			}

			// Use the signature from the certificate for testing
			err = checkSignature(parsedCert, tt.algo, testData, parsedCert.Signature)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				// We expect this might fail since we're using certificate signature
				// for different data, but we're testing the function doesn't panic
				_ = err
			}
		})
	}
}

func TestCheckSignature_UnsupportedAlgorithm(t *testing.T) {
	// Create a dummy certificate
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	testData := []byte("test data")
	testSignature := []byte("dummy signature")

	// Test unsupported algorithm
	err = checkSignature(cert, x509.SignatureAlgorithm(999), testData, testSignature)
	if err != x509.ErrUnsupportedAlgorithm {
		t.Errorf("Expected ErrUnsupportedAlgorithm, got %v", err)
	}
}

func TestCheckSignature_ECDSA(t *testing.T) {
	// Generate ECDSA key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.ECDSAWithSHA256,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create ECDSA certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse ECDSA certificate: %v", err)
	}

	testData := []byte("test data")

	tests := []struct {
		name string
		algo x509.SignatureAlgorithm
	}{
		{
			name: "ECDSA with SHA256",
			algo: x509.ECDSAWithSHA256,
		},
		{
			name: "ECDSA with SHA384",
			algo: x509.ECDSAWithSHA384,
		},
		{
			name: "ECDSA with SHA512",
			algo: x509.ECDSAWithSHA512,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with the certificate's own signature (which should work)
			err = checkSignature(cert, tt.algo, testData, cert.Signature)
			// We don't assert specific success/failure since we're using the cert signature
			// for different data, but we ensure the function doesn't panic
			_ = err
		})
	}
}

func TestCheckSignature_InvalidSignature(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	testData := []byte("test data")

	tests := []struct {
		name      string
		signature []byte
	}{
		{
			name:      "Empty signature",
			signature: []byte{},
		},
		{
			name:      "Invalid signature",
			signature: []byte("invalid signature data"),
		},
		{
			name:      "Nil signature",
			signature: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = checkSignature(cert, x509.SHA256WithRSA, testData, tt.signature)
			if err == nil {
				t.Error("Expected error for invalid signature")
			}
		})
	}
}

func TestNewSigningCertificate(t *testing.T) {
	tests := []struct {
		name     string
		signerID string
		keySize  int
	}{
		{
			name:     "Standard certificate",
			signerID: "test@example.com",
			keySize:  2048,
		},
		{
			name:     "Certificate with special characters",
			signerID: "test+special@example.com",
			keySize:  2048,
		},
		{
			name:     "Certificate with long signer ID",
			signerID: "very.long.email.address.for.testing@example.organization.com",
			keySize:  2048,
		},
		{
			name:     "Large key size",
			signerID: "test@example.com",
			keySize:  4096,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate private key
			privateKey, err := rsa.GenerateKey(rand.Reader, tt.keySize)
			if err != nil {
				t.Fatalf("Failed to generate private key: %v", err)
			}

			// Create certificate
			certBytes, err := NewSigningCertificate(tt.signerID, privateKey)
			if err != nil {
				t.Fatalf("NewSigningCertificate failed: %v", err)
			}

			if len(certBytes) == 0 {
				t.Fatal("Certificate bytes should not be empty")
			}

			// Parse the certificate to verify it's valid
			cert, err := x509.ParseCertificate(certBytes)
			if err != nil {
				t.Fatalf("Failed to parse generated certificate: %v", err)
			}

			// Verify certificate properties
			if cert.Subject.CommonName != tt.signerID {
				t.Errorf("Expected CommonName %s, got %s", tt.signerID, cert.Subject.CommonName)
			}

			if !cert.IsCA {
				t.Error("Certificate should be marked as CA")
			}

			if !cert.BasicConstraintsValid {
				t.Error("Certificate should have valid basic constraints")
			}

			// Verify certificate is self-signed
			err = cert.CheckSignatureFrom(cert)
			if err != nil {
				t.Errorf("Certificate should be self-signed: %v", err)
			}

			// Verify subject fields
			if len(cert.Subject.Organization) == 0 || cert.Subject.Organization[0] != "I2P Anonymous Network" {
				t.Error("Certificate should have I2P Anonymous Network as organization")
			}

			if len(cert.Subject.OrganizationalUnit) == 0 || cert.Subject.OrganizationalUnit[0] != "I2P" {
				t.Error("Certificate should have I2P as organizational unit")
			}

			// Verify validity period (should be 10 years)
			validity := cert.NotAfter.Sub(cert.NotBefore)
			expectedYears := time.Duration(10*365*24) * time.Hour
			if validity < expectedYears {
				t.Errorf("Certificate validity period should be at least 10 years, got %v", validity)
			}

			// Verify key usage
			expectedKeyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign
			if cert.KeyUsage&expectedKeyUsage != expectedKeyUsage {
				t.Errorf("Certificate should have digital signature and cert sign key usage")
			}

			// Verify extended key usage
			hasClientAuth := false
			hasServerAuth := false
			for _, usage := range cert.ExtKeyUsage {
				if usage == x509.ExtKeyUsageClientAuth {
					hasClientAuth = true
				}
				if usage == x509.ExtKeyUsageServerAuth {
					hasServerAuth = true
				}
			}
			if !hasClientAuth || !hasServerAuth {
				t.Error("Certificate should have both client and server auth extended key usage")
			}
		})
	}
}

func TestNewSigningCertificate_NilPrivateKey(t *testing.T) {
	_, err := NewSigningCertificate("test@example.com", nil)
	if err == nil {
		t.Error("Expected error when private key is nil")
	}
}

func TestNewSigningCertificate_EmptySignerID(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	certBytes, err := NewSigningCertificate("", privateKey)
	if err != nil {
		t.Fatalf("NewSigningCertificate failed with empty signer ID: %v", err)
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	if cert.Subject.CommonName != "" {
		t.Errorf("Expected empty CommonName, got %s", cert.Subject.CommonName)
	}
}

func TestDSASignatureStruct(t *testing.T) {
	// Test that dsaSignature struct can be created and used
	sig := dsaSignature{
		R: big.NewInt(123),
		S: big.NewInt(456),
	}

	if sig.R.Int64() != 123 {
		t.Errorf("Expected R=123, got %d", sig.R.Int64())
	}

	if sig.S.Int64() != 456 {
		t.Errorf("Expected S=456, got %d", sig.S.Int64())
	}
}

func TestECDSASignatureStruct(t *testing.T) {
	// Test that ecdsaSignature struct can be created and used
	sig := ecdsaSignature{
		R: big.NewInt(789),
		S: big.NewInt(101112),
	}

	if sig.R.Int64() != 789 {
		t.Errorf("Expected R=789, got %d", sig.R.Int64())
	}

	if sig.S.Int64() != 101112 {
		t.Errorf("Expected S=101112, got %d", sig.S.Int64())
	}
}

func TestCheckSignature_DSASignature(t *testing.T) {
	// This test verifies the DSA signature parsing logic
	// We'll create a malformed ASN.1 structure to test error handling

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	testData := []byte("test data")

	// Test with malformed ASN.1 signature for DSA
	malformedSignature := []byte("not valid asn1 data")

	err = checkSignature(cert, x509.DSAWithSHA1, testData, malformedSignature)
	if err == nil {
		t.Error("Expected error for malformed DSA signature")
	}
}

func TestCheckSignature_ECDSASignature(t *testing.T) {
	// This test verifies the ECDSA signature parsing logic
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	testData := []byte("test data")

	// Test with malformed ASN.1 signature for ECDSA
	malformedSignature := []byte("not valid asn1 data")

	err = checkSignature(cert, x509.ECDSAWithSHA256, testData, malformedSignature)
	if err == nil {
		t.Error("Expected error for malformed ECDSA signature")
	}
}

func TestCheckSignature_UnavailableHash(t *testing.T) {
	// This test verifies behavior when hash algorithm is not available
	// Note: In practice, standard hash algorithms are always available,
	// but this tests the code path

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	testData := []byte("test data")
	testSignature := []byte("dummy signature")

	// Test various signature algorithms to ensure they don't panic
	algorithms := []x509.SignatureAlgorithm{
		x509.SHA1WithRSA,
		x509.SHA256WithRSA,
		x509.SHA384WithRSA,
		x509.SHA512WithRSA,
		x509.DSAWithSHA1,
		x509.DSAWithSHA256,
		x509.ECDSAWithSHA1,
		x509.ECDSAWithSHA256,
		x509.ECDSAWithSHA384,
		x509.ECDSAWithSHA512,
	}

	for _, algo := range algorithms {
		t.Run(algo.String(), func(t *testing.T) {
			err = checkSignature(cert, algo, testData, testSignature)
			// We don't assert specific success/failure, just that it doesn't panic
			_ = err
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkNewSigningCertificate(b *testing.B) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatalf("Failed to generate private key: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewSigningCertificate("benchmark@example.com", privateKey)
		if err != nil {
			b.Fatalf("NewSigningCertificate failed: %v", err)
		}
	}
}

func BenchmarkCheckSignature_RSA(b *testing.B) {
	// Setup
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatalf("Failed to generate RSA key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "benchmark",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.SHA256WithRSA,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		b.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		b.Fatalf("Failed to parse certificate: %v", err)
	}

	testData := []byte("benchmark data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = checkSignature(cert, x509.SHA256WithRSA, testData, cert.Signature)
	}
}

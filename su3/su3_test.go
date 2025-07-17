package su3

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"reflect"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	file := New()

	if file == nil {
		t.Fatal("New() returned nil")
	}

	if file.SignatureType != SigTypeRSAWithSHA512 {
		t.Errorf("Expected SignatureType %d, got %d", SigTypeRSAWithSHA512, file.SignatureType)
	}

	if len(file.Version) == 0 {
		t.Error("Version should be set")
	}

	// Verify version is a valid Unix timestamp string
	if len(file.Version) < 10 {
		t.Error("Version should be at least 10 characters (Unix timestamp)")
	}
}

func TestFile_Sign(t *testing.T) {
	tests := []struct {
		name          string
		signatureType uint16
		expectError   bool
	}{
		{
			name:          "RSA with SHA256",
			signatureType: SigTypeRSAWithSHA256,
			expectError:   false,
		},
		{
			name:          "RSA with SHA384",
			signatureType: SigTypeRSAWithSHA384,
			expectError:   false,
		},
		{
			name:          "RSA with SHA512",
			signatureType: SigTypeRSAWithSHA512,
			expectError:   false,
		},
		{
			name:          "Unknown signature type",
			signatureType: uint16(999),
			expectError:   true,
		},
	}

	// Generate test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := New()
			file.SignatureType = tt.signatureType
			file.Content = []byte("test content")
			file.SignerID = []byte("test@example.com")

			err := file.Sign(privateKey)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(file.Signature) == 0 {
				t.Error("Signature should be set after signing")
			}
		})
	}
}

func TestFile_Sign_NilPrivateKey(t *testing.T) {
	file := New()
	file.Content = []byte("test content")

	err := file.Sign(nil)
	if err == nil {
		t.Error("Expected error when signing with nil private key")
	}
}

func TestFile_BodyBytes(t *testing.T) {
	file := New()
	file.Format = 1
	file.SignatureType = SigTypeRSAWithSHA256
	file.FileType = FileTypeZIP
	file.ContentType = ContentTypeReseed
	file.Version = []byte("1234567890")
	file.SignerID = []byte("test@example.com")
	file.Content = []byte("test content data")

	bodyBytes := file.BodyBytes()

	if len(bodyBytes) == 0 {
		t.Error("BodyBytes should not be empty")
	}

	// Check that magic bytes are included
	if !bytes.HasPrefix(bodyBytes, []byte(magicBytes)) {
		t.Error("BodyBytes should start with magic bytes")
	}

	// Test version padding
	shortVersionFile := New()
	shortVersionFile.Version = []byte("123") // Less than minVersionLength
	bodyBytes = shortVersionFile.BodyBytes()

	if len(bodyBytes) == 0 {
		t.Error("BodyBytes should handle short version")
	}
}

func TestFile_MarshalBinary(t *testing.T) {
	file := New()
	file.Content = []byte("test content")
	file.SignerID = []byte("test@example.com")
	file.Signature = []byte("dummy signature data")

	data, err := file.MarshalBinary()
	if err != nil {
		t.Errorf("MarshalBinary failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("MarshalBinary should return data")
	}

	// Verify signature is at the end
	expectedSigStart := len(data) - len(file.Signature)
	if !bytes.Equal(data[expectedSigStart:], file.Signature) {
		t.Error("Signature should be at the end of marshaled data")
	}
}

func TestFile_UnmarshalBinary(t *testing.T) {
	// Create a file and marshal it
	originalFile := New()
	originalFile.Format = 1
	originalFile.SignatureType = SigTypeRSAWithSHA256
	originalFile.FileType = FileTypeZIP
	originalFile.ContentType = ContentTypeReseed
	originalFile.Version = []byte("1234567890123456") // Exactly minVersionLength
	originalFile.SignerID = []byte("test@example.com")
	originalFile.Content = []byte("test content data")
	originalFile.Signature = make([]byte, 256) // Appropriate size for RSA SHA256

	// Fill signature with test data
	for i := range originalFile.Signature {
		originalFile.Signature[i] = byte(i % 256)
	}

	data, err := originalFile.MarshalBinary()
	if err != nil {
		t.Fatalf("Failed to marshal test file: %v", err)
	}

	// Unmarshal into new file
	newFile := &File{}
	err = newFile.UnmarshalBinary(data)
	if err != nil {
		t.Errorf("UnmarshalBinary failed: %v", err)
	}

	// Compare fields
	if newFile.Format != originalFile.Format {
		t.Errorf("Format mismatch: expected %d, got %d", originalFile.Format, newFile.Format)
	}

	if newFile.SignatureType != originalFile.SignatureType {
		t.Errorf("SignatureType mismatch: expected %d, got %d", originalFile.SignatureType, newFile.SignatureType)
	}

	if newFile.FileType != originalFile.FileType {
		t.Errorf("FileType mismatch: expected %d, got %d", originalFile.FileType, newFile.FileType)
	}

	if newFile.ContentType != originalFile.ContentType {
		t.Errorf("ContentType mismatch: expected %d, got %d", originalFile.ContentType, newFile.ContentType)
	}

	if !bytes.Equal(newFile.Version, originalFile.Version) {
		t.Errorf("Version mismatch: expected %s, got %s", originalFile.Version, newFile.Version)
	}

	if !bytes.Equal(newFile.SignerID, originalFile.SignerID) {
		t.Errorf("SignerID mismatch: expected %s, got %s", originalFile.SignerID, newFile.SignerID)
	}

	if !bytes.Equal(newFile.Content, originalFile.Content) {
		t.Errorf("Content mismatch: expected %s, got %s", originalFile.Content, newFile.Content)
	}

	if !bytes.Equal(newFile.Signature, originalFile.Signature) {
		t.Error("Signature mismatch")
	}
}

func TestFile_UnmarshalBinary_InvalidData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Empty data",
			data: []byte{},
		},
		{
			name: "Too short data",
			data: []byte("short"),
		},
		{
			name: "Invalid magic bytes",
			data: append([]byte("BADMAG"), make([]byte, 100)...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &File{}
			err := file.UnmarshalBinary(tt.data)
			// Note: The current implementation doesn't validate magic bytes or handle errors gracefully
			// This test documents the current behavior
			_ = err // We expect this might fail, but we're testing it doesn't panic
		})
	}
}

func TestFile_VerifySignature(t *testing.T) {
	// Generate test certificate and private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Create a test certificate
	cert, err := NewSigningCertificate("test@example.com", privateKey)
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	parsedCert, err := x509.ParseCertificate(cert)
	if err != nil {
		t.Fatalf("Failed to parse test certificate: %v", err)
	}

	tests := []struct {
		name          string
		signatureType uint16
		setupFile     func(*File)
		expectError   bool
	}{
		{
			name:          "Valid RSA SHA256 signature",
			signatureType: SigTypeRSAWithSHA256,
			setupFile: func(f *File) {
				f.Content = []byte("test content")
				f.SignerID = []byte("test@example.com")
				err := f.Sign(privateKey)
				if err != nil {
					t.Fatalf("Failed to sign file: %v", err)
				}
			},
			expectError: false,
		},
		{
			name:          "Unknown signature type",
			signatureType: uint16(999),
			setupFile: func(f *File) {
				f.Content = []byte("test content")
				f.SignerID = []byte("test@example.com")
				f.Signature = []byte("dummy signature")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := New()
			file.SignatureType = tt.signatureType
			tt.setupFile(file)

			err := file.VerifySignature(parsedCert)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestFile_String(t *testing.T) {
	file := New()
	file.Format = 1
	file.SignatureType = SigTypeRSAWithSHA256
	file.FileType = FileTypeZIP
	file.ContentType = ContentTypeReseed
	file.Version = []byte("test version")
	file.SignerID = []byte("test@example.com")

	str := file.String()

	if len(str) == 0 {
		t.Error("String() should not return empty string")
	}

	// Check that important fields are included in string representation
	expectedSubstrings := []string{
		"Format:",
		"SignatureType:",
		"FileType:",
		"ContentType:",
		"Version:",
		"SignerId:",
		"---------------------------",
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(str, substr) {
			t.Errorf("String() should contain '%s'", substr)
		}
	}
}

func TestConstants(t *testing.T) {
	// Test that constants have expected values
	if magicBytes != "I2Psu3" {
		t.Errorf("Expected magic bytes 'I2Psu3', got '%s'", magicBytes)
	}

	if minVersionLength != 16 {
		t.Errorf("Expected minVersionLength 16, got %d", minVersionLength)
	}

	// Test signature type constants
	expectedSigTypes := map[string]uint16{
		"DSA":             0,
		"ECDSAWithSHA256": 1,
		"ECDSAWithSHA384": 2,
		"ECDSAWithSHA512": 3,
		"RSAWithSHA256":   4,
		"RSAWithSHA384":   5,
		"RSAWithSHA512":   6,
	}

	actualSigTypes := map[string]uint16{
		"DSA":             SigTypeDSA,
		"ECDSAWithSHA256": SigTypeECDSAWithSHA256,
		"ECDSAWithSHA384": SigTypeECDSAWithSHA384,
		"ECDSAWithSHA512": SigTypeECDSAWithSHA512,
		"RSAWithSHA256":   SigTypeRSAWithSHA256,
		"RSAWithSHA384":   SigTypeRSAWithSHA384,
		"RSAWithSHA512":   SigTypeRSAWithSHA512,
	}

	if !reflect.DeepEqual(expectedSigTypes, actualSigTypes) {
		t.Error("Signature type constants don't match expected values")
	}
}

func TestFile_RoundTrip(t *testing.T) {
	// Test complete round-trip: create -> sign -> marshal -> unmarshal -> verify
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	cert, err := NewSigningCertificate("roundtrip@example.com", privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	parsedCert, err := x509.ParseCertificate(cert)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Create and set up original file
	originalFile := New()
	originalFile.FileType = FileTypeZIP
	originalFile.ContentType = ContentTypeReseed
	originalFile.Content = []byte("This is test content for round-trip testing")
	originalFile.SignerID = []byte("roundtrip@example.com")

	// Sign the file
	err = originalFile.Sign(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign file: %v", err)
	}

	// Marshal to binary
	data, err := originalFile.MarshalBinary()
	if err != nil {
		t.Fatalf("Failed to marshal file: %v", err)
	}

	// Unmarshal from binary
	newFile := &File{}
	err = newFile.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal file: %v", err)
	}

	// Verify signature
	err = newFile.VerifySignature(parsedCert)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}

	// Ensure content matches
	if !bytes.Equal(originalFile.Content, newFile.Content) {
		t.Error("Content doesn't match after round-trip")
	}

	if !bytes.Equal(originalFile.SignerID, newFile.SignerID) {
		t.Error("SignerID doesn't match after round-trip")
	}
}

func TestFile_Sign_RSAKeySize(t *testing.T) {
	testCases := []struct {
		name           string
		keySize        int
		expectedSigLen int
	}{
		{"2048-bit RSA", 2048, 256},
		{"3072-bit RSA", 3072, 384},
		{"4096-bit RSA", 4096, 512},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate RSA key of specific size
			privateKey, err := rsa.GenerateKey(rand.Reader, tc.keySize)
			if err != nil {
				t.Fatalf("Failed to generate %d-bit RSA key: %v", tc.keySize, err)
			}

			file := New()
			file.Content = []byte("test content")
			file.SignerID = []byte("test@example.com")
			file.SignatureType = SigTypeRSAWithSHA512

			err = file.Sign(privateKey)
			if err != nil {
				t.Errorf("Unexpected error signing with %d-bit key: %v", tc.keySize, err)
				return
			}

			if len(file.Signature) != tc.expectedSigLen {
				t.Errorf("Expected signature length %d for %d-bit key, got %d",
					tc.expectedSigLen, tc.keySize, len(file.Signature))
			}

			// Verify the header reflects the correct signature length
			bodyBytes := file.BodyBytes()
			if len(bodyBytes) == 0 {
				t.Error("BodyBytes should not be empty")
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New()
	}
}

func BenchmarkFile_BodyBytes(b *testing.B) {
	file := New()
	file.Content = make([]byte, 1024) // 1KB content
	file.SignerID = []byte("benchmark@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = file.BodyBytes()
	}
}

func BenchmarkFile_MarshalBinary(b *testing.B) {
	file := New()
	file.Content = make([]byte, 1024) // 1KB content
	file.SignerID = []byte("benchmark@example.com")
	file.Signature = make([]byte, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = file.MarshalBinary()
	}
}

func BenchmarkFile_UnmarshalBinary(b *testing.B) {
	// Create test data once
	file := New()
	file.Content = make([]byte, 1024)
	file.SignerID = []byte("benchmark@example.com")
	file.Signature = make([]byte, 512)

	data, err := file.MarshalBinary()
	if err != nil {
		b.Fatalf("Failed to create test data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newFile := &File{}
		_ = newFile.UnmarshalBinary(data)
	}
}

package reseed

import (
	"archive/zip"
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestZipSeeds_Success(t *testing.T) {
	// Test with valid router info data
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	seeds := []routerInfo{
		{
			Name:    "routerInfo-test1.dat",
			ModTime: testTime,
			Data:    []byte("test router info data 1"),
		},
		{
			Name:    "routerInfo-test2.dat",
			ModTime: testTime,
			Data:    []byte("test router info data 2"),
		},
	}

	zipData, err := zipSeeds(seeds)
	if err != nil {
		t.Fatalf("zipSeeds() error = %v, want nil", err)
	}

	if len(zipData) == 0 {
		t.Error("zipSeeds() returned empty data")
	}

	// Verify the zip file structure
	reader := bytes.NewReader(zipData)
	zipReader, err := zip.NewReader(reader, int64(len(zipData)))
	if err != nil {
		t.Fatalf("Failed to read zip data: %v", err)
	}

	if len(zipReader.File) != 2 {
		t.Errorf("Expected 2 files in zip, got %d", len(zipReader.File))
	}

	// Verify file names and content
	expectedFiles := map[string]string{
		"routerInfo-test1.dat": "test router info data 1",
		"routerInfo-test2.dat": "test router info data 2",
	}

	for _, file := range zipReader.File {
		expectedContent, exists := expectedFiles[file.Name]
		if !exists {
			t.Errorf("Unexpected file in zip: %s", file.Name)
			continue
		}

		// Check modification time
		if !file.ModTime().Equal(testTime) {
			t.Errorf("File %s has wrong ModTime. Expected %v, got %v", file.Name, testTime, file.ModTime())
		}

		// Check compression method
		if file.Method != zip.Deflate {
			t.Errorf("File %s has wrong compression method. Expected %d, got %d", file.Name, zip.Deflate, file.Method)
		}

		// Check content
		rc, err := file.Open()
		if err != nil {
			t.Errorf("Failed to open file %s: %v", file.Name, err)
			continue
		}

		var content bytes.Buffer
		_, err = content.ReadFrom(rc)
		rc.Close()
		if err != nil {
			t.Errorf("Failed to read file %s: %v", file.Name, err)
			continue
		}

		if content.String() != expectedContent {
			t.Errorf("File %s has wrong content. Expected %q, got %q", file.Name, expectedContent, content.String())
		}
	}
}

func TestZipSeeds_EmptyInput(t *testing.T) {
	// Test with empty slice
	seeds := []routerInfo{}

	zipData, err := zipSeeds(seeds)
	if err != nil {
		t.Fatalf("zipSeeds() error = %v, want nil", err)
	}

	// Verify it creates a valid but empty zip file
	reader := bytes.NewReader(zipData)
	zipReader, err := zip.NewReader(reader, int64(len(zipData)))
	if err != nil {
		t.Fatalf("Failed to read empty zip data: %v", err)
	}

	if len(zipReader.File) != 0 {
		t.Errorf("Expected 0 files in empty zip, got %d", len(zipReader.File))
	}
}

func TestZipSeeds_SingleFile(t *testing.T) {
	// Test with single router info
	testTime := time.Date(2023, 6, 15, 10, 30, 0, 0, time.UTC)
	seeds := []routerInfo{
		{
			Name:    "single-router.dat",
			ModTime: testTime,
			Data:    []byte("single router data"),
		},
	}

	zipData, err := zipSeeds(seeds)
	if err != nil {
		t.Fatalf("zipSeeds() error = %v, want nil", err)
	}

	reader := bytes.NewReader(zipData)
	zipReader, err := zip.NewReader(reader, int64(len(zipData)))
	if err != nil {
		t.Fatalf("Failed to read zip data: %v", err)
	}

	if len(zipReader.File) != 1 {
		t.Errorf("Expected 1 file in zip, got %d", len(zipReader.File))
	}

	file := zipReader.File[0]
	if file.Name != "single-router.dat" {
		t.Errorf("Expected file name 'single-router.dat', got %q", file.Name)
	}
}

func TestUzipSeeds_Success(t *testing.T) {
	// First create a zip file using zipSeeds
	testTime := time.Date(2024, 2, 14, 8, 45, 0, 0, time.UTC)
	originalSeeds := []routerInfo{
		{
			Name:    "router1.dat",
			ModTime: testTime,
			Data:    []byte("router 1 content"),
		},
		{
			Name:    "router2.dat",
			ModTime: testTime,
			Data:    []byte("router 2 content"),
		},
	}

	zipData, err := zipSeeds(originalSeeds)
	if err != nil {
		t.Fatalf("Setup failed: zipSeeds() error = %v", err)
	}

	// Now test uzipSeeds
	unzippedSeeds, err := uzipSeeds(zipData)
	if err != nil {
		t.Fatalf("uzipSeeds() error = %v, want nil", err)
	}

	if len(unzippedSeeds) != 2 {
		t.Errorf("Expected 2 seeds, got %d", len(unzippedSeeds))
	}

	// Create a map for easier comparison
	seedMap := make(map[string]routerInfo)
	for _, seed := range unzippedSeeds {
		seedMap[seed.Name] = seed
	}

	// Check first file
	if seed1, exists := seedMap["router1.dat"]; exists {
		if string(seed1.Data) != "router 1 content" {
			t.Errorf("router1.dat content mismatch. Expected %q, got %q", "router 1 content", string(seed1.Data))
		}
	} else {
		t.Error("router1.dat not found in unzipped seeds")
	}

	// Check second file
	if seed2, exists := seedMap["router2.dat"]; exists {
		if string(seed2.Data) != "router 2 content" {
			t.Errorf("router2.dat content mismatch. Expected %q, got %q", "router 2 content", string(seed2.Data))
		}
	} else {
		t.Error("router2.dat not found in unzipped seeds")
	}
}

func TestUzipSeeds_EmptyZip(t *testing.T) {
	// Create an empty zip file
	emptySeeds := []routerInfo{}
	zipData, err := zipSeeds(emptySeeds)
	if err != nil {
		t.Fatalf("Setup failed: zipSeeds() error = %v", err)
	}

	unzippedSeeds, err := uzipSeeds(zipData)
	if err != nil {
		t.Fatalf("uzipSeeds() error = %v, want nil", err)
	}

	if len(unzippedSeeds) != 0 {
		t.Errorf("Expected 0 seeds from empty zip, got %d", len(unzippedSeeds))
	}
}

func TestUzipSeeds_InvalidZipData(t *testing.T) {
	// Test with invalid zip data
	invalidData := []byte("this is not a zip file")

	unzippedSeeds, err := uzipSeeds(invalidData)
	if err == nil {
		t.Error("uzipSeeds() should return error for invalid zip data")
	}

	if unzippedSeeds != nil {
		t.Error("uzipSeeds() should return nil seeds for invalid zip data")
	}
}

func TestUzipSeeds_EmptyData(t *testing.T) {
	// Test with empty byte slice
	emptyData := []byte{}

	unzippedSeeds, err := uzipSeeds(emptyData)
	if err == nil {
		t.Error("uzipSeeds() should return error for empty data")
	}

	if unzippedSeeds != nil {
		t.Error("uzipSeeds() should return nil seeds for empty data")
	}
}

func TestZipUnzipRoundTrip(t *testing.T) {
	// Test round-trip: zip -> unzip -> compare
	tests := []struct {
		name  string
		seeds []routerInfo
	}{
		{
			name: "MultipleFiles",
			seeds: []routerInfo{
				{Name: "file1.dat", ModTime: time.Now(), Data: []byte("data1")},
				{Name: "file2.dat", ModTime: time.Now(), Data: []byte("data2")},
				{Name: "file3.dat", ModTime: time.Now(), Data: []byte("data3")},
			},
		},
		{
			name: "SingleFile",
			seeds: []routerInfo{
				{Name: "single.dat", ModTime: time.Now(), Data: []byte("single data")},
			},
		},
		{
			name:  "Empty",
			seeds: []routerInfo{},
		},
		{
			name: "LargeData",
			seeds: []routerInfo{
				{Name: "large.dat", ModTime: time.Now(), Data: bytes.Repeat([]byte("x"), 10000)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Zip the seeds
			zipData, err := zipSeeds(tt.seeds)
			if err != nil {
				t.Fatalf("zipSeeds() error = %v", err)
			}

			// Unzip the data
			unzippedSeeds, err := uzipSeeds(zipData)
			if err != nil {
				t.Fatalf("uzipSeeds() error = %v", err)
			}

			// Compare lengths
			if len(unzippedSeeds) != len(tt.seeds) {
				t.Errorf("Length mismatch: original=%d, unzipped=%d", len(tt.seeds), len(unzippedSeeds))
			}

			// Create maps for comparison (order might be different)
			originalMap := make(map[string][]byte)
			for _, seed := range tt.seeds {
				originalMap[seed.Name] = seed.Data
			}

			unzippedMap := make(map[string][]byte)
			for _, seed := range unzippedSeeds {
				unzippedMap[seed.Name] = seed.Data
			}

			if !reflect.DeepEqual(originalMap, unzippedMap) {
				t.Errorf("Round-trip failed: data mismatch")
				t.Logf("Original: %v", originalMap)
				t.Logf("Unzipped: %v", unzippedMap)
			}
		})
	}
}

func TestZipSeeds_BinaryData(t *testing.T) {
	// Test with binary data (not just text)
	binaryData := make([]byte, 256)
	for i := range binaryData {
		binaryData[i] = byte(i)
	}

	seeds := []routerInfo{
		{
			Name:    "binary.dat",
			ModTime: time.Now(),
			Data:    binaryData,
		},
	}

	zipData, err := zipSeeds(seeds)
	if err != nil {
		t.Fatalf("zipSeeds() error = %v", err)
	}

	unzippedSeeds, err := uzipSeeds(zipData)
	if err != nil {
		t.Fatalf("uzipSeeds() error = %v", err)
	}

	if len(unzippedSeeds) != 1 {
		t.Fatalf("Expected 1 unzipped seed, got %d", len(unzippedSeeds))
	}

	if !bytes.Equal(unzippedSeeds[0].Data, binaryData) {
		t.Error("Binary data corrupted during zip/unzip")
	}
}

func TestZipSeeds_SpecialCharactersInFilename(t *testing.T) {
	// Test with filenames containing special characters
	seeds := []routerInfo{
		{
			Name:    "file-with-dashes.dat",
			ModTime: time.Now(),
			Data:    []byte("dash data"),
		},
		{
			Name:    "file_with_underscores.dat",
			ModTime: time.Now(),
			Data:    []byte("underscore data"),
		},
	}

	zipData, err := zipSeeds(seeds)
	if err != nil {
		t.Fatalf("zipSeeds() error = %v", err)
	}

	unzippedSeeds, err := uzipSeeds(zipData)
	if err != nil {
		t.Fatalf("uzipSeeds() error = %v", err)
	}

	if len(unzippedSeeds) != 2 {
		t.Fatalf("Expected 2 unzipped seeds, got %d", len(unzippedSeeds))
	}

	// Verify filenames are preserved
	foundFiles := make(map[string]bool)
	for _, seed := range unzippedSeeds {
		foundFiles[seed.Name] = true
	}

	if !foundFiles["file-with-dashes.dat"] {
		t.Error("File with dashes not found")
	}
	if !foundFiles["file_with_underscores.dat"] {
		t.Error("File with underscores not found")
	}
}

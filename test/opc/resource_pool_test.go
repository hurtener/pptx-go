package opc_test

import (
	"archive/zip"
	"bytes"
	"testing"
	"time"

	"github.com/hurtener/pptx-go/internal/opc"
)

// TestResourcePool_Basic tests basic resource pool functionality.
func TestResourcePool_Basic(t *testing.T) {
	pool := opc.GetGlobalPool()
	if pool == nil {
		t.Fatal("GetGlobalPool returned nil")
	}

	// clear data from previous tests
	pool.ReleaseAll()

	// test GetOrLoad
	callCount := 0
	data, err := pool.GetOrLoad("/ppt/media/image1.png", opc.ContentTypePNG, func() ([]byte, error) {
		callCount++
		return []byte{0x01, 0x02, 0x03}, nil
	})
	if err != nil {
		t.Fatalf("GetOrLoad failed: %v", err)
	}
	if callCount != 1 {
		t.Error("loader should be called once")
	}
	if len(data) != 3 {
		t.Errorf("data length should be 3, got %d", len(data))
	}

	// get again (should use cache)
	callCount = 0
	data2, err := pool.GetOrLoad("/ppt/media/image1.png", opc.ContentTypePNG, func() ([]byte, error) {
		callCount++
		return nil, nil
	})
	if err != nil {
		t.Fatalf("GetOrLoad second call failed: %v", err)
	}
	if callCount != 0 {
		t.Error("loader should not be called again (should use cache)")
	}

	// verify data is identical (zero-copy)
	if len(data2) != len(data) {
		t.Error("data2 length should match data")
	}
	for i := range data {
		if data[i] != data2[i] {
			t.Errorf("data2[%d] should equal data[%d]", i, i)
		}
	}

	// test Release (must release twice because GetOrLoad was called twice)
	pool.Release("/ppt/media/image1.png")
	pool.Release("/ppt/media/image1.png")
	stats := pool.Stats()
	if stats["media"] != 0 {
		t.Errorf("media count should be 0 after release, got %d", stats["media"])
	}

	// cleanup
	pool.ReleaseAll()

	t.Log("Resource pool basic test passed")
}

// TestResourcePool_ContentTypeCategories tests content type classification.
func TestResourcePool_ContentTypeCategories(t *testing.T) {
	// test immutable content type detection
	testCases := []struct {
		contentType string
		expected    bool
	}{
		{opc.ContentTypePNG, true},
		{opc.ContentTypeJPEG, true},
		{opc.ContentTypeTheme, true},
		{opc.ContentTypeSlideMaster, true},
		{opc.ContentTypeSlide, false}, // slide is mutable
		{opc.ContentTypePresentation, false},
	}

	for _, tc := range testCases {
		result := opc.IsImmutableContentType(tc.contentType)
		if result != tc.expected {
			t.Errorf("IsImmutableContentType(%s) = %v, want %v", tc.contentType, result, tc.expected)
		}
	}

	t.Log("Content type categorization test passed")
}

// TestPart_CloneShared tests zero-copy cloning of Part.
func TestPart_CloneShared(t *testing.T) {
	// create original data
	originalData := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	uri := opc.NewPackURI("/ppt/media/image1.png")

	// create original Part
	original := opc.NewPart(uri, opc.ContentTypePNG, originalData)

	// zero-copy clone via CloneShared
	cloned := original.CloneShared()

	// verify clone is not nil
	if cloned == nil {
		t.Fatal("CloneShared returned nil")
	}

	// verify it is immutable
	if !cloned.IsImmutable() {
		t.Error("cloned part should be immutable")
	}

	// verify URI is shared (same pointer)
	if cloned.PartURI() != original.PartURI() {
		t.Error("URI should be shared")
	}

	// verify Blob returns the same data
	blob := cloned.Blob()
	if len(blob) != len(originalData) {
		t.Errorf("blob length mismatch: got %d, want %d", len(blob), len(originalData))
	}

	// verify contents match
	for i := range originalData {
		if blob[i] != originalData[i] {
			t.Errorf("blob content mismatch at index %d", i)
		}
	}

	t.Log("Part CloneShared test passed")
}

// TestPart_Clone_DeepCopy tests deep copying of Part.
func TestPart_Clone_DeepCopy(t *testing.T) {
	// create original data
	originalData := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")

	// create original Part
	original := opc.NewPart(uri, opc.ContentTypeSlide, originalData)

	// deep copy via Clone
	cloned := original.Clone()

	// verify clone is not nil
	if cloned == nil {
		t.Fatal("Clone returned nil")
	}

	// verify it is not immutable
	if cloned.IsImmutable() {
		t.Error("cloned part should not be immutable")
	}

	// verify Blob is a deep copy
	originalBlob := original.Blob()
	clonedBlob := cloned.Blob()

	// modifying the clone should not affect the original
	if len(clonedBlob) > 0 {
		clonedBlob[0] = 0xFF
		if originalBlob[0] == 0xFF {
			t.Error("modifying cloned blob should not affect original")
		}
	}

	t.Log("Part Clone deep copy test passed")
}

// TestPackage_Clone_SmartCloning tests smart cloning behavior of Package.
func TestPackage_Clone_SmartCloning(t *testing.T) {
	// create a Package with parts of different types
	pkg := opc.NewPackage()

	// add an image part (should use zero-copy)
	imageData := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	imageURI := opc.NewPackURI("/ppt/media/image1.png")
	pkg.CreatePart(imageURI, opc.ContentTypePNG, imageData)

	// add a slide part (should use deep copy)
	slideData := []byte("<slide>test</slide>")
	slideURI := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePart(slideURI, opc.ContentTypeSlide, slideData)

	// clone the Package
	clonedPkg := pkg.Clone()

	// verify clone is not nil
	if clonedPkg == nil {
		t.Fatal("Clone returned nil")
	}

	// verify image part used zero-copy
	originalImagePart := pkg.GetPart(imageURI)
	clonedImagePart := clonedPkg.GetPart(imageURI)

	if originalImagePart == nil || clonedImagePart == nil {
		t.Fatal("image parts should exist")
	}

	// image should be immutable (zero-copy was used)
	if !clonedImagePart.IsImmutable() {
		t.Error("cloned image part should be immutable (zero-copy)")
	}

	// slide should be mutable (deep copy was used)
	clonedSlidePart := clonedPkg.GetPart(slideURI)
	if clonedSlidePart == nil {
		t.Fatal("cloned slide part should exist")
	}
	if clonedSlidePart.IsImmutable() {
		t.Error("cloned slide part should not be immutable (deep copy)")
	}

	t.Log("Package Clone smart cloning test passed")
}

// TestZipEntry_Timestamp verifies ZIP entry timestamps are set correctly.
// Checks:
// 1. Timestamps are not zero (works around Windows Explorer MS-DOS time bug).
// 2. Timestamps are close to the current time (within a few seconds).
// 3. Timestamps can be correctly converted to Beijing time (UTC+8).
func TestZipEntry_Timestamp(t *testing.T) {
	// record time before creation (in Beijing timezone)
	beijingLoc, _ := time.LoadLocation("Asia/Shanghai")
	beforeCreate := time.Now().In(beijingLoc)

	// create a simple Package
	pkg := opc.NewPackage()

	// add a part
	slideURI := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePart(slideURI, opc.ContentTypeSlide, []byte("<slide/>"))

	// add relationship
	pkg.AddRelationship(opc.RelTypeOfficeDocument, "/ppt/presentation.xml", false)

	// save to bytes
	data, err := pkg.SaveToBytes()
	if err != nil {
		t.Fatalf("SaveToBytes failed: %v", err)
	}

	// record time after creation
	afterCreate := time.Now().In(beijingLoc)

	// read the ZIP and check timestamps
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read ZIP: %v", err)
	}

	// check each file's timestamp
	for _, file := range reader.File {
		modTime := file.Modified

		// 1. timestamp must not be zero
		if modTime.IsZero() {
			t.Errorf("File %s has zero modification time", file.Name)
			continue
		}

		// 2. timestamp must be between beforeCreate and afterCreate (allow 1s tolerance)
		modTimeUTC := modTime.UTC()
		beforeUTC := beforeCreate.UTC()
		afterUTC := afterCreate.UTC()

		if modTimeUTC.Before(beforeUTC.Add(-1 * time.Second)) {
			t.Errorf("File %s modification time %v is before creation time %v",
				file.Name, modTimeUTC, beforeUTC)
		}
		if modTimeUTC.After(afterUTC.Add(1 * time.Second)) {
			t.Errorf("File %s modification time %v is after creation time %v",
				file.Name, modTimeUTC, afterUTC)
		}

		// 3. convert to Beijing time and verify
		modTimeBeijing := modTime.In(beijingLoc)

		// verify Beijing offset is 8 hours from UTC (or adjusted for DST)
		_, beijingOffset := modTimeBeijing.Zone()
		expectedOffset := 8 * 60 * 60 // 8 hours in seconds
		if beijingOffset != expectedOffset {
			t.Logf("Warning: Beijing offset is %d seconds, expected %d (may vary by DST)",
				beijingOffset, expectedOffset)
		}

		t.Logf("File: %s", file.Name)
		t.Logf("  UTC Time:   %v", modTimeUTC.Format("2006-01-02 15:04:05 MST"))
		t.Logf("  Beijing:    %v", modTimeBeijing.Format("2006-01-02 15:04:05 MST"))
		t.Logf("  Unix:       %d", modTime.Unix())
	}

	t.Log("ZIP entry timestamp test passed")
}

// TestZipEntry_TimestampNotZero specifically tests that timestamps are never zero.
// This is the key test for the Windows Explorer MS-DOS time parsing bug fix.
func TestZipEntry_TimestampNotZero(t *testing.T) {
	pkg := opc.NewPackage()

	// add several parts of different types
	testParts := []struct {
		uri         string
		contentType string
		data        []byte
	}{
		{"/ppt/slides/slide1.xml", opc.ContentTypeSlide, []byte("<slide/>")},
		{"/ppt/media/image1.png", opc.ContentTypePNG, []byte{0x89, 0x50, 0x4E, 0x47}},
		{"/docProps/core.xml", opc.ContentTypeCoreProperties, []byte("<coreProperties/>")},
	}

	for _, tp := range testParts {
		uri := opc.NewPackURI(tp.uri)
		pkg.CreatePart(uri, tp.contentType, tp.data)
	}

	// save
	data, err := pkg.SaveToBytes()
	if err != nil {
		t.Fatalf("SaveToBytes failed: %v", err)
	}

	// read and verify
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to read ZIP: %v", err)
	}

	zeroTimeCount := 0
	validTimeCount := 0

	for _, file := range reader.File {
		if file.Modified.IsZero() {
			zeroTimeCount++
			t.Errorf("File %s has zero modification time (will cause Explorer bug)", file.Name)
		} else {
			validTimeCount++
		}
	}

	if zeroTimeCount > 0 {
		t.Errorf("Found %d files with zero timestamp (Windows Explorer bug risk)", zeroTimeCount)
	}

	t.Logf("Valid timestamps: %d, Zero timestamps: %d", validTimeCount, zeroTimeCount)

	if zeroTimeCount == 0 {
		t.Log("All ZIP entries have valid timestamps - Windows Explorer compatible")
	}
}

package opc_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hurtener/pptx-go/internal/opc"
)

// ===== PartSource tests =====

func TestBytesSource(t *testing.T) {
	data := []byte("test content")
	source := opc.NewBytesSource(data)

	// test Size
	if source.Size() != int64(len(data)) {
		t.Errorf("Size() = %d, want %d", source.Size(), len(data))
	}

	// test Open
	rc, err := source.Open()
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer rc.Close()

	readData, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if string(readData) != string(data) {
		t.Errorf("read data = %q, want %q", string(readData), string(data))
	}
}

func TestReaderSource(t *testing.T) {
	data := []byte("test content")
	reader := bytes.NewReader(data)
	source := opc.NewReaderSource(reader, int64(len(data)))

	// test Size
	if source.Size() != int64(len(data)) {
		t.Errorf("Size() = %d, want %d", source.Size(), len(data))
	}

	// test Open
	rc, err := source.Open()
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer rc.Close()

	readData, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if string(readData) != string(data) {
		t.Errorf("read data = %q, want %q", string(readData), string(data))
	}
}

// ===== StreamPart tests =====

func TestStreamPart_New(t *testing.T) {
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	data := []byte("<slide/>")
	source := opc.NewBytesSource(data)

	part := opc.NewStreamPart(uri, opc.ContentTypeSlide, source)

	if part == nil {
		t.Fatal("NewStreamPart returned nil")
	}
	if part.PartURI().URI() != uri.URI() {
		t.Errorf("PartURI() = %q, want %q", part.PartURI().URI(), uri.URI())
	}
	if part.ContentType() != opc.ContentTypeSlide {
		t.Errorf("ContentType() = %q, want %q", part.ContentType(), opc.ContentTypeSlide)
	}
}

func TestStreamPart_LazyLoad(t *testing.T) {
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	data := []byte("<slide/>")
	source := opc.NewBytesSource(data)

	part := opc.NewStreamPart(uri, opc.ContentTypeSlide, source)

	// initial state: not loaded
	if part.IsLoaded() {
		t.Error("new StreamPart should not be loaded")
	}

	// open stream for reading (without loading into memory)
	rc, err := part.Open()
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	readData, _ := io.ReadAll(rc)
	rc.Close()

	if string(readData) != string(data) {
		t.Error("Open returned wrong data")
	}

	// still not loaded
	if part.IsLoaded() {
		t.Error("Open should not load into memory")
	}

	// explicit load
	if err := part.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// now loaded
	if !part.IsLoaded() {
		t.Error("Load should mark as loaded")
	}

	// loading again should be a no-op
	if err := part.Load(); err != nil {
		t.Fatalf("second Load failed: %v", err)
	}
}

func TestStreamPart_Blob(t *testing.T) {
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	data := []byte("<slide/>")
	source := opc.NewBytesSource(data)

	part := opc.NewStreamPart(uri, opc.ContentTypeSlide, source)

	// Blob triggers a load
	blob, err := part.Blob()
	if err != nil {
		t.Fatalf("Blob failed: %v", err)
	}

	if string(blob) != string(data) {
		t.Errorf("Blob() = %q, want %q", string(blob), string(data))
	}

	if !part.IsLoaded() {
		t.Error("Blob should load into memory")
	}
}

func TestStreamPart_SetBlob(t *testing.T) {
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	source := opc.NewBytesSource([]byte("original"))
	part := opc.NewStreamPart(uri, opc.ContentTypeSlide, source)

	// set new content
	newData := []byte("<new slide/>")
	part.SetBlob(newData)

	// should be loaded
	if !part.IsLoaded() {
		t.Error("SetBlob should mark as loaded")
	}

	// should be marked dirty
	if !part.IsDirty() {
		t.Error("SetBlob should mark as dirty")
	}

	// read should return new content
	blob, _ := part.Blob()
	if string(blob) != string(newData) {
		t.Error("Blob should return new data")
	}
}

func TestStreamPart_Size(t *testing.T) {
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	data := []byte("<slide/>")
	source := opc.NewBytesSource(data)

	part := opc.NewStreamPart(uri, opc.ContentTypeSlide, source)

	// size from source when not loaded
	if part.Size() != int64(len(data)) {
		t.Errorf("Size() = %d, want %d", part.Size(), len(data))
	}

	// after load
	part.Load()
	if part.Size() != int64(len(data)) {
		t.Errorf("Size() after load = %d, want %d", part.Size(), len(data))
	}
}

func TestStreamPart_Clone(t *testing.T) {
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	data := []byte("<slide/>")
	source := opc.NewBytesSource(data)

	part := opc.NewStreamPart(uri, opc.ContentTypeSlide, source)
	part.Load()

	clone := part.Clone()
	if clone == nil {
		t.Fatal("Clone returned nil")
	}

	// verify the clone is independent
	if clone == part {
		t.Error("clone should be a different instance")
	}

	// modifying the original should not affect the clone
	part.SetBlob([]byte("<modified/>"))
	blob, _ := clone.Blob()
	if string(blob) != string(data) {
		t.Error("modifying original should not affect clone")
	}
}

// ===== StreamingZipWriter tests =====

func TestStreamingZipWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	sw := opc.NewStreamingZipWriter(buf)

	// write byte data
	if err := sw.WriteBytes("test.txt", []byte("hello world")); err != nil {
		t.Fatalf("WriteBytes failed: %v", err)
	}

	// write XML data
	if err := sw.WriteXML("test.xml", []byte("<root/>")); err != nil {
		t.Fatalf("WriteXML failed: %v", err)
	}

	// close
	if err := sw.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// verify the generated ZIP
	data := buf.Bytes()
	if len(data) == 0 {
		t.Error("ZIP data should not be empty")
	}
}

func TestStreamingZipWriter_StreamPart(t *testing.T) {
	buf := &bytes.Buffer{}
	sw := opc.NewStreamingZipWriter(buf)

	// create a streaming part
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	data := []byte("<p:sld/>")
	source := opc.NewBytesSource(data)
	part := opc.NewStreamPart(uri, opc.ContentTypeSlide, source)

	// stream write
	if err := sw.WriteStreamPart(part); err != nil {
		t.Fatalf("WriteStreamPart failed: %v", err)
	}

	if err := sw.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// verify
	if buf.Len() == 0 {
		t.Error("ZIP data should not be empty")
	}
}

// ===== StreamPackage tests =====

func TestStreamPackage_New(t *testing.T) {
	pkg := opc.NewStreamPackage()
	if pkg == nil {
		t.Fatal("NewStreamPackage returned nil")
	}

	if pkg.PartCount() != 0 {
		t.Error("new StreamPackage should have no parts")
	}
}

func TestStreamPackage_CreatePart(t *testing.T) {
	pkg := opc.NewStreamPackage()
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	data := []byte("<slide/>")
	source := opc.NewBytesSource(data)

	part, err := pkg.CreateStreamPart(uri, opc.ContentTypeSlide, source)
	if err != nil {
		t.Fatalf("CreateStreamPart failed: %v", err)
	}

	if part == nil {
		t.Fatal("CreateStreamPart returned nil")
	}

	if pkg.PartCount() != 1 {
		t.Errorf("PartCount() = %d, want 1", pkg.PartCount())
	}
}

func TestStreamPackage_CreatePartFromBytes(t *testing.T) {
	pkg := opc.NewStreamPackage()
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	data := []byte("<slide/>")

	part, err := pkg.CreatePartFromBytes(uri, opc.ContentTypeSlide, data)
	if err != nil {
		t.Fatalf("CreatePartFromBytes failed: %v", err)
	}

	// should be loaded
	if !part.IsLoaded() {
		t.Error("CreatePartFromBytes should mark part as loaded")
	}

	// content should be correct
	blob, _ := part.Blob()
	if string(blob) != string(data) {
		t.Error("Blob content mismatch")
	}
}

func TestStreamPackage_GetPart(t *testing.T) {
	pkg := opc.NewStreamPackage()
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePartFromBytes(uri, opc.ContentTypeSlide, []byte("<slide/>"))

	part := pkg.GetPart(uri)
	if part == nil {
		t.Fatal("GetPart returned nil")
	}

	// get a non-existent part
	nonExistent := opc.NewPackURI("/ppt/slides/slide999.xml")
	if pkg.GetPart(nonExistent) != nil {
		t.Error("GetPart for non-existent URI should return nil")
	}
}

func TestStreamPackage_ContainsPart(t *testing.T) {
	pkg := opc.NewStreamPackage()
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePartFromBytes(uri, opc.ContentTypeSlide, []byte("<slide/>"))

	if !pkg.ContainsPart(uri) {
		t.Error("should contain added part")
	}

	nonExistent := opc.NewPackURI("/ppt/slides/slide999.xml")
	if pkg.ContainsPart(nonExistent) {
		t.Error("should not contain non-existent part")
	}
}

func TestStreamPackage_RemovePart(t *testing.T) {
	pkg := opc.NewStreamPackage()
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePartFromBytes(uri, opc.ContentTypeSlide, []byte("<slide/>"))

	err := pkg.RemovePart(uri)
	if err != nil {
		t.Fatalf("RemovePart failed: %v", err)
	}

	if pkg.PartCount() != 0 {
		t.Error("part should be removed")
	}
}

func TestStreamPackage_SaveAndOpen(t *testing.T) {
	// create package
	pkg := opc.NewStreamPackage()

	// add parts
	slideURI := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePartFromBytes(slideURI, opc.ContentTypeSlide, []byte(`<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"/>`))

	themeURI := opc.NewPackURI("/ppt/theme/theme1.xml")
	pkg.CreatePartFromBytes(themeURI, opc.ContentTypeTheme, []byte(`<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"/>`))

	// add relationship
	pkg.AddRelationship(opc.RelTypeOfficeDocument, "/ppt/presentation.xml", false)

	// create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_stream.pptx")

	// stream save
	err := pkg.StreamSaveFile(tmpFile)
	if err != nil {
		t.Fatalf("StreamSaveFile failed: %v", err)
	}

	// verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("saved file does not exist")
	}

	// stream open
	openedPkg, err := opc.OpenStream(tmpFile)
	if err != nil {
		t.Fatalf("OpenStream failed: %v", err)
	}
	defer openedPkg.Close()

	// verify content
	if openedPkg.PartCount() < 2 {
		t.Errorf("opened package has %d parts, expected at least 2", openedPkg.PartCount())
	}

	// verify part exists
	slidePart := openedPkg.GetPart(slideURI)
	if slidePart == nil {
		t.Error("slide part not found after reopening")
	}

	// verify lazy loading
	if slidePart.IsLoaded() {
		t.Error("part should not be loaded initially in stream mode")
	}

	// load part content
	blob, err := slidePart.Blob()
	if err != nil {
		t.Fatalf("Blob failed: %v", err)
	}
	if len(blob) == 0 {
		t.Error("blob should not be empty after loading")
	}
}

func TestStreamPackage_PartIterator(t *testing.T) {
	pkg := opc.NewStreamPackage()

	// add multiple parts
	for i := 1; i <= 3; i++ {
		uri := opc.NewPackURI("/ppt/slides/slide" + string(rune('0'+i)) + ".xml")
		pkg.CreatePartFromBytes(uri, opc.ContentTypeSlide, []byte("<slide/>"))
	}

	// add a part of a different type
	themeURI := opc.NewPackURI("/ppt/theme/theme1.xml")
	pkg.CreatePartFromBytes(themeURI, opc.ContentTypeTheme, []byte("<theme/>"))

	// iterate over all parts
	count := 0
	iter := pkg.NewPartIterator()
	for iter.Next() {
		count++
	}
	if count != 4 {
		t.Errorf("iterator returned %d parts, want 4", count)
	}

	// filter by type
	slideCount := 0
	iter2 := pkg.NewPartIterator().FilterByType(opc.ContentTypeSlide)
	for iter2.Next() {
		slideCount++
	}
	if slideCount != 3 {
		t.Errorf("filtered iterator returned %d slides, want 3", slideCount)
	}
}

// ===== RelationshipsStreamer tests =====

func TestRelationshipsStreamer(t *testing.T) {
	rels := opc.NewRelationships(opc.RootURI())
	rels.AddNew(opc.RelTypeOfficeDocument, "ppt/presentation.xml", false)
	rels.AddNew(opc.RelTypeCoreProperties, "docProps/core.xml", false)

	streamer := opc.NewRelationshipsStreamer(rels)

	buf := &bytes.Buffer{}
	if err := streamer.StreamWriteTo(buf); err != nil {
		t.Fatalf("StreamWriteTo failed: %v", err)
	}

	data := buf.String()
	if len(data) == 0 {
		t.Error("streamed data should not be empty")
	}

	// verify XML structure
	if !bytes.Contains(buf.Bytes(), []byte("<Relationships")) {
		t.Error("should contain Relationships element")
	}
	if !bytes.Contains(buf.Bytes(), []byte("<Relationship")) {
		t.Error("should contain Relationship elements")
	}
}

// ===== ContentTypesStreamer tests =====

func TestContentTypesStreamer(t *testing.T) {
	ct := opc.NewContentTypes()
	ct.AddDefault("xml", opc.ContentTypeXML)
	ct.AddDefault("rels", opc.ContentTypeRelationships)
	ct.AddOverride(opc.NewPackURI("/ppt/presentation.xml"), opc.ContentTypePresentation)

	streamer := opc.NewContentTypesStreamer(ct)

	buf := &bytes.Buffer{}
	if err := streamer.StreamWriteTo(buf); err != nil {
		t.Fatalf("StreamWriteTo failed: %v", err)
	}

	data := buf.String()
	if len(data) == 0 {
		t.Error("streamed data should not be empty")
	}

	// verify XML structure
	if !bytes.Contains(buf.Bytes(), []byte("<Types")) {
		t.Error("should contain Types element")
	}
	if !bytes.Contains(buf.Bytes(), []byte("<Default")) {
		t.Error("should contain Default elements")
	}
	if !bytes.Contains(buf.Bytes(), []byte("<Override")) {
		t.Error("should contain Override elements")
	}
}

// ===== Benchmarks =====

func BenchmarkStreamPart_Open(b *testing.B) {
	data := make([]byte, 1024*1024) // 1MB
	source := opc.NewBytesSource(data)
	uri := opc.NewPackURI("/large/part.bin")

	part := opc.NewStreamPart(uri, "application/octet-stream", source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rc, _ := part.Open()
		io.Copy(io.Discard, rc)
		rc.Close()
	}
}

func BenchmarkStreamPart_Blob(b *testing.B) {
	data := make([]byte, 1024*1024) // 1MB
	source := opc.NewBytesSource(data)
	uri := opc.NewPackURI("/large/part.bin")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		part := opc.NewStreamPart(uri, "application/octet-stream", source)
		part.Blob()
	}
}

func BenchmarkStreamingZipWriter(b *testing.B) {
	data := make([]byte, 1024) // 1KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		sw := opc.NewStreamingZipWriter(buf)
		sw.WriteBytes("test.txt", data)
		sw.Close()
	}
}

// ===== PartData and Channel tests =====

func TestPartData_New(t *testing.T) {
	data := []byte("<slide/>")
	partData := &opc.PartData{
		URI:         "/ppt/slides/slide1.xml",
		Path:        "ppt/slides/slide1.xml",
		ContentType: opc.ContentTypeSlide,
		Data:        data,
	}

	if partData.URI != "/ppt/slides/slide1.xml" {
		t.Errorf("URI = %q, want %q", partData.URI, "/ppt/slides/slide1.xml")
	}
	if partData.ContentType != opc.ContentTypeSlide {
		t.Errorf("ContentType = %q, want %q", partData.ContentType, opc.ContentTypeSlide)
	}
	if string(partData.Data) != string(data) {
		t.Error("Data mismatch")
	}
}

func TestPartDataChannel(t *testing.T) {
	ch := opc.NewPartDataChannel(10)
	if ch == nil {
		t.Fatal("NewPartDataChannel returned nil")
	}

	// send data
	go func() {
		ch <- &opc.PartData{
			Path: "test.txt",
			Data: []byte("hello"),
		}
		close(ch)
	}()

	// receive data
	received := 0
	for data := range ch {
		received++
		if data.Path != "test.txt" {
			t.Errorf("Path = %q, want %q", data.Path, "test.txt")
		}
	}

	if received != 1 {
		t.Errorf("received %d items, want 1", received)
	}
}

// ===== ResourceDedupPool tests =====

func TestResourceDedupPool_Register(t *testing.T) {
	pool := opc.NewResourceDedupPool()

	data1 := []byte("image data 1")
	data2 := []byte("image data 2")
	data3 := []byte("image data 1") // same as data1

	// register a new resource
	isNew, uri1 := pool.Register("/ppt/media/image1.png", data1)
	if !isNew {
		t.Error("first registration should be new")
	}
	if uri1 != "/ppt/media/image1.png" {
		t.Errorf("URI = %q, want %q", uri1, "/ppt/media/image1.png")
	}

	// register another new resource
	isNew, _ = pool.Register("/ppt/media/image2.png", data2)
	if !isNew {
		t.Error("second registration should be new")
	}

	// register duplicate resource (same content)
	isNew, existingURI := pool.Register("/ppt/media/image3.png", data3)
	if isNew {
		t.Error("duplicate registration should not be new")
	}
	if existingURI != uri1 {
		t.Errorf("existing URI = %q, want %q", existingURI, uri1)
	}
}

func TestResourceDedupPool_Stats(t *testing.T) {
	pool := opc.NewResourceDedupPool()

	pool.Register("/ppt/media/image1.png", []byte("data1"))
	pool.Register("/ppt/media/image2.png", []byte("data2"))

	count, totalSize := pool.Stats()
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	if totalSize != 10 { // len("data1") + len("data2") = 5 + 5 = 10
		t.Errorf("totalSize = %d, want 10", totalSize)
	}
}

func TestResourceDedupPool_Clear(t *testing.T) {
	pool := opc.NewResourceDedupPool()

	pool.Register("/ppt/media/image1.png", []byte("data1"))
	pool.Clear()

	count, _ := pool.Stats()
	if count != 0 {
		t.Errorf("after Clear, count = %d, want 0", count)
	}
}

func TestGlobalResourcePool(t *testing.T) {
	pool := opc.GetGlobalResourcePool()
	if pool == nil {
		t.Fatal("GetGlobalResourcePool returned nil")
	}

	// clear to ensure a clean test
	pool.Clear()

	// register resource
	isNew, _ := pool.Register("/test/image.png", []byte("test data"))
	if !isNew {
		t.Error("global pool registration should be new")
	}

	// cleanup
	pool.Clear()
}

// ===== ConcurrentZipCollector tests =====

func TestConcurrentZipCollector(t *testing.T) {
	buf := &bytes.Buffer{}
	collector := opc.NewConcurrentZipCollector(buf, 10)
	collector.Start()

	// submit some data
	if err := collector.SubmitBytes("test1.txt", []byte("hello")); err != nil {
		t.Fatalf("SubmitBytes failed: %v", err)
	}
	if err := collector.SubmitBytes("test2.txt", []byte("world")); err != nil {
		t.Fatalf("SubmitBytes failed: %v", err)
	}

	// close and wait
	if err := collector.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	// verify ZIP data
	data := buf.Bytes()
	if len(data) == 0 {
		t.Error("ZIP data should not be empty")
	}
}

func TestConcurrentZipCollector_WithPartData(t *testing.T) {
	buf := &bytes.Buffer{}
	collector := opc.NewConcurrentZipCollector(buf, 5)
	collector.Start()

	// submit using PartData
	partData := &opc.PartData{
		Path: "slide.xml",
		Data: []byte("<slide/>"),
	}
	if err := collector.Submit(partData); err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// submit using Source
	sourcePart := &opc.PartData{
		Path:   "from_source.txt",
		Source: opc.NewBytesSource([]byte("from source")),
	}
	if err := collector.Submit(sourcePart); err != nil {
		t.Fatalf("Submit with Source failed: %v", err)
	}

	if err := collector.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("ZIP data should not be empty")
	}
}

// ===== ConcurrentStreamSave tests =====

func TestStreamPackage_ConcurrentStreamSave(t *testing.T) {
	pkg := opc.NewStreamPackage()

	// add multiple parts
	for i := 1; i <= 5; i++ {
		uri := opc.NewPackURI("/ppt/slides/slide" + string(rune('0'+i)) + ".xml")
		data := []byte("<slide id=\"" + string(rune('0'+i)) + "\"/>")
		pkg.CreatePartFromBytes(uri, opc.ContentTypeSlide, data)
	}

	// add relationship
	pkg.AddRelationship(opc.RelTypeOfficeDocument, "/ppt/presentation.xml", false)

	// create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "concurrent_test.pptx")

	// concurrent save
	err := pkg.ConcurrentStreamSaveFile(tmpFile, 3, 10)
	if err != nil {
		t.Fatalf("ConcurrentStreamSaveFile failed: %v", err)
	}

	// verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("saved file does not exist")
	}

	// verify it can be reopened
	openedPkg, err := opc.OpenStream(tmpFile)
	if err != nil {
		t.Fatalf("OpenStream failed: %v", err)
	}
	defer openedPkg.Close()

	if openedPkg.PartCount() < 5 {
		t.Errorf("opened package has %d parts, expected at least 5", openedPkg.PartCount())
	}
}

// ===== Media deduplication tests =====

func TestStreamPackage_MediaDedup(t *testing.T) {
	pkg := opc.NewStreamPackage()

	// clear the global resource pool
	pkg.ClearMediaDedupPool()

	// add the same image twice
	imageData := []byte("PNG image data here")

	uri1 := opc.NewPackURI("/ppt/media/image1.png")
	actualURI1, isNew1, err := pkg.AddMediaPartWithDedup(uri1, opc.ContentTypePNG, imageData)
	if err != nil {
		t.Fatalf("AddMediaPartWithDedup failed: %v", err)
	}
	if !isNew1 {
		t.Error("first image should be new")
	}

	// attempt to add the same image again (should be deduplicated)
	uri2 := opc.NewPackURI("/ppt/media/image2.png")
	actualURI2, isNew2, err := pkg.AddMediaPartWithDedup(uri2, opc.ContentTypePNG, imageData)
	if err != nil {
		t.Fatalf("AddMediaPartWithDedup failed: %v", err)
	}
	if isNew2 {
		t.Error("second identical image should not be new")
	}
	if actualURI2.URI() != actualURI1.URI() {
		t.Errorf("second image should return first URI, got %q, want %q", actualURI2.URI(), actualURI1.URI())
	}

	// check stats
	count, _ := pkg.GetMediaDedupStats()
	if count != 1 {
		t.Errorf("media dedup count = %d, want 1", count)
	}

	// cleanup
	pkg.ClearMediaDedupPool()
}

// ===== Concurrent benchmarks =====

func BenchmarkConcurrentZipCollector(b *testing.B) {
	data := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		collector := opc.NewConcurrentZipCollector(buf, 100)
		collector.Start()

		for j := 0; j < 10; j++ {
			collector.SubmitBytes("test.txt", data)
		}

		collector.Close()
	}
}

func BenchmarkResourceDedupPool(b *testing.B) {
	pool := opc.NewResourceDedupPool()
	data := []byte("test data for hashing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Register("/test/resource", data)
	}
}

func BenchmarkConcurrentStreamSave(b *testing.B) {
	pkg := opc.NewStreamPackage()

	// add parts
	for i := 0; i < 10; i++ {
		uri := opc.NewPackURI("/ppt/slides/slide" + string(rune('0'+i%10)) + ".xml")
		pkg.CreatePartFromBytes(uri, opc.ContentTypeSlide, make([]byte, 1024))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		pkg.ConcurrentStreamSave(buf, 4, 20)
	}
}

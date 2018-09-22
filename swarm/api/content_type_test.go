package api

import (
	"os"
	"testing"
)

func TestDetectContentTypeOfNotExistFile(t *testing.T) {
	// internally use http.DetectContentType, so here are test cases only about fallback to file extension check
	testDetectCTNotExistFile(t, "./path/to/not/exist/file.pdf", "application/pdf")
	testDetectCTNotExistFile(t, "./path/to/not/exist/file.css", "text/css; charset=utf-8")
	testDetectCTNotExistFile(t, "./path/to/not/exist/.css", "text/css; charset=utf-8")
	testDetectCTNotExistFile(t, "./path/to/not/exist/file.md", "application/octet-stream")
	testDetectCTNotExistFile(t, "./path/to/not/exist/file.strangeext", "application/octet-stream")
	testDetectCTNotExistFile(t, "./path/to/not/exist/.gitignore", "application/octet-stream")
	testDetectCTNotExistFile(t, "./path/to/not/exist/file-no-extension", "application/octet-stream")
}

func testDetectCTNotExistFile(t *testing.T, path, expectedContentType string) {
	detected := DetectContentType(path)
	if detected != expectedContentType {
		t.Fatalf("Expected mime type %s, got %s", expectedContentType, detected)
	}
}

func TestDetectContentTypeOfExistFile(t *testing.T) {
	// internally use http.DetectContentType, so here are test cases only about fallback to file extension check
	testDetectCTCreateFile(t, "file.css", "Lorem Ipsum", "text/css; charset=utf-8")
}

func testDetectCTCreateFile(t *testing.T, fileName, content, expectedContentType string) {
	path := os.TempDir() + fileName
	_, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	detected := DetectContentType(path)
	if detected != expectedContentType {
		t.Fatalf("Expected mime type %s, got %s", expectedContentType, detected)
	}
}

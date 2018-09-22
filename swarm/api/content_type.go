package api

import (
	"github.com/ethereum/go-ethereum/log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

// detect content type by file content
// otherwise by file extension
// returns "application/octet-stream" in worst case
func DetectContentType(file string) string {
	var contentType = "application/octet-stream" // default value

	f, err := os.Open(file)
	if err != nil {
		log.Warn("detectMimeType: can't open file", "file", file, "err", err)
	}
	defer f.Close()
	buf := make([]byte, 512)
	if n, _ := f.Read(buf); n > 0 {
		contentType = http.DetectContentType(buf)
	}

	// if found specific contentType - return it, else check file extension
	if contentType != "" && contentType != "application/octet-stream" {
		return contentType
	}

	if ext := filepath.Ext(file); ext != "" {
		contentType = mime.TypeByExtension(ext)
	}

	if contentType == "" {
		return "application/octet-stream"
	}

	return contentType
}

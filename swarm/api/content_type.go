package api

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/log"
)

// detect first by file extension
// otherwise detect by file content
// if it cannot determine a more specific one, it
// returns "application/octet-stream".
func DetectContentType(file string) string {
	if ext := filepath.Ext(file); ext != "" {
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			return mimeType
		}
	}

	f, err := os.Open(file)
	if err != nil {
		log.Warn("detectMimeType: can't open file", "file", file, "err", err)
		return "application/octet-stream"
	}
	defer f.Close()
	buf := make([]byte, 512)
	if n, _ := f.Read(buf); n > 0 {
		return http.DetectContentType(buf)
	}
	return "application/octet-stream"
}

package api

import "testing"

func TestDetectContentType(t *testing.T) {
	for _, v := range []struct {
		path                string
		expectedContentType string
	}{
		{
			path:                "./path/to/file.pdf",
			expectedContentType: "application/pdf",
		},
		{
			path:                "./path/to/file.md",
			expectedContentType: "application/octet-stream",
		},
		{
			path:                "",
			expectedContentType: "application/octet-stream",
		},
		{
			path:                "noextension",
			expectedContentType: "application/octet-stream",
		},
		{
			path:                "./path/to/noextension",
			expectedContentType: "application/octet-stream",
		},
		{
			path:                "./1.css",
			expectedContentType: "text/css; charset=utf-8",
		},
	} {
		detected := DetectContentType(v.path)
		if detected != v.expectedContentType {
			t.Fatalf("Expected mime type %s, got %s", v.expectedContentType, detected)
		}
	}
}

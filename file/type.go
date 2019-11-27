package file

import "hash"

// SectionWriter is a chainable interface for file-based operations in swarm
type SectionWriter interface {
	hash.Hash
}

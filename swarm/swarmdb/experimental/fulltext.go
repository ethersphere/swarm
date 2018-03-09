package swarmdb

import ()

// FullTextIndex is a B+tree.
type FullTextIndex struct {
	Tree
}

func (t *FullTextIndex) GetDocs(words []string) (docs [][]byte, err error) {
	return docs, err
}

func (t *FullTextIndex) Put(key []byte /*K*/, v []byte /*V*/) (okresult bool, err error) {
	return okresult, err
}

func (t *FullTextIndex) StartBuffer() (ok bool, err error) {
	return ok, err
}

func (t *FullTextIndex) FlushBuffer() (ok bool, err error) {
	return ok, err
}

func NewFullTextIndex(swarmdb SwarmDB, hashid []byte) *FullTextIndex {
	var t *FullTextIndex
	return t
}

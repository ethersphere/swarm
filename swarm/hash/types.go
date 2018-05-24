package hash

import (
	"hash"
	"sync"
)

var (
	defaultHash string
	hashers     map[string]*Hasher
)

type SwarmHash interface {
	hash.Hash
	ResetWithLength([]byte)
}

type SwarmHasher func() SwarmHash

type Hasher struct {
	typ      string
	hashFunc func() SwarmHash
	pool     sync.Pool
	size     int
}

func init() {
	hashers = make(map[string]*Hasher)
}

func SetDefaultHash(typ string) bool {
	if _, ok := hashers[typ]; !ok {
		return false
	}
	defaultHash = typ
	return true
}

func GetDefaultHashLength() int {
	if _, ok := hashers[defaultHash]; !ok {
		return 0
	}
	return hashers[defaultHash].Size()
}

func AddHasher(typ string, hasherCount int, hashFunc func(string) SwarmHasher) {
	h := &Hasher{
		typ: typ,
		pool: sync.Pool{
			New: func() interface{} {
				return hashFunc(typ)()
			},
		},
	}
	for i := 0; i < hasherCount; i++ {
		hf := hashFunc(typ)()
		if h.size == 0 {
			h.size = hf.Size()
		}
		h.pool.Put(hf)
	}
	hashers[typ] = h
}

func GetDefaultHasher() *Hasher {
	return hashers[defaultHash]
}

func GetHasher(typ string) *Hasher {
	return hashers[typ]
}

func (h *Hasher) Size() int {
	return h.size
}

func (h *Hasher) Hash(data []byte) []byte {
	hasher := h.pool.Get().(SwarmHash)
	hasher.ResetWithLength(data)
	hasher.Write(data)
	return hasher.Sum(nil)
}

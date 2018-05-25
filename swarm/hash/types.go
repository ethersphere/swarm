package swarmhash

import (
	"fmt"
	"hash"
	"sync"
)

const (
	defaultHashName    = "BMT"
	DefaultWorkerCount = 32
)

var (
	defaultHash string
	pools       map[string]*Hasher
	initFunc    sync.Once
)

// the basic swarm hash interface
// an instance represents an individual worker in a pool
type Hash interface {
	hash.Hash
	ResetWithLength([]byte)
}

// function for creating a new individual worker
type HashFunc func() Hash

// Hasher maintains a pool of workers
// It is the entrypoint for invoking hashing jobs
// And to query metadata about the type of worker in its pool
type Hasher struct {
	typ      string
	hashFunc func() Hash
	pool     sync.Pool
	size     int
}

func init() {
	pools = make(map[string]*Hasher)
}

// SetDefaultHash sets the default hash to be used when calling GetHash
func Init(typ string) {
	initFunc.Do(func() {
		if defaultHash != "" {
			panic("cannot change default hash after initialization")
		} else if _, ok := pools[typ]; !ok {
			panic(fmt.Sprintf("hash %s not registered", typ))
		}
		defaultHash = typ
	})
}

// New creates a new hasher pool
// The identifier of the hasher pool is the first parameter
// hasherCount indicates the number of workers to maintain in the pool
// hashFunc is the function to be used to spawn workers
func Add(typ string, hasherCount int, hashFunc HashFunc) error {
	if defaultHash != "" {
		panic("cannot add hash after initialization")
	} else if _, ok := pools[typ]; ok {
		panic(fmt.Sprintf("hash %s already registered", typ))
	}
	h := &Hasher{
		typ: typ,
		pool: sync.Pool{
			New: func() interface{} {
				return hashFunc()
			},
		},
	}
	for i := 0; i < hasherCount; i++ {
		hf := hashFunc()
		if h.size == 0 {
			h.size = hf.Size()
		}
		h.pool.Put(hf)
	}
	pools[typ] = h
	return nil
}

// GetHash returns the hasher pool for a specified hash type
func GetHash() *Hasher {
	return pools[defaultHash]
}

// GetHashByName returns the hasher pool identified by typ
// If the hash entry does not exist in the pool, nil is returned
func GetHashByName(typ string) *Hasher {
	return pools[typ]
}

// GetHashLength returns the digest size of the hash set as default
// It returns 0 if the default hash is not set
func GetHashSize() int {
	return getHashLength(defaultHash)
}

// GetHashLengthByName returns the digets size of the hash identified by typ
// It returns 0 if the hash pool identifier doesn't exist
func GetHashSizeByName(typ string) int {
	return getHashLength(typ)
}

func getHashLength(typ string) int {
	if _, ok := pools[typ]; !ok {
		return 0
	}
	return pools[typ].Size()
}

// Returns the digest size the hashers of the hash pool will produce
func (h *Hasher) Size() int {
	return h.size
}

// Performs one hash job
func (h *Hasher) Hash(data ...[]byte) []byte {
	hasher := h.pool.Get().(Hash)
	defer h.pool.Put(hasher)
	hasher.Reset()
	return h.doHash(hasher, data)
}

func (h *Hasher) HashWithLength(length []byte, data ...[]byte) []byte {
	hasher := h.pool.Get().(Hash)
	defer h.pool.Put(hasher)
	hasher.ResetWithLength(length)
	return h.doHash(hasher, data)
}

func (h *Hasher) doHash(hasher Hash, data [][]byte) []byte {
	for _, d := range data {
		hasher.Write(d)
	}
	return hasher.Sum(nil)
}

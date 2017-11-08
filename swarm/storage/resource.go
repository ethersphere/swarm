package storage

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/idna"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/ens"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

/***
*
* Resource updates is a data update scheme built on swarm chunks
* with chunk keys following a predictable, versionable pattern.
*
* The intended use case is to update data locations of certain
* hashes
*
* Updates are defined to be periodic in nature, where periods are
* expressed in terms of number of blocks.
*
* The root entry of a resource update is tied to a unique identifier,
* typically - but not necessarily - an ens name. It also contains the
* block number when the resource update was first registered, and
* the block frequency with which the resource will be updated.
* Thus, a resource update for identifier "foo.bar" starting at block 4200
* with frequency 42 will have updates on block 4242, 4284, 4326 and so on.
*
* the identifier is supplied as a string, but will be IDNA converted and
* passed through the ENS namehash function. Pure ascii identifiers without
* periods will thus merely be hashed.
*
* The data format of the root entry is the ENS namehash as key and
* blocknumber and frequency in little-endian uint64 representation as values,
* concatenated in that order, for a total of 16 bytes.
*
* Note that the root entry is not required for the resource update scheme to
* work. A normal chunk of the blocknumber/frequency data can also be created,
* and pointed to by an actual ENS entry instead.
*
* Actual data updates are also made in the form of swarm chunks. The keys
* of the updates are the hash of a concatenation of properties as follows:
*
* sha256(namehash|blocknumber|version)
*
* The blocknumber here is the next block period after the current block
* calculated from the start block and frequency of the resource update.
* Using our previous example, this means that an update made at block 4285,
* and even 4284, will have 4326 as the block number.
*
* If more than one update is made to the same block number, incremental
* version numbers are used successively.
*
* A lookup agent need only know the identifier name
*
* NOTE: the following is yet to be implemented
* The resource update chunks will be stored in the swarm, but receive special
* treatment as their keys do not validate as hashes of their data. They are also
* stored using a separate store, and forwarding/syncing protocols carry per-chunk
* flags to tell whether the chunk can be validated or not; if not it is to be
* treated as a resource update chunk.
*
* TODO: signature and signature validation
 */

// Encapsulates an actual resource update. When synced it contains the most recent
// version of the resource update data.
type resource struct {
	name       string
	ensname    common.Hash
	startblock uint64
	lastblock  uint64
	frequency  uint64
	version    uint64
	data       []byte
	updated    time.Time
}

// Main interface to resource updates. Creates/opens the resource database,
// and sets up rpc client to the blockchain api
type ResourceHandler struct {
	ChunkStore
	ethapi       *rpc.Client
	resources    map[string]*resource
	hashLock     *sync.Mutex
	resourceLock *sync.Mutex
	hasher       SwarmHash
}

// Create or open resource update chunk store
func NewResourceHandler(datadir string, cloudStore CloudStore, ethapi *rpc.Client) (*ResourceHandler, error) {
	path := filepath.Join(datadir, "resource")
	dbStore, err := NewDbStore(datadir, nil, singletonSwarmDbCapacity, 0)
	if err != nil {
		return nil, err
	}
	localStore := &LocalStore{
		memStore: NewMemStore(dbStore, singletonSwarmDbCapacity),
		DbStore:  dbStore,
	}
	hasher := MakeHashFunc("SHA256")
	return &ResourceHandler{
		ChunkStore:   newResourceChunkStore(path, hasher, localStore, cloudStore),
		ethapi:       ethapi,
		resources:    make(map[string]*resource),
		resourceLock: &sync.Mutex{},
		hashLock:     &sync.Mutex{},
		hasher:       hasher(),
	}, nil
}

// Creates a standalone resource object
//
// Can be passed to SetResource if external root data lookups are used
func NewResource(name string, startblock uint64, frequency uint64) (*resource, error) {
	validname, err := idna.ToASCII(name)
	if err != nil {
		return nil, err
	}

	return &resource{
		name:       validname,
		ensname:    ens.EnsNode(validname),
		startblock: startblock,
		frequency:  frequency,
	}, nil
}

// Creates a new root entry for a resource update identified by `name` with the specified `frequency`.
//
// The start block of the resource update will be the actual current block height of the connected network.
func (self *ResourceHandler) NewResource(name string, frequency uint64) (*resource, error) {

	// make sure our ens identifier is idna safe
	validname, err := idna.ToASCII(name)
	if err != nil {
		return nil, err
	}
	ensname := ens.EnsNode(validname)

	// get our blockheight at this time
	currentblock, err := self.getBlock()
	if err != nil {
		return nil, err
	}

	// chunk with key equal to namehash points to data of first blockheight + update frequency
	// from this we know from what blockheight we should look for updates, and how often
	chunk := NewChunk(Key(ensname[:]), nil)
	chunk.SData = make([]byte, 24)

	// resource update root chunks follow same convention as "normal" chunks
	// with 8 bytes prefix specifying size
	val := make([]byte, 8)
	chunk.SData[0] = 16 // size, little-endian
	binary.LittleEndian.PutUint64(val, currentblock)
	copy(chunk.SData[8:16], val)
	binary.LittleEndian.PutUint64(val, frequency)
	copy(chunk.SData[16:], val)
	self.Put(chunk)
	log.Debug("new resource", "name", validname, "key", ensname, "startblock", currentblock, "frequency", frequency)

	self.resourceLock.Lock()
	defer self.resourceLock.Unlock()
	self.resources[name] = &resource{
		name:       validname,
		ensname:    ensname,
		startblock: currentblock,
		frequency:  frequency,
		updated:    time.Now(),
	}
	return self.resources[name], nil
}

// Set an externally defined resource object
//
// If the resource update root chunk is located externally (for example as a normal
// chunk looked up by ENS) the data would be manually added with this method).
//
// Method will fail if resource is already registered in this session, unless
// `allowOverwrite` is set
func (self *ResourceHandler) SetResource(rsrc *resource, allowOverwrite bool) error {

	if rsrc.name == "" {
		return fmt.Errorf("Resource name cannot be empty")
	}

	utfname, err := idna.ToUnicode(rsrc.name)
	if err != nil {
		return fmt.Errorf("Invalid IDNA rsrc name '%s'", rsrc.name)
	}
	if !allowOverwrite {
		if _, ok := self.resources[utfname]; ok {
			return fmt.Errorf("Resource exists")
		}
	}

	// get our blockheight at this time
	currentblock, err := self.getBlock()
	if err != nil {
		return err
	}

	if rsrc.startblock > currentblock {
		return fmt.Errorf("Startblock cannot be higher than current block (%d > %d)", rsrc.startblock, currentblock)
	}

	if rsrc.frequency == 0 {
		return fmt.Errorf("Frequency cannot be 0")
	}

	if len(rsrc.ensname) > 0 {
		ensname := ens.EnsNode(rsrc.name)
		if ensname != rsrc.ensname {
			return fmt.Errorf("Namehash %x is not a valid namehash of IDNA name '%s'", rsrc.ensname, rsrc.name)
		}
	}
	self.resources[utfname] = rsrc
	return nil
}

// Searches and retrieves the last version of the resource update identified by `name`
//
// It starts at the next period after the current block height, and upon failure
// tries the corresponding keys of each previous period until one is found
// (or startblock is reached, in which case there are no updates).
// If an update is found, version numbers are iterated until failure, and the last
// successfully retrieved version is copied to the corresponding resources map entry
// and returned.
//
// If refresh is set to true, the resource data will be reloaded from the resource update
// root chunk.
// It is the callers responsibility to make sure that this chunk exists (if the resource
// update root data was retrieved externally, it typically doesn't)
func (self *ResourceHandler) OpenResource(name string, refresh bool) (*resource, error) {

	rsrc := &resource{}

	// if the resource is not known to this session we must load it
	// if refresh is set, we force load
	if _, ok := self.resources[name]; !ok || refresh {

		// make sure our ens identifier is idna safe
		validname, err := idna.ToASCII(name)
		if err != nil {
			return nil, err
		}
		rsrc.name = validname
		rsrc.ensname = ens.EnsNode(validname)

		// get the root info chunk and update the cached value
		chunk, err := self.Get(Key(rsrc.ensname[:]))
		if err != nil {
			return nil, err
		}

		// sanity check for chunk data
		// data is prefixed by 8 bytes of size
		if len(chunk.SData) < 24 {
			return nil, fmt.Errorf("Invalid chunk length %d", len(chunk.SData))
		} else {
			chunklength := binary.LittleEndian.Uint64(chunk.SData[:8])
			if chunklength != uint64(16) {
				return nil, fmt.Errorf("Invalid chunk length header %d", chunklength)
			}
		}
		rsrc.startblock = binary.LittleEndian.Uint64(chunk.SData[8:16])
		rsrc.frequency = binary.LittleEndian.Uint64(chunk.SData[16:])
	} else {
		rsrc.name = self.resources[name].name
		rsrc.ensname = self.resources[name].ensname
		rsrc.startblock = self.resources[name].startblock
		rsrc.frequency = self.resources[name].frequency
	}

	// get our blockheight at this time and the next block of the update period
	currentblock, err := self.getBlock()
	if err != nil {
		return nil, err
	}
	nextblock := getNextBlock(rsrc.startblock, currentblock, rsrc.frequency)

	// start from the last possible block period, and iterate previous ones until we find a match
	// if we hit startblock we're out of options
	version := uint64(1)
	for nextblock > rsrc.startblock {
		key := self.resourceHash(rsrc.ensname, nextblock, version)
		chunk, err := self.Get(key)
		if err == nil {
			// check if we have versions > 1. If a version fails, the previous version is used and returned.
			log.Trace("rsrc update version 1 found, checking for version updates", "nextblock", nextblock, "key", key)
			for {
				newversion := version + 1
				key := self.resourceHash(rsrc.ensname, nextblock, newversion)
				newchunk, err := self.Get(key)
				if err != nil {
					// rsrc update data chunks are total hacks
					// and have no size prefix :D
					if len(chunk.SData) == 0 {
						return nil, fmt.Errorf("Update contains no data")
					}
					// update our rsrcs entry map
					rsrc.lastblock = nextblock
					rsrc.version = version
					rsrc.data = make([]byte, len(chunk.SData))
					rsrc.updated = time.Now()
					copy(rsrc.data, chunk.SData)
					log.Debug("Resource synced", "name", rsrc.name, "key", chunk.Key, "block", nextblock, "version", version)
					self.resourceLock.Lock()
					self.resources[name] = rsrc
					self.resourceLock.Unlock()
					return rsrc, nil
				}
				log.Trace("version update found, checking next", "version", version, "block", nextblock, "key", key)
				chunk = newchunk
				version = newversion
			}
		}
		log.Trace("rsrc update not found, checking previous period", "block", nextblock, "key", key)
		nextblock -= rsrc.frequency
	}
	return nil, fmt.Errorf("no updates found")
}

// Adds an actual data update
//
// Uses the data currently loaded in the resources map entry.
// It is the caller's responsibility to make sure that this data is not stale.
//
// A resource update cannot span chunks, and thus has max length 4096
func (self *ResourceHandler) Update(name string, data []byte) (Key, error) {

	// can be only one chunk long
	if len(data) > 4096 {
		return nil, fmt.Errorf("Data overflow: %d / 4096 bytes", len(data))
	}

	// get the cached information
	self.resourceLock.Lock()
	defer self.resourceLock.Unlock()
	resource, ok := self.resources[name]
	if !ok {
		return nil, fmt.Errorf("No such resource")
	} else if resource.updated.IsZero() {
		return nil, fmt.Errorf("Invalid resource")
	}

	// get our blockheight at this time and the next block of the update period
	currentblock, err := self.getBlock()
	if err != nil {
		return nil, err
	}
	nextblock := getNextBlock(resource.startblock, currentblock, resource.frequency)

	// if we already have an update for this block then increment version
	var version uint64
	if nextblock == resource.lastblock {
		version = resource.version
	}
	version++

	// create the update chunk and send it
	key := self.resourceHash(resource.ensname, nextblock, version)
	chunk := NewChunk(key, nil)
	chunk.SData = data
	chunk.Size = int64(len(data))
	self.Put(chunk)
	log.Trace("resource update", "name", resource.name, "key", key, "currentblock", currentblock, "lastblock", nextblock, "version", version)

	// update our resources map entry and return the new key
	resource.lastblock = nextblock
	resource.version = version
	resource.data = make([]byte, len(data))
	copy(resource.data, data)
	return key, nil
}

// Closes the datastore.
// Always call this at shutdown to avoid data corruption.
func (self *ResourceHandler) Close() {
	self.ChunkStore.Close()
}

func (self *ResourceHandler) getBlock() (uint64, error) {
	// get the block height and convert to uint64
	var currentblock string
	err := self.ethapi.Call(&currentblock, "eth_blockNumber")
	if err != nil {
		return 0, err
	}
	if currentblock == "0x0" {
		return 0, nil
	}
	return strconv.ParseUint(currentblock, 10, 64)
}

func (self *ResourceHandler) resourceHash(namehash common.Hash, blockheight uint64, version uint64) Key {
	// format is: hash(namehash|blockheight|version)
	self.hashLock.Lock()
	defer self.hashLock.Unlock()
	self.hasher.Reset()
	self.hasher.Write(namehash[:])
	b := make([]byte, 8)
	c := binary.PutUvarint(b, blockheight)
	self.hasher.Write(b)
	// PutUvarint only overwrites first c bytes
	for i := 0; i < c; i++ {
		b[i] = 0
	}
	c = binary.PutUvarint(b, version)
	self.hasher.Write(b)
	return self.hasher.Sum(nil)
}

type resourceChunkStore struct {
	localStore ChunkStore
	netStore   ChunkStore
}

func newResourceChunkStore(path string, hasher SwarmHasher, localStore *LocalStore, cloudStore CloudStore) *resourceChunkStore {
	return &resourceChunkStore{
		localStore: localStore,
		netStore:   NewNetStore(hasher, localStore, cloudStore, NewStoreParams(path)),
	}
}

func (r *resourceChunkStore) Get(key Key) (*Chunk, error) {
	chunk, err := r.netStore.Get(key)
	if err != nil {
		return nil, err
	}
	// if the chunk has to be remotely retrieved, we define a timeout of how long to wait for it before failing.
	// sadly due to the nature of swarm, the error will never be conclusive as to whether it was a network issue
	// that caused the failure or that the chunk doesn't exist.
	if chunk.Req == nil {
		return chunk, nil
	}
	t := time.NewTimer(time.Second * 1)
	select {
	case <-t.C:
		return nil, fmt.Errorf("timeout")
	case <-chunk.C:
		log.Trace("Received resource update chunk", "peer", chunk.Req.Source)
	}
	return chunk, nil
}

func (r *resourceChunkStore) Put(chunk *Chunk) {
	r.netStore.Put(chunk)
}

func (r *resourceChunkStore) Close() {
	r.netStore.Close()
	r.localStore.Close()
}

func getNextBlock(start uint64, current uint64, frequency uint64) uint64 {
	blockdiff := current - start
	periods := (blockdiff / frequency) + 1
	return start + (frequency * periods)
}

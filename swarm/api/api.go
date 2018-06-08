// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"path"
	"strings"

	"bytes"
	"mime"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/ens"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/multihash"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/ethereum/go-ethereum/swarm/storage/mru"
)

type ErrResourceReturn struct {
	key string
}

func (e *ErrResourceReturn) Error() string {
	return "resourceupdate"
}

func (e *ErrResourceReturn) Key() string {
	return e.key
}

var (
	apiResolveCount    = metrics.NewRegisteredCounter("api.resolve.count", nil)
	apiResolveFail     = metrics.NewRegisteredCounter("api.resolve.fail", nil)
	apiPutCount        = metrics.NewRegisteredCounter("api.put.count", nil)
	apiPutFail         = metrics.NewRegisteredCounter("api.put.fail", nil)
	apiGetCount        = metrics.NewRegisteredCounter("api.get.count", nil)
	apiGetNotFound     = metrics.NewRegisteredCounter("api.get.notfound", nil)
	apiGetHttp300      = metrics.NewRegisteredCounter("api.get.http.300", nil)
	apiModifyCount     = metrics.NewRegisteredCounter("api.modify.count", nil)
	apiModifyFail      = metrics.NewRegisteredCounter("api.modify.fail", nil)
	apiAddFileCount    = metrics.NewRegisteredCounter("api.addfile.count", nil)
	apiAddFileFail     = metrics.NewRegisteredCounter("api.addfile.fail", nil)
	apiRmFileCount     = metrics.NewRegisteredCounter("api.removefile.count", nil)
	apiRmFileFail      = metrics.NewRegisteredCounter("api.removefile.fail", nil)
	apiAppendFileCount = metrics.NewRegisteredCounter("api.appendfile.count", nil)
	apiAppendFileFail  = metrics.NewRegisteredCounter("api.appendfile.fail", nil)
	apiGetInvalid      = metrics.NewRegisteredCounter("api.get.invalid", nil)
)

type Resolver interface {
	Resolve(string) (common.Hash, error)
}

type ResolveValidator interface {
	Resolver
	Owner(node [32]byte) (common.Address, error)
	HeaderByNumber(context.Context, *big.Int) (*types.Header, error)
}

// NoResolverError is returned by MultiResolver.Resolve if no resolver
// can be found for the address.
type NoResolverError struct {
	TLD string
}

func NewNoResolverError(tld string) *NoResolverError {
	return &NoResolverError{TLD: tld}
}

func (e *NoResolverError) Error() string {
	if e.TLD == "" {
		return "no ENS resolver"
	}
	return fmt.Sprintf("no ENS endpoint configured to resolve .%s TLD names", e.TLD)
}

// MultiResolver is used to resolve URL addresses based on their TLDs.
// Each TLD can have multiple resolvers, and the resoluton from the
// first one in the sequence will be returned.
type MultiResolver struct {
	resolvers map[string][]ResolveValidator
	nameHash  func(string) common.Hash
}

// MultiResolverOption sets options for MultiResolver and is used as
// arguments for its constructor.
type MultiResolverOption func(*MultiResolver)

// MultiResolverOptionWithResolver adds a Resolver to a list of resolvers
// for a specific TLD. If TLD is an empty string, the resolver will be added
// to the list of default resolver, the ones that will be used for resolution
// of addresses which do not have their TLD resolver specified.
func MultiResolverOptionWithResolver(r ResolveValidator, tld string) MultiResolverOption {
	return func(m *MultiResolver) {
		m.resolvers[tld] = append(m.resolvers[tld], r)
	}
}

func MultiResolverOptionWithNameHash(nameHash func(string) common.Hash) MultiResolverOption {
	return func(m *MultiResolver) {
		m.nameHash = nameHash
	}
}

// NewMultiResolver creates a new instance of MultiResolver.
func NewMultiResolver(opts ...MultiResolverOption) (m *MultiResolver) {
	m = &MultiResolver{
		resolvers: make(map[string][]ResolveValidator),
		nameHash:  ens.EnsNode,
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

// Resolve resolves address by choosing a Resolver by TLD.
// If there are more default Resolvers, or for a specific TLD,
// the Hash from the the first one which does not return error
// will be returned.
func (m *MultiResolver) Resolve(addr string) (h common.Hash, err error) {
	rs, err := m.getResolveValidator(addr)
	if err != nil {
		return h, err
	}
	for _, r := range rs {
		h, err = r.Resolve(addr)
		if err == nil {
			return
		}
	}
	return
}

func (m *MultiResolver) ValidateOwner(name string, address common.Address) (bool, error) {
	rs, err := m.getResolveValidator(name)
	if err != nil {
		return false, err
	}
	var addr common.Address
	for _, r := range rs {
		addr, err = r.Owner(m.nameHash(name))
		// we hide the error if it is not for the last resolver we check
		if err == nil {
			return addr == address, nil
		}
	}
	return false, err
}

func (m *MultiResolver) HeaderByNumber(ctx context.Context, name string, blockNr *big.Int) (*types.Header, error) {
	rs, err := m.getResolveValidator(name)
	if err != nil {
		return nil, err
	}
	for _, r := range rs {
		var header *types.Header
		header, err = r.HeaderByNumber(ctx, blockNr)
		// we hide the error if it is not for the last resolver we check
		if err == nil {
			return header, nil
		}
	}
	return nil, err
}

func (m *MultiResolver) getResolveValidator(name string) ([]ResolveValidator, error) {
	rs := m.resolvers[""]
	tld := path.Ext(name)
	if tld != "" {
		tld = tld[1:]
		rstld, ok := m.resolvers[tld]
		if ok {
			return rstld, nil
		}
	}
	if len(rs) == 0 {
		return rs, NewNoResolverError(tld)
	}
	return rs, nil
}

func (m *MultiResolver) SetNameHash(nameHash func(string) common.Hash) {
	m.nameHash = nameHash
}

/*
Api implements webserver/file system related content storage and retrieval
on top of the FileStore
it is the public interface of the FileStore which is included in the ethereum stack
*/
type Api struct {
	resource  *mru.Handler
	fileStore *storage.FileStore
	dns       Resolver
}

//the api constructor initialises
func NewApi(fileStore *storage.FileStore, dns Resolver, resourceHandler *mru.Handler) (self *Api) {
	self = &Api{
		fileStore: fileStore,
		dns:       dns,
		resource:  resourceHandler,
	}
	return
}

// to be used only in TEST
func (self *Api) Upload(uploadDir, index string, toEncrypt bool) (hash string, err error) {
	fs := NewFileSystem(self)
	hash, err = fs.Upload(uploadDir, index, toEncrypt)
	return hash, err
}

// FileStore reader API
func (self *Api) Retrieve(addr storage.Address) (reader storage.LazySectionReader, isEncrypted bool) {
	return self.fileStore.Retrieve(addr)
}

func (self *Api) Store(data io.Reader, size int64, toEncrypt bool) (addr storage.Address, wait func(), err error) {
	log.Debug("api.store", "size", size)
	return self.fileStore.Store(data, size, toEncrypt)
}

type ErrResolve error

// DNS Resolver
func (self *Api) Resolve(uri *URI) (storage.Address, error) {
	apiResolveCount.Inc(1)
	log.Trace("resolving", "uri", uri.Addr)

	// if the URI is immutable, check if the address looks like a hash
	if uri.Immutable() {
		key := uri.Address()
		if key == nil {
			return nil, fmt.Errorf("immutable address not a content hash: %q", uri.Addr)
		}
		return key, nil
	}

	// if DNS is not configured, check if the address is a hash
	if self.dns == nil {
		key := uri.Address()
		if key == nil {
			apiResolveFail.Inc(1)
			return nil, fmt.Errorf("no DNS to resolve name: %q", uri.Addr)
		}
		return key, nil
	}

	// try and resolve the address
	resolved, err := self.dns.Resolve(uri.Addr)
	if err == nil {
		return resolved[:], nil
	}

	key := uri.Address()
	if key == nil {
		apiResolveFail.Inc(1)
		return nil, err
	}
	return key, nil
}

// Put provides singleton manifest creation on top of FileStore store
func (self *Api) Put(content, contentType string, toEncrypt bool) (k storage.Address, wait func(), err error) {
	apiPutCount.Inc(1)
	r := strings.NewReader(content)
	key, waitContent, err := self.fileStore.Store(r, int64(len(content)), toEncrypt)
	if err != nil {
		apiPutFail.Inc(1)
		return nil, nil, err
	}
	manifest := fmt.Sprintf(`{"entries":[{"hash":"%v","contentType":"%s"}]}`, key, contentType)
	r = strings.NewReader(manifest)
	key, waitManifest, err := self.fileStore.Store(r, int64(len(manifest)), toEncrypt)
	if err != nil {
		apiPutFail.Inc(1)
		return nil, nil, err
	}
	return key, func() {
		waitContent()
		waitManifest()
	}, nil
}

// Get uses iterative manifest retrieval and prefix matching
// to resolve basePath to content using FileStore retrieve
// it returns a section reader, mimeType, status, the key of the actual content and an error
func (self *Api) Get(manifestAddr storage.Address, path string) (reader storage.LazySectionReader, mimeType string, status int, contentAddr storage.Address, err error) {
	log.Debug("api.get", "key", manifestAddr, "path", path)
	apiGetCount.Inc(1)
	trie, err := loadManifest(self.fileStore, manifestAddr, nil)
	if err != nil {
		apiGetNotFound.Inc(1)
		status = http.StatusNotFound
		log.Warn(fmt.Sprintf("loadManifestTrie error: %v", err))
		return
	}

	log.Debug("trie getting entry", "key", manifestAddr, "path", path)
	entry, _ := trie.getEntry(path)

	if entry != nil {
		log.Debug("trie got entry", "key", manifestAddr, "path", path, "entry.Hash", entry.Hash)
		// we need to do some extra work if this is a mutable resource manifest
		if entry.ContentType == ResourceContentType {

			// get the resource root chunk key
			log.Trace("resource type", "key", manifestAddr, "hash", entry.Hash)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			rsrc, err := self.resource.Load(storage.Address(common.FromHex(entry.Hash)))
			if err != nil {
				apiGetNotFound.Inc(1)
				status = http.StatusNotFound
				log.Debug(fmt.Sprintf("get resource content error: %v", err))
				return reader, mimeType, status, nil, err
			}

			// use this key to retrieve the latest update
			rsrc, err = self.resource.LookupLatest(ctx, rsrc.NameHash(), true, &mru.LookupParams{})
			if err != nil {
				apiGetNotFound.Inc(1)
				status = http.StatusNotFound
				log.Debug(fmt.Sprintf("get resource content error: %v", err))
				return reader, mimeType, status, nil, err
			}

			// if it's multihash, we will transparently serve the content this multihash points to
			// \TODO this resolve is rather expensive all in all, review to see if it can be achieved cheaper
			if rsrc.Multihash {

				// get the data of the update
				_, rsrcData, err := self.resource.GetContent(rsrc.NameHash().Hex())
				if err != nil {
					apiGetNotFound.Inc(1)
					status = http.StatusNotFound
					log.Warn(fmt.Sprintf("get resource content error: %v", err))
					return reader, mimeType, status, nil, err
				}

				// validate that data as multihash
				decodedMultihash, err := multihash.Decode(rsrcData)
				if err != nil {
					apiGetInvalid.Inc(1)
					status = http.StatusInternalServerError
					log.Warn(fmt.Sprintf("could not decode resource multihash: %v", err))
					return reader, mimeType, status, nil, err
				} else if decodedMultihash.Code != multihash.KECCAK_256 {
					apiGetInvalid.Inc(1)
					status = http.StatusUnprocessableEntity
					log.Warn(fmt.Sprintf("invalid resource multihash code: %x", decodedMultihash.Code))
					return reader, mimeType, status, nil, err
				}
				manifestAddr = storage.Address(decodedMultihash.Digest)
				log.Trace("resource is multihash", "key", manifestAddr)

				// get the manifest the multihash digest points to
				trie, err := loadManifest(self.fileStore, manifestAddr, nil)
				if err != nil {
					apiGetNotFound.Inc(1)
					status = http.StatusNotFound
					log.Warn(fmt.Sprintf("loadManifestTrie (resource multihash) error: %v", err))
					return reader, mimeType, status, nil, err
				}

				// finally, get the manifest entry
				// it will always be the entry on path ""
				entry, _ = trie.getEntry(path)
				if entry == nil {
					status = http.StatusNotFound
					apiGetNotFound.Inc(1)
					err = fmt.Errorf("manifest (resource multihash) entry for '%s' not found", path)
					log.Trace("manifest (resource multihash) entry not found", "key", manifestAddr, "path", path)
					return reader, mimeType, status, nil, err
				}

			} else {
				// data is returned verbatim since it's not a multihash
				return rsrc, "application/octet-stream", http.StatusOK, nil, nil
			}
		}

		// regardless of resource update manifests or normal manifests we will converge at this point
		// get the key the manifest entry points to and serve it if it's unambiguous
		contentAddr = common.Hex2Bytes(entry.Hash)
		status = entry.Status
		if status == http.StatusMultipleChoices {
			apiGetHttp300.Inc(1)
			return nil, entry.ContentType, status, contentAddr, err
		} else {
			mimeType = entry.ContentType
			log.Debug("content lookup key", "key", contentAddr, "mimetype", mimeType)
			reader, _ = self.fileStore.Retrieve(contentAddr)
		}
	} else {
		// no entry found
		status = http.StatusNotFound
		apiGetNotFound.Inc(1)
		err = fmt.Errorf("manifest entry for '%s' not found", path)
		log.Trace("manifest entry not found", "key", contentAddr, "path", path)
	}
	return
}

func (self *Api) Modify(addr storage.Address, path, contentHash, contentType string) (storage.Address, error) {
	apiModifyCount.Inc(1)
	quitC := make(chan bool)
	trie, err := loadManifest(self.fileStore, addr, quitC)
	if err != nil {
		apiModifyFail.Inc(1)
		return nil, err
	}
	if contentHash != "" {
		entry := newManifestTrieEntry(&ManifestEntry{
			Path:        path,
			ContentType: contentType,
		}, nil)
		entry.Hash = contentHash
		trie.addEntry(entry, quitC)
	} else {
		trie.deleteEntry(path, quitC)
	}

	if err := trie.recalcAndStore(); err != nil {
		apiModifyFail.Inc(1)
		return nil, err
	}
	return trie.ref, nil
}

func (self *Api) AddFile(mhash, path, fname string, content []byte, nameresolver bool) (storage.Address, string, error) {
	apiAddFileCount.Inc(1)

	uri, err := Parse("bzz:/" + mhash)
	if err != nil {
		apiAddFileFail.Inc(1)
		return nil, "", err
	}
	mkey, err := self.Resolve(uri)
	if err != nil {
		apiAddFileFail.Inc(1)
		return nil, "", err
	}

	// trim the root dir we added
	if path[:1] == "/" {
		path = path[1:]
	}

	entry := &ManifestEntry{
		Path:        filepath.Join(path, fname),
		ContentType: mime.TypeByExtension(filepath.Ext(fname)),
		Mode:        0700,
		Size:        int64(len(content)),
		ModTime:     time.Now(),
	}

	mw, err := self.NewManifestWriter(mkey, nil)
	if err != nil {
		apiAddFileFail.Inc(1)
		return nil, "", err
	}

	fkey, err := mw.AddEntry(bytes.NewReader(content), entry)
	if err != nil {
		apiAddFileFail.Inc(1)
		return nil, "", err
	}

	newMkey, err := mw.Store()
	if err != nil {
		apiAddFileFail.Inc(1)
		return nil, "", err

	}

	return fkey, newMkey.String(), nil

}

func (self *Api) RemoveFile(mhash, path, fname string, nameresolver bool) (string, error) {
	apiRmFileCount.Inc(1)

	uri, err := Parse("bzz:/" + mhash)
	if err != nil {
		apiRmFileFail.Inc(1)
		return "", err
	}
	mkey, err := self.Resolve(uri)
	if err != nil {
		apiRmFileFail.Inc(1)
		return "", err
	}

	// trim the root dir we added
	if path[:1] == "/" {
		path = path[1:]
	}

	mw, err := self.NewManifestWriter(mkey, nil)
	if err != nil {
		apiRmFileFail.Inc(1)
		return "", err
	}

	err = mw.RemoveEntry(filepath.Join(path, fname))
	if err != nil {
		apiRmFileFail.Inc(1)
		return "", err
	}

	newMkey, err := mw.Store()
	if err != nil {
		apiRmFileFail.Inc(1)
		return "", err

	}

	return newMkey.String(), nil
}

func (self *Api) AppendFile(mhash, path, fname string, existingSize int64, content []byte, oldAddr storage.Address, offset int64, addSize int64, nameresolver bool) (storage.Address, string, error) {
	apiAppendFileCount.Inc(1)

	buffSize := offset + addSize
	if buffSize < existingSize {
		buffSize = existingSize
	}

	buf := make([]byte, buffSize)

	oldReader, _ := self.Retrieve(oldAddr)
	io.ReadAtLeast(oldReader, buf, int(offset))

	newReader := bytes.NewReader(content)
	io.ReadAtLeast(newReader, buf[offset:], int(addSize))

	if buffSize < existingSize {
		io.ReadAtLeast(oldReader, buf[addSize:], int(buffSize))
	}

	combinedReader := bytes.NewReader(buf)
	totalSize := int64(len(buf))

	// TODO(jmozah): to append using pyramid chunker when it is ready
	//oldReader := self.Retrieve(oldKey)
	//newReader := bytes.NewReader(content)
	//combinedReader := io.MultiReader(oldReader, newReader)

	uri, err := Parse("bzz:/" + mhash)
	if err != nil {
		apiAppendFileFail.Inc(1)
		return nil, "", err
	}
	mkey, err := self.Resolve(uri)
	if err != nil {
		apiAppendFileFail.Inc(1)
		return nil, "", err
	}

	// trim the root dir we added
	if path[:1] == "/" {
		path = path[1:]
	}

	mw, err := self.NewManifestWriter(mkey, nil)
	if err != nil {
		apiAppendFileFail.Inc(1)
		return nil, "", err
	}

	err = mw.RemoveEntry(filepath.Join(path, fname))
	if err != nil {
		apiAppendFileFail.Inc(1)
		return nil, "", err
	}

	entry := &ManifestEntry{
		Path:        filepath.Join(path, fname),
		ContentType: mime.TypeByExtension(filepath.Ext(fname)),
		Mode:        0700,
		Size:        totalSize,
		ModTime:     time.Now(),
	}

	fkey, err := mw.AddEntry(io.Reader(combinedReader), entry)
	if err != nil {
		apiAppendFileFail.Inc(1)
		return nil, "", err
	}

	newMkey, err := mw.Store()
	if err != nil {
		apiAppendFileFail.Inc(1)
		return nil, "", err

	}

	return fkey, newMkey.String(), nil

}

func (self *Api) BuildDirectoryTree(mhash string, nameresolver bool) (addr storage.Address, manifestEntryMap map[string]*manifestTrieEntry, err error) {

	uri, err := Parse("bzz:/" + mhash)
	if err != nil {
		return nil, nil, err
	}
	addr, err = self.Resolve(uri)
	if err != nil {
		return nil, nil, err
	}

	quitC := make(chan bool)
	rootTrie, err := loadManifest(self.fileStore, addr, quitC)
	if err != nil {
		return nil, nil, fmt.Errorf("can't load manifest %v: %v", addr.String(), err)
	}

	manifestEntryMap = map[string]*manifestTrieEntry{}
	err = rootTrie.listWithPrefix(uri.Path, quitC, func(entry *manifestTrieEntry, suffix string) {
		manifestEntryMap[suffix] = entry
	})

	if err != nil {
		return nil, nil, fmt.Errorf("list with prefix failed %v: %v", addr.String(), err)
	}
	return addr, manifestEntryMap, nil
}

// Look up mutable resource updates at specific periods and versions
func (self *Api) ResourceLookup(ctx context.Context, addr storage.Address, period uint32, version uint32, maxLookup *mru.LookupParams) (string, []byte, error) {
	var err error
	rsrc, err := self.resource.Load(addr)
	if err != nil {
		return "", nil, err
	}
	if version != 0 {
		if period == 0 {
			return "", nil, mru.NewError(mru.ErrInvalidValue, "Period can't be 0")
		}
		_, err = self.resource.LookupVersion(ctx, rsrc.NameHash(), period, version, true, maxLookup)
	} else if period != 0 {
		_, err = self.resource.LookupHistorical(ctx, rsrc.NameHash(), period, true, maxLookup)
	} else {
		_, err = self.resource.LookupLatest(ctx, rsrc.NameHash(), true, maxLookup)
	}
	if err != nil {
		return "", nil, err
	}
	var data []byte
	_, data, err = self.resource.GetContent(rsrc.NameHash().Hex())
	if err != nil {
		return "", nil, err
	}
	return rsrc.Name(), data, nil
}

func (self *Api) ResourceCreate(ctx context.Context, name string, frequency uint64) (storage.Address, error) {
	key, _, err := self.resource.New(ctx, name, frequency)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (self *Api) ResourceUpdateMultihash(ctx context.Context, name string, data []byte) (storage.Address, uint32, uint32, error) {
	return self.resourceUpdate(ctx, name, data, true)
}
func (self *Api) ResourceUpdate(ctx context.Context, name string, data []byte) (storage.Address, uint32, uint32, error) {
	return self.resourceUpdate(ctx, name, data, false)
}

func (self *Api) resourceUpdate(ctx context.Context, name string, data []byte, multihash bool) (storage.Address, uint32, uint32, error) {
	var addr storage.Address
	var err error
	if multihash {
		addr, err = self.resource.UpdateMultihash(ctx, name, data)
	} else {
		addr, err = self.resource.Update(ctx, name, data)
	}
	period, _ := self.resource.GetLastPeriod(name)
	version, _ := self.resource.GetVersion(name)
	return addr, period, version, err
}

func (self *Api) ResourceHashSize() int {
	return self.resource.HashSize
}

func (self *Api) ResourceIsValidated() bool {
	return self.resource.IsValidated()
}

func (self *Api) ResolveResourceManifest(addr storage.Address) (storage.Address, error) {
	trie, err := loadManifest(self.fileStore, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot load resource manifest: %v", err)
	}

	entry, _ := trie.getEntry("")
	if entry.ContentType != ResourceContentType {
		return nil, fmt.Errorf("not a resource manifest: %s", addr)
	}

	return storage.Address(common.FromHex(entry.Hash)), nil
}

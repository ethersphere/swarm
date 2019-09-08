// Copyright 2017 The go-ethereum Authors
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

package http

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethersphere/swarm/api"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/feed"
	"github.com/ethersphere/swarm/storage/localstore"
	"github.com/ethersphere/swarm/storage/pin"
	lru "github.com/hashicorp/golang-lru"
)

const (
	cacheCapacity = 100000
)

type TestServer interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func NewTestSwarmServer(t *testing.T, serverFunc func(*api.API, *pin.API) TestServer, resolver api.Resolver, o *localstore.Options) *TestSwarmServer {
	t.Helper()
	ts, err := newTestSwarmServer(serverFunc, resolver, o)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

func newTestSwarmServer(serverFunc func(*api.API, *pin.API) TestServer, resolver api.Resolver, o *localstore.Options) (*TestSwarmServer, error) {

	swarmDir, err := ioutil.TempDir("", "swarm-storage-test")
	if err != nil {
		return nil, err
	}

	stateStore, err := state.NewDBStore(filepath.Join(swarmDir, "state-store.db"))
	if err != nil {
		return nil, err
	}

	localStore, err := localstore.New(swarmDir, make([]byte, 32), o)
	if err != nil {
		os.RemoveAll(swarmDir)
		return nil, err
	}

	tags := chunk.NewTags()
	cache, _ := lru.New(cacheCapacity)
	store := &testStore{cache, localStore}
	fileStore := storage.NewFileStore(store, store, storage.NewFileStoreParams(), tags)

	// Swarm feeds test setup
	feedsDir, err := ioutil.TempDir("", "swarm-feeds-test")
	if err != nil {
		return nil, err
	}

	feeds, err := feed.NewTestHandler(feedsDir, &feed.HandlerParams{})
	if err != nil {
		return nil, err
	}

	swarmApi := api.NewAPI(fileStore, resolver, feeds.Handler, nil, tags)
	pinAPI := pin.NewAPI(localStore, stateStore, nil, tags, swarmApi)
	apiServer := httptest.NewServer(serverFunc(swarmApi, pinAPI))

	tss := &TestSwarmServer{
		Server:    apiServer,
		FileStore: fileStore,
		Tags:      tags,
		dir:       swarmDir,
		Hasher:    storage.MakeHashFunc(storage.DefaultHash)(),
		cleanup: func() {
			apiServer.Close()
			fileStore.Close()
			feeds.Close()
			os.RemoveAll(swarmDir)
			os.RemoveAll(feedsDir)
		},
		CurrentTime: 42,
	}
	feed.TimestampProvider = tss
	return tss, nil
}

type TestSwarmServer struct {
	*httptest.Server
	Hasher      storage.SwarmHash
	FileStore   *storage.FileStore
	Tags        *chunk.Tags
	dir         string
	cleanup     func()
	CurrentTime uint64
}

func (t *TestSwarmServer) Close() {
	t.cleanup()
}

func (t *TestSwarmServer) Now() feed.Timestamp {
	return feed.Timestamp{Time: t.CurrentTime}
}

var (
	Delay = 1
)

type testStore struct {
	cache *lru.Cache
	*localstore.DB
}

func (ts *testStore) Get(ctx context.Context, mode chunk.ModeGet, ref storage.Address) (chunk.Chunk, error) {
	val, ok := ts.cache.Get(ref.Hex())
	if ok {
		log.Warn("from cache", "ref", ref.Hex()[:8])
		return val.(chunk.Chunk), nil
	}
	ch, err := ts.DB.Get(ctx, mode, ref)
	// log.Warn("db", "delay", Delay, "ref", ref.Hex()[:8])
	time.Sleep(time.Duration(Delay) * time.Microsecond)
	ts.cache.Add(ref.Hex(), ch)
	// log.Warn("from db", "ref", ref.Hex()[:8])
	return ch, err
}

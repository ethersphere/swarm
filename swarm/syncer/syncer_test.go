package syncer

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/network/simulation"
	"github.com/ethereum/go-ethereum/swarm/network/stream"
	"github.com/ethereum/go-ethereum/swarm/pot"
	"github.com/ethereum/go-ethereum/swarm/pss"
	"github.com/ethereum/go-ethereum/swarm/state"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

const neighbourhoodSize = 2

// tests the syncer
// tests how dispatcher of a push-syncing node interfaces with the sync db
// communicate with storers via mock PubSub
func TestSyncerWithLoopbackPubSub(t *testing.T) {
	// mock pubsub messenger
	lb := &loopback{make(map[string][]func(msg []byte, p *p2p.Peer) error)}

	// initialise syncer
	baseAddr := network.RandomAddr().OAddr
	chunkStore := storage.NewMapChunkStore()
	dbpath, err := ioutil.TempDir(os.TempDir(), "syncertest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dbpath)
	s, err := New(dbpath, baseAddr, chunkStore, lb)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// add a client storer and hook it to loopback pubsub
	storerChunkStore := storage.NewMapChunkStore()
	newStorer(storerChunkStore).withPubSub(lb)

	// fill up with chunks
	chunkCnt := 100
	for i := 0; i < chunkCnt; i++ {
		ch := storage.GenerateRandomChunk(int64(rand.Intn(chunk.DefaultSize)))
		s.Put("test", ch)
	}

	err = waitTillEmpty(s.db)
	if err != nil {
		t.Fatal(err)
	}

}

// pubSubOracle implements a fake comms protocol to serve as pubsub, that dispatches chunks to nodes
// nearest neighbours
type pubSubOracle struct {
	pot           *pot.Pot
	pof           pot.Pof
	chunkHandlers map[string]func(msg []byte, p *p2p.Peer) error
	proofHandlers map[string]func(msg []byte, p *p2p.Peer) error
}

func newPubSubOracle() *pubSubOracle {
	return &pubSubOracle{
		pof:           pot.DefaultPof(256),
		chunkHandlers: make(map[string]func(msg []byte, p *p2p.Peer) error),
		proofHandlers: make(map[string]func(msg []byte, p *p2p.Peer) error),
	}
}

func (p *pubSubOracle) Send(to []byte, topic string, msg []byte) error {
	addr := storage.Address(to)
	peer := p2p.NewPeer(enode.ID{}, "", nil)
	if topic == pssReceiptTopic {
		f, ok := p.proofHandlers[addr.Hex()]
		if !ok {
			return fmt.Errorf("cannot send to %v", addr.Hex())
		}
		f(msg, peer)
		return nil
	}
	nns := p.getNNS(addr)
	for _, nn := range nns {
		f := p.chunkHandlers[storage.Address(nn).Hex()]
		f(msg, peer)
	}
	return nil

}

func (p *pubSubOracle) getNNS(addr []byte) (peers [][]byte) {
	if p.pot == nil {
		p.initPOT()
	}
	n := 0
	p.pot.EachNeighbour(addr, p.pof, func(v pot.Val, i int) bool {
		peers = append(peers, v.([]byte))
		n++
		return n < neighbourhoodSize
	})
	return peers
}

func (p *pubSubOracle) initPOT() {
	p.pot = pot.NewPot(nil, 0)
	for k := range p.chunkHandlers {
		addr := common.Hex2Bytes(k)
		p.pot, _, _ = pot.Add(p.pot, addr, p.pof)
	}
}

type pubsub struct {
	*pubSubOracle
	addr storage.Address
}

func (ps *pubsub) Register(topic string, handler func(msg []byte, p *p2p.Peer) error) {
	if topic == pssReceiptTopic {
		ps.pubSubOracle.proofHandlers[ps.addr.Hex()] = handler
		return
	}
	ps.pubSubOracle.chunkHandlers[ps.addr.Hex()] = handler
}

func (p *pubSubOracle) new(b []byte) PubSub {
	return &pubsub{pubSubOracle: p, addr: storage.Address(b)}
}

var (
	bucketKeySyncer = simulation.BucketKey("syncer")
)

func upload(ctx context.Context, s *Syncer, tagname string, n int) (addrs []storage.Address, err error) {
	tg, err := s.NewTag(tagname, n)
	if err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		ch := storage.GenerateRandomChunk(int64(chunk.DefaultSize))
		addrs = append(addrs, ch.Address())
		s.Put(tagname, ch)
	}
	err = tg.WaitTill(ctx, SYNCED)
	if err != nil {
		return nil, err
	}
	return addrs, nil
}

func download(ctx context.Context, s *Syncer, addrs []storage.Address) error {
	errc := make(chan error)
	for _, addr := range addrs {
		go func(addr storage.Address) {
			_, err := s.db.chunkStore.Get(ctx, addr)
			select {
			case errc <- err:
			case <-ctx.Done():
			}
		}(addr)
	}
	i := 0
	for err := range errc {
		if err != nil {
			return err
		}
		i++
		if i == len(addrs) {
			break
		}
	}
	return nil
}

// tests syncing using simulation framework
// - snapshot sets up a healthy kademlia
// - pubSubOracle (offband delivery) is mocking transport layer for syncing protocol to the neighbourhood
// - uses real streamer retrieval requests to download
// - it performs several trials concurrently
// - each trial an uploader and a downloader node is selected and uploader uploads a number of chunks
//   and after it is synced the downloader node is retrieving the content
//   there is no retries, if any of the downloads times out or the
func TestSyncerWithPubSubOracle(t *testing.T) {
	nodeCnt := 128
	chunkCnt := 100
	trials := 100

	// offband syncing to nearest neighbourhood
	psg := newPubSubOracle()
	psSyncerF := func(addr []byte, _ *pss.Pss) PubSub {
		return psg.new(addr)
	}
	err := testSyncerWithPubSub(nodeCnt, chunkCnt, trials, newServiceFunc(psSyncerF, nil))
	if err != nil {
		t.Fatal(err)
	}
}

// test syncer using pss
func TestSyncerWithPss(t *testing.T) {
	nodeCnt := 32
	chunkCnt := 1
	trials := 1
	psSyncerF := func(_ []byte, p *pss.Pss) PubSub {
		return NewPss(p, false)
	}
	psStorerF := func(_ []byte, p *pss.Pss) PubSub {
		return NewPss(p, true)
	}
	err := testSyncerWithPubSub(nodeCnt, chunkCnt, trials, newServiceFunc(psSyncerF, psStorerF))
	if err != nil {
		t.Fatal(err)
	}
}

func testSyncerWithPubSub(nodeCnt, chunkCnt, trials int, sf simulation.ServiceFunc) error {
	sim := simulation.New(map[string]simulation.ServiceFunc{
		"streamer": sf,
	})
	defer sim.Close()

	err := sim.UploadSnapshot(fmt.Sprintf("../network/stream/testing/snapshot_%d.json", nodeCnt))
	if err != nil {
		return err
	}

	choose2 := func(n int) (i, j int) {
		i = rand.Intn(n)
		j = rand.Intn(n - 1)
		if j >= i {
			j++
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	errc := make(chan error)
	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		for i := 0; i < trials; i++ {
			if i%10 == 0 && i > 0 {
				time.Sleep(1000 * time.Millisecond)
			}
			go func(i int) {
				u, d := choose2(nodeCnt)
				uid := sim.UpNodeIDs()[u]
				syncer, _ := sim.NodeItem(uid, bucketKeySyncer)
				did := sim.UpNodeIDs()[d]
				tagname := fmt.Sprintf("tag-%v-%v-%d", label(uid[:]), label(did[:]), i)
				log.Error("uploading", "peer", uid, "chunks", chunkCnt, "tagname", tagname)
				what, err := upload(ctx, syncer.(*Syncer), tagname, chunkCnt)
				if err != nil {
					select {
					case errc <- err:
					case <-ctx.Done():
					}
					return
				}
				log.Error("synced", "peer", did, "chunks", chunkCnt, "tagname", tagname)
				log.Error("downloading", "peer", did, "chunks", chunkCnt, "tagname", tagname)

				syncer, _ = sim.NodeItem(did, bucketKeySyncer)
				err = download(ctx, syncer.(*Syncer), what)
				select {
				case errc <- err:
				case <-ctx.Done():
				}
				log.Error("downloaded", "peer", did, "chunks", chunkCnt, "tagname", tagname)

			}(i)
		}
		i := 0
		for err := range errc {
			if err != nil {
				return err
			}
			i++
			if i == trials {
				break
			}
		}
		return nil
	})

	if result.Error != nil {
		return fmt.Errorf("simulation error: %v", result.Error)
	}
	log.Error("PASS")
	return nil
}

func newServiceFunc(psSyncer, psStorer func([]byte, *pss.Pss) PubSub) func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
	return func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
		n := ctx.Config.Node()
		addr := network.NewAddr(n)
		datadir, err := ioutil.TempDir(os.TempDir(), fmt.Sprintf("chunkstore-%s", n.ID().TerminalString()))
		if err != nil {
			return nil, nil, err
		}
		params := storage.NewDefaultLocalStoreParams()
		params.ChunkDbPath = datadir
		params.BaseKey = addr.Over()
		localStore, err := storage.NewTestLocalStoreForAddr(params)
		if err != nil {
			os.RemoveAll(datadir)
			return nil, nil, err
		}
		netStore, err := storage.NewNetStore(localStore, nil)
		if err != nil {
			return nil, nil, err
		}

		// pss
		kadParams := network.NewKadParams()
		kadParams.MinProxBinSize = 2
		kad := network.NewKademlia(addr.Over(), kadParams)
		privKey, err := crypto.GenerateKey()
		pssp := pss.NewPssParams().WithPrivateKey(privKey)
		ps, err := pss.NewPss(kad, pssp)
		if err != nil {
			return nil, nil, err
		}

		// streamer
		delivery := stream.NewDelivery(kad, netStore)
		netStore.NewNetFetcherFunc = network.NewFetcherFactory(delivery.RequestFromPeers, true).New

		r := stream.NewRegistry(addr.ID(), delivery, netStore, state.NewInmemoryStore(), &stream.RegistryOptions{
			Syncing:   stream.SyncingDisabled,
			Retrieval: stream.RetrievalEnabled,
		}, nil)

		// set up syncer
		dbpath, err := ioutil.TempDir(os.TempDir(), fmt.Sprintf("syncdb-%s", n.ID().TerminalString()))
		if err != nil {
			os.RemoveAll(datadir)
			return nil, nil, err
		}
		defer os.RemoveAll(dbpath)
		p := psSyncer(addr.OAddr, ps)
		syn, err := New(dbpath, addr.OAddr, netStore, p)
		if err != nil {
			os.RemoveAll(datadir)
			os.RemoveAll(dbpath)
			return nil, nil, err
		}
		bucket.Store(bucketKeySyncer, syn)

		// also work as a syncer storer client
		if psStorer != nil {
			p = psStorer(addr.OAddr, ps)
		}
		st := newStorer(netStore).withPubSub(p)
		_ = st

		cleanup = func() {
			syn.Close()
			netStore.Close()
			os.RemoveAll(datadir)
			os.RemoveAll(dbpath)
			r.Close()
		}

		return &StreamerAndPss{r, ps}, cleanup, nil
	}
}

// implements the node.Service interface
type StreamerAndPss struct {
	*stream.Registry
	pss *pss.Pss
}

func (s *StreamerAndPss) Protocols() []p2p.Protocol {
	return append(s.Registry.Protocols(), s.pss.Protocols()...)
}

func (s *StreamerAndPss) Start(srv *p2p.Server) error {
	err := s.Registry.Start(srv)
	if err != nil {
		return err
	}
	return s.pss.Start(srv)
}

func (s *StreamerAndPss) Stop() error {
	err := s.Registry.Stop()
	if err != nil {
		return err
	}
	return s.pss.Stop()
}

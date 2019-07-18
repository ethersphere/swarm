package client

import (
	"strings"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/storage"
)

type Bzz struct {
	client *rpc.Client
}

// NewBzz is a constructor for a Bzz API
func NewBzz(client *rpc.Client) *Bzz {
	return &Bzz{
		client: client,
	}
}

// GetChunksBitVector returns a bit vector of presence for a given slice of chunks
func (b *Bzz) GetChunksBitVector(addrs []storage.Address) (string, error) {
	var hostChunks string
	const trackChunksPageSize = 7500

	for len(addrs) > 0 {
		var pageChunks string
		// get current page size, so that we avoid a slice out of bounds on the last page
		pagesize := trackChunksPageSize
		if len(addrs) < trackChunksPageSize {
			pagesize = len(addrs)
		}

		err := b.client.Call(&pageChunks, "bzz_has", addrs[:pagesize])
		if err != nil {
			return "", err
		}
		hostChunks += pageChunks
		addrs = addrs[pagesize:]
	}

	return hostChunks, nil
}

// GetBzzAddr returns the bzzAddr of the node
func (b *Bzz) GetBzzAddr() (string, error) {
	var hive string

	err := b.client.Call(&hive, "bzz_hive")
	if err != nil {
		return "", err
	}

	// we make an ugly assumption about the output format of the hive.String() method
	// ideally we should replace this with an API call that returns the bzz addr for a given host,
	// but this also works for now (provided we don't change the hive.String() method, which we haven't in some time
	ss := strings.Split(strings.Split(hive, "\n")[3], " ")
	return ss[len(ss)-1], nil
}

// IsPullSyncing is checking if the node is still receiving chunk deliveries due to pull syncing
func (b *Bzz) IsPullSyncing() (bool, error) {
	var isSyncing bool

	err := b.client.Call(&isSyncing, "bzz_isPullSyncing")
	if err != nil {
		log.Error("error calling host for isPullSyncing", "err", err)
		return false, err
	}

	return isSyncing, nil
}

// IsPushSynced checks if the given `tag` is done syncing, i.e. we've received receipts for all chunks
func (b *Bzz) IsPushSynced(tagname string) (bool, error) {
	var isSynced bool

	err := b.client.Call(&isSynced, "bzz_isPushSynced", tagname)
	if err != nil {
		log.Error("error calling host for isPushSynced", "err", err)
		return false, err
	}

	return isSynced, nil
}

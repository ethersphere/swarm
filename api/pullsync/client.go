package pullsync

import stream "github.com/ethersphere/swarm/network/stream/v2"



// the node-wide pullsync.Client
type Client struct {
	stream.Syncer // embed stream.Syncer
	// when pullsync
	// here  you simply put the update sync logic listening to kademlia depth changes
	// and call `Request`
	// remember the request, when no longer relevant just call request.Cancel()
}

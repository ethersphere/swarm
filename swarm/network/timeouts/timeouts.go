package timeouts

import "time"

// FailedPeerSkipDelay is the time we consider a peer to be skipped for a particular request/chunk,
// because this peer failed to deliver it during the SearchTimeout interval
var FailedPeerSkipDelay = 10 * time.Second

// FetcherGlobalTimeout is the max time a node tries to find a chunk for a client, after which it returns a 404
// Basically this is the amount of time a singleflight request for a given chunk lives
var FetcherGlobalTimeout = 10 * time.Second

// SearchTimeout is the max time we wait for a peer to deliver a chunk we requests, after which we try another peer
var SearchTimeout = 1 * time.Second

package ens

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/swarm/services/ens/contract"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

var domainAndVersion = regexp.MustCompile("[@:;,]+")
var qtypeChash = [32]byte{ 0x43, 0x48, 0x41, 0x53, 0x48}

// swarm domain name registry and resolver
// the ENS instance can be directly wrapped in rpc.Api
type ENS struct {
	session *contract.ResolverSession
	contractBackend bind.ContractBackend
}

// NewENS creates a proxy instance wrapping the abigen interface to the ENS contract
// using the transaction options passed as first argument, it sets up a session
func NewENS(transactOpts *bind.TransactOpts, contractAddr common.Address, contractBackend bind.ContractBackend) *ENS {
	ens, err := contract.NewResolver(contractAddr, contractBackend)
	if err != nil {
		glog.V(logger.Debug).Infof("error setting up name server on %v, skipping: %v", contractAddr.Hex(), err)
	}
	return &ENS{
		&contract.ResolverSession{
			Contract:     ens,
			TransactOpts: *transactOpts,
		},
		contractBackend,
	}
}

// resolve is a non-tranasctional call, returns hash as storage.Key
func (self *ENS) Resolve(hostPort string) (storage.Key, error) {
	host := hostPort
	parts := domainAndVersion.Split(host, 3)
	if len(parts) > 1 && parts[1] != "" {
		host = parts[0]
	}
	return self.resolveName(host)
}

func (self *ENS) resolveName(host string) (storage.Key, error) {
	labels := strings.Split(host, ".")
	resolver := self
	var nodeId [12]byte;

	for i := len(labels) - 1; i >= 0; i-- {
		hash := crypto.Sha3Hash([]byte(labels[i]))
		ret, err := resolver.session.FindResolver(nodeId, hash)
		if err != nil {
			return nil, fmt.Errorf("error resolving label '%v' of '%v': %v", labels[i], host, err)
		}
		if ret.Rcode != 0 {
			return nil, fmt.Errorf("error resolving label '%v' of '%v': got response code %v", labels[i], host, ret.Rcode)
		}
		nodeId = ret.Rnode;
		resolver = NewENS(&resolver.session.TransactOpts, ret.Raddress, resolver.contractBackend)
	}

	ret, err := resolver.session.Resolve(nodeId, qtypeChash, 0)
	if err != nil {
		return nil, fmt.Errorf("error looking up RR on '%v': %v", host, err)
	}
	if ret.Rcode != 0 {
		return nil, fmt.Errorf("error looking up RR on '%v': got response code %v", host, ret.Rcode)
	}
	return storage.Key(ret.Data[:]), nil
}

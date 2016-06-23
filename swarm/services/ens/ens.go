package ens

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/services/ens/contract"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

var domainAndVersion = regexp.MustCompile("[@:;,]+")
var qtypeChash = [32]byte{ 0x43, 0x48, 0x41, 0x53, 0x48}

// swarm domain name registry and resolver
// the ENS instance can be directly wrapped in rpc.Api
type ENS struct {
	transactOpts *bind.TransactOpts;
	contractBackend bind.ContractBackend;
	rootAddress common.Address;
}

func NewENS(transactOpts *bind.TransactOpts, contractAddr common.Address, contractBackend bind.ContractBackend) *ENS {
	return &ENS{
		transactOpts: transactOpts,
		contractBackend: contractBackend,
		rootAddress: contractAddr,
	}
}

func (self *ENS) newResolver(contractAddr common.Address) (*contract.ResolverSession, error) {
	resolver, err := contract.NewResolver(contractAddr, self.contractBackend)
	if err != nil {
		return nil, err
	}
	return &contract.ResolverSession{
		Contract: resolver,
		TransactOpts: *self.transactOpts,
	}, nil
}

// resolve is a non-tranasctional call, returns hash as storage.Key
func (self *ENS) Resolve(hostPort string) (storage.Key, error) {
	host := hostPort
	parts := domainAndVersion.Split(host, 3)
	if len(parts) > 1 && parts[1] != "" {
		host = parts[0]
	}
	return self.resolveName(self.rootAddress, host)
}

func (self *ENS) findResolver(rootAddress common.Address, host string) (*contract.ResolverSession, [12]byte, error) {
	var nodeId [12]byte

	resolver, err := self.newResolver(rootAddress)
	if err != nil {
		return nil, [12]byte{}, err
	}

	labels := strings.Split(host, ".")

	for i := len(labels) - 1; i >= 0; i-- {
		hash := crypto.Sha3Hash([]byte(labels[i]))
		ret, err := resolver.FindResolver(nodeId, hash)
		if err != nil {
			err = fmt.Errorf("error resolving label '%v' of '%v': %v", labels[i], host, err)
			return nil, [12]byte{}, err
		}
		if ret.Rcode != 0 {
			err = fmt.Errorf("error resolving label '%v' of '%v': got response code %v", labels[i], host, ret.Rcode)
			return nil, [12]byte{}, err
		}
		nodeId = ret.Rnode;
		resolver, err = self.newResolver(ret.Raddress)
		if err != nil {
			return nil, [12]byte{}, err
		}
	}

	return resolver, nodeId, nil
}

func (self *ENS) resolveName(rootAddress common.Address, host string) (storage.Key, error) {
	resolver, nodeId, err := self.findResolver(rootAddress, host)
	if err != nil {
		return nil, err
	}

	ret, err := resolver.Resolve(nodeId, qtypeChash, 0)
	if err != nil {
		return nil, fmt.Errorf("error looking up RR on '%v': %v", host, err)
	}
	if ret.Rcode != 0 {
		return nil, fmt.Errorf("error looking up RR on '%v': got response code %v", host, ret.Rcode)
	}
	return storage.Key(ret.Data[:]), nil
}



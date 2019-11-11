package resolver

import (
	"errors"

	config "github.com/rsksmart/rds-swarm/config"
	multichainresolver "github.com/rsksmart/rds-swarm/resolver/multi_chain_resolver"
	rskresolver "github.com/rsksmart/rds-swarm/resolver/rsk_resolver"
	"github.com/rsksmart/rds-swarm/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ErrNoAddress is returned when there is no registered address through RNS
var ErrNoAddress = errors.New("domain without registered address in RNS")

// ErrNoContent is returned when there is no registered content through RNS
var ErrNoContent = errors.New("domain without registered content in RNS")

// Resolver interface is implemented by all types which can resolve both the address of a domain as well as its content.
type Resolver interface {
	Addr(opts *bind.CallOpts, node [32]byte) (common.Address, error)
	Content(opts *bind.CallOpts, node [32]byte) ([32]byte, error)
}

func getPublicResolver(client *ethclient.Client, configuration config.Configuration) (Resolver, error) {
	resolverAddress := common.HexToAddress(configuration.ResolverAddresses.RSK)
	resolver, resolverError := rskresolver.NewRskresolver(resolverAddress, client)
	if resolverError != nil {
		return nil, resolverError
	}

	return resolver, nil
}

func getMultiChainResolver(client *ethclient.Client, configuration config.Configuration) (Resolver, error) {
	resolverAddress := common.HexToAddress(configuration.ResolverAddresses.MultiChain)
	resolver, resolverError := multichainresolver.NewMultichainresolver(resolverAddress, client)
	if resolverError != nil {
		return nil, resolverError
	}

	return resolver, nil
}

func setUpResolver(resolverConstructor func(client *ethclient.Client, configuration config.Configuration) (Resolver, error)) (Resolver, error) {
	configuration := config.GetConfiguration()

	client, clientError := ethclient.Dial(configuration.NetworkNodeAddress)
	if clientError != nil {
		return nil, clientError
	}
	defer client.Close()

	resolver, resolverError := resolverConstructor(client, configuration)
	if resolverError != nil {
		return nil, resolverError
	}

	return resolver, nil
}

func resolveAddressFromResolver(domainAddress [32]byte, getResolverFunction func(client *ethclient.Client, configuration config.Configuration) (Resolver, error)) (common.Address, error) {
	var emptyAddress common.Address

	resolver, resolverError := setUpResolver(getResolverFunction)
	if resolverError != nil {
		return emptyAddress, resolverError
	}

	resolvedAddress, resolutionError := resolveAddress(domainAddress, resolver)
	if resolutionError != nil {
		return emptyAddress, resolutionError
	}

	return resolvedAddress, nil
}

func resolveAddress(byteArrayAddress [32]byte, resolver Resolver) (common.Address, error) {
	return resolver.Addr(&bind.CallOpts{}, byteArrayAddress)
}

func resolveContentFromResolver(domainAddress [32]byte, getResolverFunction func(client *ethclient.Client, configuration config.Configuration) (Resolver, error)) ([32]byte, error) {
	var emptyContent [32]byte

	resolver, resolverError := setUpResolver(getResolverFunction)
	if resolverError != nil {
		return emptyContent, resolverError
	}

	resolvedContent, resolutionError := resolveContent(domainAddress, resolver)
	if resolutionError != nil {
		return emptyContent, resolutionError
	}

	return resolvedContent, nil
}

func resolveContent(byteArrayAddress [32]byte, resolver Resolver) ([32]byte, error) {
	return resolver.Content(&bind.CallOpts{}, byteArrayAddress)
}

// ResolveDomainAddress receives a domain string and returns its RNS-resolved hex address.
// It will attempt to solve the address through the Multi-Chain resolver first, and through the Public resolver later if the former results in an empty address.
func ResolveDomainAddress(domain string) (common.Address, error) {
	domainAddress := utils.DomainToHashedByteArray(domain)
	var emptyAddress, resolvedAddress common.Address
	var resolvedError error

	resolvedAddress, resolvedError = resolveAddressFromResolver(domainAddress, getMultiChainResolver)
	if resolvedError != nil {
		return emptyAddress, resolvedError
	}

	if resolvedAddress == emptyAddress {
		resolvedAddress, resolvedError = resolveAddressFromResolver(domainAddress, getPublicResolver)
		if resolvedError != nil {
			return emptyAddress, resolvedError
		}
	}

	if resolvedAddress == emptyAddress {
		resolvedError = ErrNoAddress
	}
	return resolvedAddress, resolvedError
}

// ResolveDomainContent receives a domain string and returns its RNS-resolved associated content hash.
// It will attempt to solve the content through the Multi-Chain resolver first, and through the Public resolver later if the former results in an empty content.
func ResolveDomainContent(domain string) (common.Hash, error) {
	domainAddress := utils.DomainToHashedByteArray(domain)
	var emptyContent, resolvedContent [32]byte
	var resolvedError error

	resolvedContent, resolvedError = resolveContentFromResolver(domainAddress, getMultiChainResolver)
	if resolvedError != nil {
		return emptyContent, resolvedError
	}

	if resolvedContent == emptyContent {
		resolvedContent, resolvedError = resolveContentFromResolver(domainAddress, getPublicResolver)
		if resolvedError != nil {
			return emptyContent, resolvedError
		}
	}

	if resolvedContent == emptyContent {
		resolvedError = ErrNoContent
	}
	return resolvedContent, resolvedError
}

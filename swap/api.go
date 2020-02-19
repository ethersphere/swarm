package swap

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	contract "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/uint256"
)

// APIs is a node.Service interface method
func (s *Swap) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "swap",
			Version:   "1.0",
			Service:   NewAPI(s),
			Public:    false,
		},
	}
}

type swapAPI interface {
	AvailableBalance() (*uint256.Uint256, error)
	PeerBalance(peer enode.ID) (int64, error)
	Balances() (map[enode.ID]int64, error)
	PeerCheques(peer enode.ID) (PeerCheques, error)
	Cheques() (map[enode.ID]*PeerCheques, error)
}

// API would be the API accessor for protocol methods
type API struct {
	swapAPI
	*contract.Params
}

// PeerCheques contains the last cheque known to have been sent to a peer, as well as the last one received from the peer
type PeerCheques struct {
	PendingCheque      *Cheque
	LastSentCheque     *Cheque
	LastReceivedCheque *Cheque
}

// NewAPI creates a new API instance
func NewAPI(s *Swap) *API {
	return &API{
		swapAPI: s,
		Params:  s.GetParams(),
	}
}

// AvailableBalance returns the total balance of the chequebook against which new cheques can be written
func (s *Swap) AvailableBalance() (*uint256.Uint256, error) {
	// get the LiquidBalance of the chequebook
	contractLiquidBalance, err := s.contract.LiquidBalance(nil)
	if err != nil {
		return nil, err
	}

	// get all cheques
	cheques, err := s.Cheques()
	if err != nil {
		return nil, err
	}

	// Compute the total worth of cheques sent and how much of of this is cashed
	sentChequesWorth := new(big.Int)
	cashedChequesWorth := new(big.Int)
	for _, peerCheques := range cheques {
		var sentCheque *Cheque
		if peerCheques.PendingCheque != nil {
			sentCheque = peerCheques.PendingCheque
		} else if peerCheques.LastSentCheque != nil {
			sentCheque = peerCheques.LastSentCheque
		} else {
			continue
		}
		cumulativePayout := sentCheque.ChequeParams.CumulativePayout.Value()
		sentChequesWorth.Add(sentChequesWorth, &cumulativePayout)
		paidOut, err := s.contract.PaidOut(nil, sentCheque.ChequeParams.Beneficiary)
		if err != nil {
			return nil, err
		}
		cashedChequesWorth.Add(cashedChequesWorth, paidOut)
	}

	totalChequesWorth := new(big.Int).Sub(cashedChequesWorth, sentChequesWorth)
	tentativeLiquidBalance := new(big.Int).Add(contractLiquidBalance, totalChequesWorth)

	return uint256.New().Set(*tentativeLiquidBalance)
}

// PeerBalance returns the balance for a given peer
func (s *Swap) PeerBalance(peer enode.ID) (balance int64, err error) {
	if swapPeer := s.getPeer(peer); swapPeer != nil {
		swapPeer.lock.Lock()
		defer swapPeer.lock.Unlock()
		return swapPeer.getBalance(), nil
	}
	err = s.store.Get(balanceKey(peer), &balance)
	return balance, err
}

// Balances returns the balances for all known SWAP peers
func (s *Swap) Balances() (map[enode.ID]int64, error) {
	balances := make(map[enode.ID]int64)

	s.peersLock.Lock()
	for peer, swapPeer := range s.peers {
		swapPeer.lock.Lock()
		balances[peer] = swapPeer.getBalance()
		swapPeer.lock.Unlock()
	}
	s.peersLock.Unlock()

	// add store balances, if peer was not already added
	balanceIterFunction := func(key []byte, value []byte) (stop bool, err error) {
		peer := keyToID(string(key), balancePrefix)
		if _, peerHasBalance := balances[peer]; !peerHasBalance {
			var peerBalance int64
			err = json.Unmarshal(value, &peerBalance)
			if err == nil {
				balances[peer] = peerBalance
			}
		}
		return stop, err
	}
	err := s.store.Iterate(balancePrefix, balanceIterFunction)
	if err != nil {
		return nil, err
	}

	return balances, nil
}

// PeerCheques returns the last sent and received cheques for a given peer
func (s *Swap) PeerCheques(peer enode.ID) (PeerCheques, error) {
	var pendingCheque, sentCheque, receivedCheque *Cheque

	swapPeer := s.getPeer(peer)
	if swapPeer != nil {
		swapPeer.lock.Lock()
		pendingCheque = swapPeer.getPendingCheque()
		sentCheque = swapPeer.getLastSentCheque()
		receivedCheque = swapPeer.getLastReceivedCheque()
		swapPeer.lock.Unlock()
	} else {
		errPendingCheque := s.store.Get(pendingChequeKey(peer), &pendingCheque)
		if errPendingCheque != nil && errPendingCheque != state.ErrNotFound {
			return PeerCheques{}, errPendingCheque
		}
		errSentCheque := s.store.Get(sentChequeKey(peer), &sentCheque)
		if errSentCheque != nil && errSentCheque != state.ErrNotFound {
			return PeerCheques{}, errSentCheque
		}
		errReceivedCheque := s.store.Get(receivedChequeKey(peer), &receivedCheque)
		if errReceivedCheque != nil && errReceivedCheque != state.ErrNotFound {
			return PeerCheques{}, errReceivedCheque
		}
	}
	return PeerCheques{pendingCheque, sentCheque, receivedCheque}, nil
}

// Cheques returns all known last sent and received cheques, grouped by peer
func (s *Swap) Cheques() (map[enode.ID]*PeerCheques, error) {
	cheques := make(map[enode.ID]*PeerCheques)

	// get peer cheques from memory
	s.peersLock.Lock()
	for peer, swapPeer := range s.peers {
		swapPeer.lock.Lock()
		pendingCheque := swapPeer.getPendingCheque()
		sentCheque := swapPeer.getLastSentCheque()
		receivedCheque := swapPeer.getLastReceivedCheque()
		// don't add peer to result if there are no cheques
		if sentCheque != nil || receivedCheque != nil || pendingCheque != nil {
			cheques[peer] = &PeerCheques{pendingCheque, sentCheque, receivedCheque}
		}
		swapPeer.lock.Unlock()
	}
	s.peersLock.Unlock()

	// get peer cheques from store
	err := s.addStoreCheques(pendingChequePrefix, cheques)
	if err != nil {
		return nil, err
	}
	err = s.addStoreCheques(sentChequePrefix, cheques)
	if err != nil {
		return nil, err
	}
	err = s.addStoreCheques(receivedChequePrefix, cheques)
	if err != nil {
		return nil, err
	}

	return cheques, nil
}

// add cheques from store for peers not already present in given cheques map
func (s *Swap) addStoreCheques(chequePrefix string, cheques map[enode.ID]*PeerCheques) error {
	chequesIterFunction := func(key []byte, value []byte) (stop bool, err error) {
		peer := keyToID(string(key), chequePrefix)
		// create struct if peer has no cheques entry yet
		if peerCheques := cheques[peer]; peerCheques == nil {
			cheques[peer] = &PeerCheques{}
		}

		// add cheque from store if not already in result
		var peerCheque Cheque
		err = json.Unmarshal(value, &peerCheque)
		if err == nil {
			switch chequePrefix {
			case pendingChequePrefix:
				cheques[peer].PendingCheque = &peerCheque
			case sentChequePrefix:
				cheques[peer].LastSentCheque = &peerCheque
			case receivedChequePrefix:
				cheques[peer].LastReceivedCheque = &peerCheque
			default:
				err = fmt.Errorf("unknown type of cheque requested through prefix %s", chequePrefix)
			}
		}
		return stop, err
	}
	return s.store.Iterate(chequePrefix, chequesIterFunction)
}

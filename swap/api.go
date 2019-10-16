package swap

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	contract "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/state"
)

const (
	sentChequeResponseKey     = "last_sent_cheque"
	receivedChequeResponseKey = "last_received_cheque"
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
	Balance(peer enode.ID) (int64, error)
	Balances() (map[enode.ID]int64, error)
	Cheques() (map[enode.ID]map[string]*Cheque, error)
	PeerCheques(peer enode.ID) (map[string]*Cheque, error)
}

// API would be the API accessor for protocol methods
type API struct {
	swapAPI
	*contract.Params
}

// NewAPI creates a new API instance
func NewAPI(s *Swap) *API {
	return &API{
		swapAPI: s,
		Params:  s.GetParams(),
	}
}

// Balance returns the balance for a given peer
func (s *Swap) Balance(peer enode.ID) (balance int64, err error) {
	if swapPeer := s.getPeer(peer); swapPeer != nil {
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

// Cheques returns all known last sent and received cheques, grouped by peer
func (s *Swap) Cheques() (map[enode.ID]map[string]*Cheque, error) {
	cheques := make(map[enode.ID]map[string]*Cheque)

	// get peer cheques from memory
	s.peersLock.Lock()
	for peer, swapPeer := range s.peers {
		swapPeer.lock.Lock()
		cheques[peer] = make(map[string]*Cheque)
		cheques[peer][sentChequeResponseKey] = swapPeer.getLastSentCheque()
		cheques[peer][receivedChequeResponseKey] = swapPeer.getLastReceivedCheque()
		swapPeer.lock.Unlock()
	}
	s.peersLock.Unlock()

	// get peer cheques from store
	err := s.addStoreCheques(sentChequePrefix, sentChequeResponseKey, cheques)
	if err != nil {
		return nil, err
	}
	err = s.addStoreCheques(receivedChequePrefix, receivedChequeResponseKey, cheques)
	if err != nil {
		return nil, err
	}

	// fill in result with missing cheques
	for _, peerCheques := range cheques {
		// add nil as type of cheque if not present
		if _, peerHasReceivedCheque := peerCheques[receivedChequeResponseKey]; !peerHasReceivedCheque {
			peerCheques[receivedChequeResponseKey] = nil
		}
		if _, peerHasSentCheque := peerCheques[sentChequeResponseKey]; !peerHasSentCheque {
			peerCheques[sentChequeResponseKey] = nil
		}
	}

	return cheques, nil
}

// add disk cheques for peers not already present in given cheques map
func (s *Swap) addStoreCheques(chequePrefix string, chequeKey string, cheques map[enode.ID]map[string]*Cheque) error {
	chequesIterFunction := func(key []byte, value []byte) (stop bool, err error) {
		peer := keyToID(string(key), chequePrefix)
		// make map if peer has no cheques entry yet
		if _, peerHasCheques := cheques[peer]; !peerHasCheques {
			cheques[peer] = make(map[string]*Cheque)
		}
		// add cheque from store if not already in result
		if peerCheque := cheques[peer][chequeKey]; peerCheque == nil {
			var peerCheque Cheque
			err = json.Unmarshal(value, &peerCheque)
			if err == nil {
				cheques[peer][chequeKey] = &peerCheque
			}
		}
		return stop, err
	}
	return s.store.Iterate(chequePrefix, chequesIterFunction)
}

// PeerCheques returns the last sent and received cheques for a given peer
func (s *Swap) PeerCheques(peer enode.ID) (map[string]*Cheque, error) {
	var sentCheque, receivedCheque *Cheque

	swapPeer := s.getPeer(peer)
	if swapPeer != nil {
		sentCheque = swapPeer.getLastSentCheque()
		receivedCheque = swapPeer.getLastReceivedCheque()
	} else {
		errSentCheque := s.store.Get(sentChequeKey(peer), &sentCheque)
		if errSentCheque != nil && errSentCheque != state.ErrNotFound {
			return nil, errSentCheque
		}
		errReceivedCheque := s.store.Get(receivedChequeKey(peer), &receivedCheque)
		if errReceivedCheque != nil && errReceivedCheque != state.ErrNotFound {
			return nil, errReceivedCheque
		}
	}

	return map[string]*Cheque{sentChequeResponseKey: sentCheque, receivedChequeResponseKey: receivedCheque}, nil
}

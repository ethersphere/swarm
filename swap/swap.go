// Copyright 2019 The go-ethereum Authors
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

package swap

import (
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/contracts/swap"
	cswap "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
)

// ErrInvalidChequeSignature indicates the signature on the cheque was invalid
var ErrInvalidChequeSignature = errors.New("invalid cheque signature")

// Swap represents the SwAP Swarm Accounting Protocol
// a peer to peer micropayment system
// A node maintains an individual balance with every peer
// Only messages which have a price will be accounted for
type Swap struct {
	api                 PublicAPI
	stateStore          state.Store          // stateStore is needed in order to keep balances across sessions
	lock                sync.RWMutex         // lock the balances
	balances            map[enode.ID]int64   // map of balances for each peer
	cheques             map[enode.ID]*Cheque // map of balances for each peer
	peers               map[enode.ID]*Peer   // map of all swap Peers
	backend             cswap.Backend        // the backend (blockchain) used
	owner               *Owner               // contract access
	params              *Params              // economic and operational parameters
	contractReference   *swap.Swap           // reference to the smart contract
	oracle              PriceOracle          // the oracle providing the ether price for honey
	paymentThreshold    int64                // balance difference required for requesting cheque
	disconnectThreshold int64                // balance difference required for dropping peer
}

// Owner encapsulates information related to accessing the contract
type Owner struct {
	Contract   common.Address    // address of swap contract
	address    common.Address    // owner address
	privateKey *ecdsa.PrivateKey // private key
	publicKey  *ecdsa.PublicKey  // public key
}

// Params encapsulates param
type Params struct {
	InitialDepositAmount uint64 //
}

// NewDefaultParams returns a Params struct filled with default values
func NewDefaultParams() *Params {
	return &Params{
		InitialDepositAmount: DefaultInitialDepositAmount,
	}
}

// New - swap constructor
func New(stateStore state.Store, prvkey *ecdsa.PrivateKey, contract common.Address, backend cswap.Backend) *Swap {
	sw := &Swap{
		stateStore:          stateStore,
		balances:            make(map[enode.ID]int64),
		backend:             backend,
		cheques:             make(map[enode.ID]*Cheque),
		peers:               make(map[enode.ID]*Peer),
		params:              NewDefaultParams(),
		paymentThreshold:    DefaultPaymentThreshold,
		disconnectThreshold: DefaultDisconnectThreshold,
		contractReference:   nil,
		oracle:              NewPriceOracle(),
	}
	sw.owner = sw.createOwner(prvkey, contract)
	return sw
}

// createOwner assings keys and addresses
func (s *Swap) createOwner(prvkey *ecdsa.PrivateKey, contract common.Address) *Owner {
	pubkey := &prvkey.PublicKey
	return &Owner{
		privateKey: prvkey,
		publicKey:  pubkey,
		Contract:   contract,
		address:    crypto.PubkeyToAddress(*pubkey),
	}
}

// DeploySuccess is for convenience log output
func (s *Swap) DeploySuccess() string {
	return fmt.Sprintf("contract: %s, owner: %s, deposit: %v, signer: %x", s.owner.Contract.Hex(), s.owner.address.Hex(), s.params.InitialDepositAmount, s.owner.publicKey)
}

// Add is the (sole) accounting function
// Swap implements the protocols.Balance interface
func (s *Swap) Add(amount int64, peer *protocols.Peer) (err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// load existing balances from the state store
	err = s.loadState(peer)
	if err != nil && err != state.ErrNotFound {
		log.Error("error while loading balance for peer", "peer", peer.ID().String())
		return
	}

	// check if balance with peer is over the disconnect threshold
	if s.balances[peer.ID()] >= s.disconnectThreshold {
		// if so, return error in order to abort the transfer
		disconnectMessage := fmt.Sprintf("balance for peer %s is over the disconnect threshold %v, disconnecting", peer.ID().String(), s.disconnectThreshold)
		log.Warn(disconnectMessage)
		return errors.New(disconnectMessage)
	}

	// calculate new balance
	var newBalance int64
	newBalance, err = s.updateBalance(peer.ID(), amount)
	if err != nil {
		return
	}

	// check if balance with peer crosses the threshold
	// it is the peer with a negative balance who sends a cheque, thus we check
	// that the balance is *below* the threshold
	if newBalance <= -s.paymentThreshold {
		//if so, send cheque
		log.Warn(fmt.Sprintf("balance for peer %s went over the payment threshold %v, sending cheque", peer.ID().String(), s.paymentThreshold))
		err = s.sendCheque(peer.ID())
		if err != nil {
			log.Error(fmt.Sprintf("error while sending cheque to peer %s: %s", peer.ID().String(), err.Error()))
		} else {
			log.Info(fmt.Sprintf("successfully sent cheque to peer %s", peer.ID().String()))
		}
	}

	return
}

func (s *Swap) updateBalance(peer enode.ID, amount int64) (int64, error) {
	//adjust the balance
	//if amount is negative, it will decrease, otherwise increase
	s.balances[peer] += amount
	//save the new balance to the state store
	peerBalance := s.balances[peer]
	err := s.stateStore.Put(peer.String(), &peerBalance)
	if err != nil {
		log.Error("error while storing balance for peer", "peer", peer.String())
	}
	log.Debug("balance for peer after accounting", "peer", peer.String(), "balance", strconv.FormatInt(peerBalance, 10))
	return peerBalance, err
}

// logBalance is a helper function to log the current balance of a peer
func (s *Swap) logBalance(peer *protocols.Peer) {
	err := s.loadState(peer)
	if err != nil && err != state.ErrNotFound {
		log.Error(fmt.Sprintf("error while loading balance for peer %s", peer.String()))
	} else {
		log.Info(fmt.Sprintf("balance for peer %s is %d", peer.ID(), s.balances[peer.ID()]))
	}
}

// sendCheque sends a cheque to peer
func (s *Swap) sendCheque(peer enode.ID) error {
	swapPeer := s.getPeer(peer)
	cheque, err := s.createCheque(peer)
	if err != nil {
		log.Error("error while creating cheque: %s", err.Error())
		return err
	}

	log.Info(fmt.Sprintf("sending cheque with serial %d, amount %d, benficiary %v, contract %v", cheque.ChequeParams.Serial, cheque.ChequeParams.Amount, cheque.Beneficiary, cheque.Contract))
	s.cheques[peer] = cheque

	err = s.stateStore.Put(peer.String()+"_cheques", &cheque)
	// TODO: error handling might be quite more complex
	if err != nil {
		log.Error("error while storing the last cheque: %s", err.Error())
		return err
	}

	emit := &EmitChequeMsg{
		Cheque: cheque,
	}

	// reset balance;
	// TODO: if sending fails it should actually be roll backed...
	s.resetBalance(peer, int64(cheque.Amount))

	err = swapPeer.Send(context.TODO(), emit)
	return err
}

// Create a Cheque structure emitted to a specific peer as a beneficiary
// The serial and amount of the cheque will depend on the last cheque and current balance for this peer
// The cheque will be signed and point to the issuer's contract
func (s *Swap) createCheque(peer enode.ID) (*Cheque, error) {
	var cheque *Cheque
	var err error

	swapPeer := s.getPeer(peer)
	beneficiary := swapPeer.beneficiary

	peerBalance := s.balances[peer]
	// the balance should be negative here, we take the absolute value:
	honey := uint64(-peerBalance)

	// convert honey to ETH
	var amount uint64
	amount, err = s.oracle.GetPrice(honey)
	if err != nil {
		log.Error("error getting price from oracle", "err", err)
		return nil, err
	}

	// we need to ignore the error check when loading from the StateStore,
	// as an error might indicate that there is no existing cheque, which
	// could mean it's the first interaction, which is absolutely valid
	_ = s.loadCheque(peer)
	lastCheque := s.cheques[peer]

	if lastCheque == nil {
		cheque = &Cheque{
			ChequeParams: ChequeParams{
				Serial: uint64(1),
				Amount: uint64(amount),
			},
		}
	} else {
		cheque = &Cheque{
			ChequeParams: ChequeParams{
				Serial: lastCheque.Serial + 1,
				Amount: lastCheque.Amount + uint64(amount),
			},
		}
	}
	cheque.ChequeParams.Timeout = defaultCashInDelay
	cheque.ChequeParams.Contract = s.owner.Contract
	cheque.ChequeParams.Honey = uint64(honey)
	cheque.Beneficiary = beneficiary

	cheque.Sig, err = s.signContent(cheque)

	return cheque, err
}

// GetPeerBalance returns the balance for a given peer
func (s *Swap) GetPeerBalance(peer enode.ID) (int64, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if p, ok := s.balances[peer]; ok {
		return p, nil
	}
	return 0, errors.New("Peer not found")
}

// GetLastCheque returns the last cheque for a given peer
func (s *Swap) GetLastCheque(peer enode.ID) (*Cheque, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if lc, ok := s.cheques[peer]; ok {
		return lc, nil
	}

	return nil, errors.New("Peer not found")
}

//GetAllBalances returns the balances for all known peers
func (s *Swap) GetAllBalances() map[enode.ID]int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.balances
}

// loadStates loads balances from the state store (persisted)
func (s *Swap) loadState(peer *protocols.Peer) (err error) {
	var peerBalance int64
	peerID := peer.ID()
	//only load if the current instance doesn't already have this peer's
	//balance in memory
	if _, ok := s.balances[peerID]; !ok {
		err = s.stateStore.Get(peerID.String(), &peerBalance)
		s.balances[peerID] = peerBalance
	}
	return
}

//loadCheque loads the last cheque for a peer from the state store (persisted)
func (s *Swap) loadCheque(peer enode.ID) (err error) {
	//only load if the current instance doesn't already have this peer's
	//last cheque in memory
	var cheque *Cheque
	if _, ok := s.cheques[peer]; !ok {
		err = s.stateStore.Get(peer.String()+"_cheques", &cheque)
		s.cheques[peer] = cheque
	}
	return
}

// saveLastReceivedCheque loads the last received cheque for peer
func (s *Swap) loadLastReceivedCheque(peer enode.ID) (cheque *Cheque) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.stateStore.Get(peer.String()+"_cheques_received", &cheque)
	return
}

// saveLastReceivedCheque saves cheque as the last received cheque for peer
func (s *Swap) saveLastReceivedCheque(peer enode.ID, cheque *Cheque) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.stateStore.Put(peer.String()+"_cheques_received", cheque)
}

// Close cleans up swap
func (s *Swap) Close() {
	s.stateStore.Close()
}

// resetBalance is called:
// * for the creditor: on cheque receival
// * for the debitor: on confirmation receival
func (s *Swap) resetBalance(peerID enode.ID, amount int64) {
	log.Info("resetting balance for peer", "peer", peerID.String(), "amount", amount)
	s.updateBalance(peerID, amount)
}

// encodeCheque encodes the cheque in the format used in the signing procedure
func (s *Swap) encodeCheque(cheque *Cheque) []byte {
	serialBytes := make([]byte, 32)
	amountBytes := make([]byte, 32)
	timeoutBytes := make([]byte, 32)
	// we need to write the last 8 bytes as we write a uint64 into a 32-byte array
	// encoded in BigEndian because EVM uses BigEndian encoding
	binary.BigEndian.PutUint64(serialBytes[24:], cheque.Serial)
	binary.BigEndian.PutUint64(amountBytes[24:], cheque.Amount)
	binary.BigEndian.PutUint64(timeoutBytes[24:], cheque.Timeout)
	// construct the actual cheque
	input := cheque.Contract.Bytes()
	input = append(input, cheque.Beneficiary.Bytes()...)
	input = append(input, serialBytes[:]...)
	input = append(input, amountBytes[:]...)
	input = append(input, timeoutBytes[:]...)

	return input
}

// sigHashCheque hashes the cheque using the prefix that would be added by eth_Sign
func (s *Swap) sigHashCheque(cheque *Cheque) []byte {
	input := crypto.Keccak256(s.encodeCheque(cheque))
	withPrefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(input), input)
	return crypto.Keccak256([]byte(withPrefix))
}

// verifyChequeSig verifies the signature on the cheque
func (s *Swap) verifyChequeSig(cheque *Cheque, expectedSigner common.Address) error {
	sigHash := s.sigHashCheque(cheque)

	if cheque.Sig == nil {
		return fmt.Errorf("tried to verify signature on cheque with sig nil")
	}
	// copy signature to avoid modifying the original
	sig := make([]byte, len(cheque.Sig))
	copy(sig, cheque.Sig)
	// reduce the v value of the signature by 27 (see signContent)
	sig[len(sig)-1] -= 27
	pubKey, err := crypto.SigToPub(sigHash, sig)
	if err != nil {
		return err
	}

	if crypto.PubkeyToAddress(*pubKey) != expectedSigner {
		return ErrInvalidChequeSignature
	}

	return nil
}

// signContent signs the cheque with supplied private key
func (s *Swap) signContentWithKey(cheque *Cheque, prv *ecdsa.PrivateKey) ([]byte, error) {
	sig, err := crypto.Sign(s.sigHashCheque(cheque), prv)
	if err != nil {
		return nil, err
	}
	// increase the v value by 27 as crypto.Sign produces 0 or 1 but the contract only accepts 27 or 28
	// this is to prevent malleable signatures. while not strictly necessary in this case the ECDSA implementation from Openzeppelin expects it.
	sig[len(sig)-1] += 27
	return sig, nil
}

// signContent signs the cheque with the owners private key
func (s *Swap) signContent(cheque *Cheque) ([]byte, error) {
	return s.signContentWithKey(cheque, s.owner.privateKey)
}

// GetParams returns contract parameters (Bin, ABI) from the contract
func (s *Swap) GetParams() *swap.Params {
	return s.contractReference.ContractParams()
}

// Deploy deploys a new swap contract
func (s *Swap) Deploy(ctx context.Context, backend swap.Backend, path string) error {
	// TODO: What to do if the contract is already deployed?
	return s.deploy(ctx, backend, path)
}

// verifyContract checks if the bytecode found at address matches the expected bytecode
func (s *Swap) verifyContract(ctx context.Context, address common.Address) error {
	swap, err := swap.InstanceAt(address, s.backend)
	if err != nil {
		return err
	}

	return swap.ValidateCode(ctx, s.backend, address)
}

// getContractOwner retrieve the owner of the chequebook at address from the blockchain
func (s *Swap) getContractOwner(ctx context.Context, address common.Address) (common.Address, error) {
	swap, err := swap.InstanceAt(address, s.backend)
	if err != nil {
		return common.Address{}, err
	}

	return swap.Instance.Issuer(nil)
}

// deploy deploys the Swap contract
func (s *Swap) deploy(ctx context.Context, backend swap.Backend, path string) error {
	opts := bind.NewKeyedTransactor(s.owner.privateKey)
	// initial topup value
	opts.Value = big.NewInt(int64(s.params.InitialDepositAmount))
	opts.Context = ctx

	log.Info(fmt.Sprintf("Deploying new swap (owner: %v)", opts.From.Hex()))
	address, err := s.deployLoop(opts, backend, s.owner.address, defaultHarddepositTimeoutDuration)
	if err != nil {
		log.Error(fmt.Sprintf("unable to deploy swap: %v", err))
		return err
	}
	s.owner.Contract = address
	log.Info(fmt.Sprintf("swap deployed at %v (owner: %v)", address.Hex(), opts.From.Hex()))

	return err
}

// deployLoop repeatedly tries to deploy the swap contract .
func (s *Swap) deployLoop(opts *bind.TransactOpts, backend swap.Backend, owner common.Address, defaultHarddepositTimeoutDuration time.Duration) (addr common.Address, err error) {
	var tx *types.Transaction
	for try := 0; try < deployRetries; try++ {
		if try > 0 {
			time.Sleep(deployDelay)
		}

		if _, s.contractReference, tx, err = swap.Deploy(opts, backend, owner, defaultHarddepositTimeoutDuration); err != nil {
			log.Warn(fmt.Sprintf("can't send chequebook deploy tx (try %d): %v", try, err))
			continue
		}
		if addr, err = bind.WaitDeployed(opts.Context, backend, tx); err != nil {
			log.Warn(fmt.Sprintf("chequebook deploy error (try %d): %v", try, err))
			continue
		}
		return addr, nil
	}
	return addr, err
}

package simulation

import (
	"errors"
	"net/http"
	"net/rpc"
)

type Simulation struct {
	adapter *Adapter
	nodes   map[string]*Node
}

func NewSimulation(adapter *Adapter) (*Simulation, error) {
	return nil, errors.New("not implemented")
}
func NewSimulationFromSnapshot(adapter *Adapter, snapshot *NetworkSnapshot) (*Simulation, error) {
	return nil, errors.New("not implemented")
}

func (s *Simulation) Get(id string) (*Node, error) {
	return nil, errors.New("not implemented")
}
func (s *Simulation) GetAll() ([]*Node, error) {
	return nil, errors.New("not implemented")
}

func (s *Simulation) Create(id string, config NodeConfig) error {
	return errors.New("not implemented")
}
func (s *Simulation) Start(id string) error {
	return errors.New("not implemented")
}
func (s *Simulation) Stop(id string) error {
	return errors.New("not implemented")
}
func (s *Simulation) StartAll() error {
	return errors.New("not implemented")
}
func (s *Simulation) StopAll() error {
	return errors.New("not implemented")
}

func (s *Simulation) RPCClient(id string) (*rpc.Client, error) {
	return nil, errors.New("not implemented")
}
func (s *Simulation) HTTPClient(id string) (*http.Client, error) {
	return nil, errors.New("not implemented")
}

func (s *Simulation) Snapshot() (NetworkSnapshot, error) {
	snap := NetworkSnapshot{}
	return snap, errors.New("not implemented")
}

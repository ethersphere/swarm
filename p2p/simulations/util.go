package simulations

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/p2p/discover"
)

const (
	UpModeStar = iota
	UpModeCircle
	UpModeCluster
)

type up struct {
	net    *Network
	nids   []discover.NodeID
	sim    *Simulation
	conns  []*Conn
	events chan *Event
	ctx    context.Context
}

func Up(ctx context.Context, net *Network, nids []discover.NodeID, mode int) error {

	eventC := make(chan *Event)
	defer close(eventC)
	u := &up{
		nids:   nids,
		net:    net,
		sim:    NewSimulation(net),
		ctx:    ctx,
		events: eventC,
	}

	switch mode {
	case UpModeStar:
		u.conns = modeStar(nids)
	case UpModeCircle:
		u.conns = modeCircle(nids)
	}

	err := u.upWithConfig()
	if err != nil {
		return err
	}
	return u.connWithConfig()
}

func (u *up) upWithConfig() error {
	quitC := make(chan struct{})
	trigger := make(chan discover.NodeID)
	sub := u.net.Events().Subscribe(u.events)
	// event sink on quit
	defer func() {
		sub.Unsubscribe()
		close(quitC)
		select {
		case <-u.events:
		default:
		}
		return
	}()
	action := func(ctx context.Context) error {
		go func() {
			for {
				select {
				case e := <-u.events:
					if e.Type == EventTypeNode {
						if e.Node.Up {
							trigger <- e.Node.ID()
						}
					}
				case <-ctx.Done():
					return
				case <-quitC:
					return
				}
			}
		}()
		go func() {
			for _, n := range u.nids {
				err := u.net.Start(n)
				if err != nil {
					return
				}
			}
		}()
		return nil
	}
	check := func(ctx context.Context, nid discover.NodeID) (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		return true, nil
	}

	d, err := time.ParseDuration(fmt.Sprintf("%dms", 10*len(u.nids)))
	if err != nil {
		return err
	}
	localctx, cancel := context.WithTimeout(u.ctx, d)
	defer cancel()

	step := u.sim.Run(localctx, &Step{
		Action:  action,
		Trigger: trigger,
		Expect: &Expectation{
			Nodes: u.nids,
			Check: check,
		},
	})
	return step.Error
}

func (u *up) connWithConfig() error {
	quitC := make(chan struct{})
	trigger := make(chan discover.NodeID)
	sub := u.net.Events().Subscribe(u.events)
	// event sink on quit
	defer func() {
		sub.Unsubscribe()
		close(quitC)
		select {
		case <-u.events:
		default:
		}
		return
	}()
	action := func(ctx context.Context) error {
		go func() {
			for {
				select {
				case e := <-u.events:
					if e.Type == EventTypeConn {
						if e.Conn.Up {
							trigger <- e.Conn.One
							trigger <- e.Conn.Other
						}
					}
				case <-ctx.Done():
					return
				case <-quitC:
					return
				}
			}
		}()
		go func() {
			for _, n := range u.conns {
				err := u.net.Connect(n.One, n.Other)
				if err != nil {
					return
				}
			}
		}()
		return nil
	}
	check := func(ctx context.Context, nid discover.NodeID) (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		return true, nil
	}

	d, err := time.ParseDuration(fmt.Sprintf("%dms", 10*len(u.nids)))
	if err != nil {
		return err
	}
	localctx, cancel := context.WithTimeout(u.ctx, d)
	defer cancel()

	step := u.sim.Run(localctx, &Step{
		Action:  action,
		Trigger: trigger,
		Expect: &Expectation{
			Nodes: u.nids,
			Check: check,
		},
	})
	return step.Error
}

func modeStar(nids []discover.NodeID) (conns []*Conn) {
	for i, n := range nids {
		if i > 0 {
			conns = append(conns, &Conn{
				One:   n,
				Other: nids[0],
			})
		}
	}
	return
}

func modeCircle(nids []discover.NodeID) (conns []*Conn) {
	for i, n := range nids {
		var otherIdx int
		if i == 0 {
			otherIdx = len(nids) - 1
		} else {
			otherIdx = i - 1
		}
		conns = append(conns, &Conn{
			One:   n,
			Other: nids[otherIdx],
		})
	}
	return
}

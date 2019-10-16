// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.
package network

import (
	"sort"
	"strconv"
	"sync"

	"github.com/ethersphere/swarm/log"
)

// resourceUseStats can be used to count uses of resources. A Resource is anything with a Key()
type resourceUseStats struct {
	resourceUses map[string]int
	waiting      map[string]chan struct{}
	lock         sync.RWMutex
	quitC        chan struct{}
}

// Resource represents anything with a Key that can be accounted with some stat.
type Resource interface {
	Key() string   // unique id in string format of the resource.
	Label() string // short string format of the key for debugging purposes.
}

type ResourceCount struct {
	resource Resource
	count    int
}

func newResourceUseStats(quitC chan struct{}) *resourceUseStats {
	return &resourceUseStats{
		resourceUses: make(map[string]int),
		waiting:      make(map[string]chan struct{}),
		quitC:        quitC,
	}
}

func (lb *resourceUseStats) sortResources(resources []Resource) []Resource {
	sorted := make([]Resource, len(resources))
	resourceCounts := lb.getAllUseCounts(resources)
	sort.Slice(resourceCounts, func(i, j int) bool {
		return resourceCounts[i].count < resourceCounts[j].count
	})
	for i, resourceCount := range resourceCounts {
		sorted[i] = resourceCount.resource
	}
	return sorted
}

func (lbp ResourceCount) String() string {
	return lbp.resource.Key() + ":" + strconv.Itoa(lbp.count)
}

func (lb *resourceUseStats) dumpAllUses() map[string]int {
	lb.lock.RLock()
	defer lb.lock.RUnlock()
	dump := make(map[string]int)
	for k, v := range lb.resourceUses {
		dump[k] = v
	}
	return dump
}

func (lb *resourceUseStats) getAllUseCounts(resources []Resource) []ResourceCount {
	peerUses := make([]ResourceCount, len(resources))
	for i, resource := range resources {
		peerUses[i] = ResourceCount{
			resource: resource,
			count:    lb.getUses(resource),
		}
	}
	return peerUses
}

func (lb *resourceUseStats) getUses(keyed Resource) int {
	return lb.getKeyUses(keyed.Key())
}

func (lb *resourceUseStats) getKeyUses(key string) int {
	lb.lock.RLock()
	defer lb.lock.RUnlock()
	return lb.resourceUses[key]
}

func (lb *resourceUseStats) addUse(resource Resource) int {
	lb.lock.Lock()
	defer lb.lock.Unlock()
	log.Debug("Adding use", "key", resource.Label())
	key := resource.Key()
	lb.resourceUses[key] = lb.resourceUses[key] + 1
	return lb.resourceUses[key]
}

// waitKey blocks until some key is added to the load balancer stats.
// As peer resource initialization is asynchronous we need a way to know that the initial uses has been initialized.
func (lb *resourceUseStats) waitKey(key string) {
	lb.lock.Lock()
	if _, ok := lb.resourceUses[key]; ok {
		lb.lock.Unlock()
		return
	}
	waitChan := make(chan struct{})
	lb.waiting[key] = waitChan
	lb.lock.Unlock()
	select {
	case <-waitChan:
		delete(lb.waiting, key)
	case <-lb.quitC:
		return
	}
}

func (lb *resourceUseStats) initKey(key string, count int) {
	lb.lock.Lock()
	defer lb.lock.Unlock()
	lb.resourceUses[key] = count
	if kChan, ok := lb.waiting[key]; ok {
		kChan <- struct{}{}
	}
}

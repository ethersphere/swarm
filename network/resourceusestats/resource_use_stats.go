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
package resourceusestats

import (
	"sort"
	"strconv"
	"sync"
)

// ResourceUseStats can be used to count uses of resources. A Resource is anything with a Key()
type ResourceUseStats struct {
	resourceUses map[string]int
	waiting      map[string]chan struct{}
	lock         sync.RWMutex
	quitC        <-chan struct{}
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

func NewResourceUseStats(quitC <-chan struct{}) *ResourceUseStats {
	return &ResourceUseStats{
		resourceUses: make(map[string]int),
		waiting:      make(map[string]chan struct{}),
		quitC:        quitC,
	}
}

func (lb *ResourceUseStats) SortResources(resources []Resource) []Resource {
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

func (lb *ResourceUseStats) Len() int {
	lb.lock.RLock()
	defer lb.lock.RUnlock()
	return len(lb.resourceUses)
}

func (lb *ResourceUseStats) DumpAllUses() map[string]int {
	lb.lock.RLock()
	defer lb.lock.RUnlock()
	dump := make(map[string]int)
	for k, v := range lb.resourceUses {
		dump[k] = v
	}
	return dump
}

func (lb *ResourceUseStats) getAllUseCounts(resources []Resource) []ResourceCount {
	lb.lock.RLock()
	defer lb.lock.RUnlock()
	peerUses := make([]ResourceCount, len(resources))
	for i, resource := range resources {
		peerUses[i] = ResourceCount{
			resource: resource,
			count:    lb.resourceUses[resource.Key()],
		}
	}
	return peerUses
}

func (lb *ResourceUseStats) GetUses(keyed Resource) int {
	return lb.GetKeyUses(keyed.Key())
}

func (lb *ResourceUseStats) GetKeyUses(key string) int {
	lb.lock.RLock()
	defer lb.lock.RUnlock()
	return lb.resourceUses[key]
}

func (lb *ResourceUseStats) AddUse(resource Resource) int {
	lb.lock.Lock()
	defer lb.lock.Unlock()
	key := resource.Key()
	prevCount := lb.resourceUses[key]
	lb.resourceUses[key] = prevCount + 1
	return lb.resourceUses[key]
}

// WaitKey blocks until some key is added to the load balancer stats.
// As peer resource initialization is asynchronous we need a way to know that the initial uses has been initialized.
func (lb *ResourceUseStats) WaitKey(key string) {
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
	}
}

func (lb *ResourceUseStats) InitKey(key string, count int) {
	lb.lock.Lock()
	defer lb.lock.Unlock()
	lb.resourceUses[key] = count
	if kChan, ok := lb.waiting[key]; ok {
		select {
		case <-lb.quitC:
		case kChan <- struct{}{}:
		}

	}
}

func (lb *ResourceUseStats) RemoveKey(key string) {
	lb.lock.Lock()
	defer lb.lock.Unlock()
	delete(lb.resourceUses, key)
}

func (lb *ResourceUseStats) RemoveResource(resource Resource) {
	lb.lock.Lock()
	defer lb.lock.Unlock()
	delete(lb.resourceUses, resource.Key())
}

package kademlia

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
)

const (
	minBucketSize = 1
	bucketSize    = 20
	maxProx       = 255
	connRetryExp  = 2
)

var (
	purgeInterval        = 42 * time.Hour
	initialRetryInterval = 4 * time.Second
)

type Kademlia struct {
	// immutable baseparam
	addr Address

	// adjustable parameters
	MaxProx              int
	ProxBinSize          int
	BucketSize           int
	MinBucketSize        int
	PurgeInterval        time.Duration
	InitialRetryInterval time.Duration
	ConnRetryExp         int

	nodeDB    [][]*NodeRecord
	nodeIndex map[Address]*NodeRecord

	GetNode func(int)

	// state
	proxLimit int
	proxSize  int

	//
	count   int
	buckets []*bucket

	dblock sync.RWMutex
	lock   sync.RWMutex
	quitC  chan bool
}

type Address common.Hash

func (a Address) String() string {
	return fmt.Sprintf("%x", a[:])
}

type Node interface {
	Addr() Address
	Url() string
	LastActive() time.Time
	Drop()
}

type NodeRecord struct {
	Addr   Address `json:address`
	Url    string  `json:url`
	Active int64   `json:active`
	After  int64   `json:after`
	after  time.Time
	node   Node
}

func (self *NodeRecord) setActive() {
	if self.node != nil {
		self.Active = self.node.LastActive().Unix()
	}
}

type kadDB struct {
	Address Address         `json:address`
	Nodes   [][]*NodeRecord `json:nodes`
}

// public constructor
// hash is a byte slice of length equal to self.HashBytes
func New() *Kademlia {
	return &Kademlia{}
}

// accessor for KAD self address
func (self *Kademlia) Addr() Address {
	return self.addr
}

// accessor for KAD self count
func (self *Kademlia) Count() (sum int) {
	for _, b := range self.buckets {
		sum += len(b.nodes)
	}
	return
	// return self.count
}

// accessor for KAD offline db count
func (self *Kademlia) DBCount() int {
	return len(self.nodeIndex)
}

func (self *Kademlia) String() string {
	var rows []string
	// rows = append(rows, fmt.Sprintf("KΛÐΞMLIΛ basenode address: %064x\n population: %d (%d)", self.addr[:], self.Count(), self.DBCount()))
	rows = append(rows, "====================================================================")
	rows = append(rows, fmt.Sprintf("%v : MaxProx: %d, ProxBinSize: %d, BucketSize: %d, MinBucketSize: %d, proxLimit: %d, proxSize: %d", time.Now(), self.MaxProx, self.ProxBinSize, self.BucketSize, self.MinBucketSize, self.proxLimit, self.proxSize))

	for i, b := range self.buckets {

		if i == self.proxLimit {
			rows = append(rows, fmt.Sprintf("===================== PROX LIMIT: %d =====================", i))
		}
		row := []string{fmt.Sprintf("%03d", i), fmt.Sprintf("%2d", len(b.nodes))}
		var k int
		for _, p := range b.nodes {
			row = append(row, fmt.Sprintf("%s", p.Addr().String()[:8]))
			if k == 4 {
				break
			}
			k++
		}
		for ; k < 5; k++ {
			row = append(row, "        ")
		}
		row = append(row, fmt.Sprintf("| %2d %2d", len(self.nodeDB[i]), b.dbcursor))

		for j, p := range self.nodeDB[i] {
			row = append(row, fmt.Sprintf("%08x", p.Addr[:4]))
			if j == 4 {
				break
			}
		}
		rows = append(rows, strings.Join(row, " "))
		if i == self.MaxProx {
			break
		}
	}
	rows = append(rows, "====================================================================")

	return strings.Join(rows, "\n")
}

// Start brings up a pool of entries potentially from an offline persisted source
// and sets default values for optional parameters
func (self *Kademlia) Start(addr Address) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	if self.quitC != nil {
		return nil
	}
	self.addr = addr
	if self.MaxProx == 0 {
		self.MaxProx = maxProx
	}
	if self.BucketSize == 0 {
		self.BucketSize = bucketSize
	}
	if self.MinBucketSize == 0 {
		self.MinBucketSize = minBucketSize
	}
	if self.InitialRetryInterval == 0 {
		self.InitialRetryInterval = initialRetryInterval
	}
	if self.PurgeInterval == 0 {
		self.PurgeInterval = purgeInterval
	}
	if self.ConnRetryExp == 0 {
		self.ConnRetryExp = connRetryExp
	}
	// runtime parameters
	if self.ProxBinSize == 0 {
		self.ProxBinSize = self.BucketSize
	}

	self.buckets = make([]*bucket, self.MaxProx+1)
	for i, _ := range self.buckets {
		self.buckets[i] = &bucket{size: self.BucketSize} // will initialise bucket{int(0),[]Node(nil),sync.Mutex}
	}

	self.nodeDB = make([][]*NodeRecord, 8*len(self.addr))
	self.nodeIndex = make(map[Address]*NodeRecord)

	self.quitC = make(chan bool)
	glog.V(logger.Info).Infof("[KΛÐ] started")
	return nil
}

// Stop saves the routing table into a persistant form
func (self *Kademlia) Stop(path string) (err error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if self.quitC == nil {
		return
	}
	close(self.quitC)
	self.quitC = nil

	if len(path) > 0 {
		err = self.Save(path)
		if err != nil {
			glog.V(logger.Warn).Infof("[KΛÐ]: unable to save node records: %v", err)
		} else {
			glog.V(logger.Info).Infof("[KΛÐ]: node records saved to '%v'", path)
		}
	}
	return
}

// RemoveNode is the entrypoint where nodes are taken offline
func (self *Kademlia) RemoveNode(node Node) (err error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	var found bool
	index := self.proximityBin(node.Addr())
	bucket := self.buckets[index]
	for i := 0; i < len(bucket.nodes); i++ {
		if node.Addr() == bucket.nodes[i].Addr() {
			found = true
			bucket.nodes = append(bucket.nodes[:i], bucket.nodes[(i+1):]...)
		}
	}

	if found {
		glog.V(logger.Info).Infof("[KΛÐ]: remove node %v from table", node)

		self.count--
		if len(bucket.nodes) < bucket.size {
			err = fmt.Errorf("insufficient nodes (%v) in bucket %v", len(bucket.nodes), index)
		}
		self.adjustProxLess(index)
		// async callback to notify user that bucket needs filling
		// action is left to the user
		if self.GetNode != nil {
			go self.GetNode(index)
		}
		r := self.nodeIndex[node.Addr()]
		r.node = nil
		now := time.Now()
		r.after = now
		r.After = now.Unix()
		r.Active = now.Unix()

	}

	return
}

// AddNode is the entry point where new nodes are registered
func (self *Kademlia) AddNode(node Node) (err error) {

	self.lock.Lock()
	defer self.lock.Unlock()

	index := self.proximityBin(node.Addr())

	bucket := self.buckets[index]
	// if bucket is full insertion replaces the worst node
	// TODO probably should give priority to peers with active traffic
	if worst, pos := bucket.insert(node); worst != nil {
		glog.V(logger.Info).Infof("[KΛÐ]: replace node %v (%d) with node %v", worst, pos, node)
		// no prox adjustment needed
		// do not change count
	} else {
		glog.V(logger.Info).Infof("[KΛÐ]: add new node %v to table", node)
		self.count++
		self.adjustProxMore(index)
	}

	self.dblock.Lock()
	defer self.dblock.Unlock()
	record, found := self.nodeIndex[node.Addr()]
	if !found {
		glog.V(logger.Info).Infof("[KΛÐ]: add new record %v to node db", node)
		record = &NodeRecord{
			Addr: node.Addr(),
			Url:  node.Url(),
		}
		self.nodeIndex[node.Addr()] = record
		self.nodeDB[index] = append(self.nodeDB[index], record)
	}
	record.node = node
	record.setActive()

	return

}

// adjust Prox (proxLimit and proxSize after an insertion of add nodes into bucket r)
func (self *Kademlia) adjustProxMore(r int) {
	if r >= self.proxLimit {
		exLimit := self.proxLimit
		exSize := self.proxSize
		self.proxSize++

		var i int
		for i = self.proxLimit; i < self.MaxProx && len(self.buckets[i].nodes) > 0 && self.proxSize-len(self.buckets[i].nodes) > self.ProxBinSize; i++ {
			self.proxSize -= len(self.buckets[i].nodes)
		}
		self.proxLimit = i

		glog.V(logger.Detail).Infof("[KΛÐ]: Max Prox Bin: Lower Limit: %v (was %v): Bin Size: %v (was %v)", self.proxLimit, exLimit, self.proxSize, exSize)
	}
}

func (self *Kademlia) adjustProxLess(r int) {
	exLimit := self.proxLimit
	exSize := self.proxSize
	if r >= self.proxLimit {
		self.proxSize--
	}

	if r < self.proxLimit && len(self.buckets[r].nodes) == 0 {
		for i := self.proxLimit - 1; i > r; i-- {
			self.proxSize += len(self.buckets[i].nodes)
		}
		self.proxLimit = r
	} else if self.proxLimit > 0 && r >= self.proxLimit-1 {
		var i int
		for i = self.proxLimit - 1; i > 0 && len(self.buckets[i].nodes)+self.proxSize <= self.ProxBinSize; i-- {
			self.proxSize += len(self.buckets[i].nodes)
		}
		self.proxLimit = i
	}

	if exLimit != self.proxLimit || exSize != self.proxSize {
		glog.V(logger.Detail).Infof("[KΛÐ]: Max Prox Bin: Lower Limit: %v (was %v): Bin Size: %v (was %v)", self.proxLimit, exLimit, self.proxSize, exSize)
	}
}

/*
GetNodes(target) returns the list of nodes belonging to the same proximity bin
as the target. The most proximate bin will be the union of the bins between
proxLimit and MaxProx. proxLimit is dynamically adjusted so that 1) there is no
empty buckets in bin < proxLimit and 2) the sum of all items are the maximum
possible but lower than ProxBinSize
*/
func (self *Kademlia) GetNodes(target Address, max int) []Node {
	return self.getNodes(target, max).nodes
}

func (self *Kademlia) getNodes(target Address, max int) (r nodesByDistance) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	r.target = target
	index := self.proximityBin(target)
	start := index
	var down bool
	if index >= self.proxLimit {
		index = self.proxLimit
		start = self.MaxProx
		down = true
	}
	var n int
	limit := max
	if max == 0 {
		limit = 1000
	}
	for {
		bucket := self.buckets[start].nodes
		for i := 0; i < len(bucket); i++ {
			r.push(bucket[i], limit)
			n++
		}
		if max == 0 && start <= index && (n > 0 || start == 0) ||
			max > 0 && down && start <= index && (n >= limit || n == self.Count() || start == 0) {
			break
		}
		if down {
			start--
		} else {
			if start == self.MaxProx {
				if index == 0 {
					break
				}
				start = index - 1
				down = true
			} else {
				start++
			}
		}
	}
	glog.V(logger.Detail).Infof("[KΛÐ]: serve %d (=<%d) nodes for target lookup %v (PO%d)", n, self.MaxProx, target, index)
	return
}

// this is used to add node records to the persisted db
// TODO: maybe db needs to be purged occasionally (reputation will take care of
// that)
func (self *Kademlia) AddNodeRecords(nrs []*NodeRecord) {
	self.dblock.Lock()
	defer self.dblock.Unlock()
	var n int
	var nodes []*NodeRecord
	for _, node := range nrs {
		_, found := self.nodeIndex[node.Addr]
		if !found && node.Addr != self.addr {
			self.nodeIndex[node.Addr] = node
			index := self.proximityBin(node.Addr)
			dbcursor := self.buckets[index].dbcursor
			nodes = self.nodeDB[index]
			newnodes := make([]*NodeRecord, len(nodes)+1)
			copy(newnodes[:], nodes[:dbcursor])
			newnodes[dbcursor] = node
			copy(newnodes[dbcursor+1:], nodes[dbcursor:])
			self.nodeDB[index] = newnodes
			n++
		}
	}
	glog.V(logger.Detail).Infof("[KΛÐ]: received %d node records, added %d new", len(nrs), n)
}

/*
GetNodeRecord gives back a node record with the highest priority for desired
connection
Used to pick candidates for live nodes to satisfy Kademlia network for Swarm

if len(nodes) < MinBucketSize, then take ith element in corresponding
db row ordered by reputation (active time?)
node record a is more favoured to b a > b iff
|proxBin(a)| < |proxBin(b)|
|| proxBin(a) < proxBin(b) && |proxBin(a)| < MinBucketSize
|| lastActive(a) < lastActive(b)

This has double role. Starting as naive node with empty db, this implements
Kademlia bootstrapping
As a mature node, it manages quickly fill in blanks or short lines
All on demand.
*/
func (self *Kademlia) GetNodeRecord() (node *NodeRecord, proxLimit int) {
	proxLimit = -1
	self.dblock.RLock()
	defer self.dblock.RUnlock()
	for rounds := 1; rounds <= self.BucketSize; rounds++ {
		var toprow int
		for i, dbrow := range self.nodeDB {
			order := i
			// for prox orders higher than MaxProx aggregate row is used
			if order == self.MaxProx {
				for j := i; j < len(self.nodeDB); j++ {
					dbrow = append(dbrow, self.nodeDB[j]...)
				}
			}
			if order > self.MaxProx {
				break
			}
			bin := self.buckets[order]

			var n, count int
			if len(bin.nodes) < rounds {
				if proxLimit < 0 {
					proxLimit = i
				}
				var purge []int

				// try all db elements that are ripe for checking
				if len(dbrow) > 0 {
					n = bin.dbcursor
				ROW:
					for count < len(dbrow) {
						if n >= len(dbrow) {
							n = 0
							bin.dbcursor = 0
						}
						node = dbrow[n]
						bin.dbcursor = n + 1
						if (node.after == time.Time{}) {
							node.after = time.Unix(node.After, 0)
						}
						glog.V(logger.Detail).Infof("[KΛÐ]: kaddb record %v (PO%03d:%d) to be checked after %d %v", node.Addr, i, n, node.After, node.after)

						delta := time.Now().Unix() - node.After
						if delta < 4 {
							node.After = 0
						}
						if node.node == nil && node.after.Before(time.Now()) {
							if node.after.Add(self.PurgeInterval).Before(time.Now()) {
								// delete node
								purge = append(purge, n)
								glog.V(logger.Detail).Infof("[KΛÐ]: inactive node record %v (PO%03d:%d) next check: %v", node.Addr, i, n, node.after)
							} else {
								if node.After == 0 {
									node.after = time.Now().Add(self.InitialRetryInterval)
									node.After = node.after.Unix()
								} else {
									node.After = (delta)*int64(self.ConnRetryExp) + node.After
									node.after = time.Unix(node.After, 0)
								}
								glog.V(logger.Detail).Infof("[KΛÐ]: serve node record %v (PO%03d:%d) next check: %d %v", node.Addr, i, n, node.After, node.after)
							}
							break ROW
						} else {
							glog.V(logger.Detail).Infof("[KΛÐ]: kaddb record %v (PO%03d:%d) not ready skipped next check: %v", node.Addr, i, n, node.after)
						}
						n++
						count++
					}

					// purge record inactive for more that PurgeInterval
					var prev int
					var nodes []*NodeRecord
					for i, next := range purge {
						// need to adjust dbcursor
						if next > 0 {
							if next <= bin.dbcursor {
								bin.dbcursor--
							}
							nodes = append(nodes, dbrow[prev:next]...)
						}
						prev = i + 1
						delete(self.nodeIndex, dbrow[next].Addr)
					}
					self.nodeDB[order] = append(nodes, dbrow[prev:]...)

					if node != nil {
						break
					}
				} // if len < rounds
			} else { // if not full
				toprow = i
			}
			glog.V(logger.Detail).Infof("[KΛÐ]: round %d: proxlimit: PO%03d, toprow: PO%03d\n%v", rounds, proxLimit, toprow, node)
			// this means there was missing element already on level 1 but
			// couldnt find anything on higher levels either

		} //for row
		// if there is a missing element in any prox bin upto max then try request
		// but if no success (tried recently, or no known peers) then fall back to
		// filling in rows with redundant nodes starting from 0 upwards if a cycle
		// finishes with proxLimit
		if proxLimit <= 0 {
			return
		}
	} // for round

	return
}

// in situ mutable bucket
type bucket struct {
	dbcursor int
	size     int
	nodes    []Node
	lock     sync.RWMutex
}

func (a Address) Bin() string {
	var bs []string
	for _, b := range a[:] {
		bs = append(bs, fmt.Sprintf("%08b", b))
	}
	return strings.Join(bs, "")
}

// nodesByDistance is a list of nodes, ordered by distance to target.
type nodesByDistance struct {
	nodes  []Node
	target Address
}

func sortedByDistanceTo(target Address, slice []Node) bool {
	var last Address
	for i, node := range slice {
		if i > 0 {
			if proxCmp(target, node.Addr(), last) < 0 {
				return false
			}
		}
		last = node.Addr()
	}
	return true
}

// push(node, max) adds the given node to the list, keeping the total size
// below max elements.
func (h *nodesByDistance) push(node Node, max int) {
	// returns the firt index ix such that func(i) returns true
	ix := sort.Search(len(h.nodes), func(i int) bool {
		return proxCmp(h.target, h.nodes[i].Addr(), node.Addr()) >= 0
	})

	if len(h.nodes) < max {
		h.nodes = append(h.nodes, node)
	}
	if ix < len(h.nodes) {
		copy(h.nodes[ix+1:], h.nodes[ix:])
		h.nodes[ix] = node
	}
}

// insert adds a peer to a bucket either by appending to existing items if
// bucket length does not exceed bucketLength, or by replacing the worst
// Node in the bucket
func (self *bucket) insert(node Node) (dropped Node, pos int) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if len(self.nodes) >= self.size { // >= allows us to add peers beyond the bucketsize limitation
		dropped, pos = self.worstNode()
		if dropped != nil {
			self.nodes[pos] = node
			dropped.Drop()
			return
		}
	}
	self.nodes = append(self.nodes, node)
	return
}

// worst expunges the single worst entry in a row, where worst entry is with a peer that has not been active the longests
func (self *bucket) worstNode() (node Node, pos int) {
	var oldest time.Time
	for p, n := range self.nodes {
		if (oldest == time.Time{}) || node.LastActive().Before(oldest) {
			oldest = n.LastActive()
			node = n
			pos = p
		}
	}
	return
}

/*
Taking the proximity value relative to a fix point x classifies the points in
the space (n byte long byte sequences) into bins the items in which are each at
most half as distant from x as items in the previous bin. Given a sample of
uniformly distrbuted items (a hash function over arbitrary sequence) the
proximity scale maps onto series of subsets with cardinalities on a negative
exponential scale.

It also has the property that any two item belonging to the same bin are at
most half as distant from each other as they are from x.

If we think of random sample of items in the bins as connections in a network of interconnected nodes than relative proximity can serve as the basis for local
decisions for graph traversal where the task is to find a route between two
points. Since in every step of forwarding, the finite distance halves, there is
a guaranteed constant maximum limit on the number of hops needed to reach one
node from the other.
*/

func (self *Kademlia) proximityBin(other Address) (ret int) {
	ret = proximity(self.addr, other)
	if ret > self.MaxProx {
		ret = self.MaxProx
	}
	return
}

/*
The distance metric MSB(x, y) of two equal length byte sequences x an y is the
value of the binary integer cast of the xor-ed byte sequence (most significant
bit first).
proximity(x, y) counts the common zeros in the front of this distance measure.
which is equivalent to the reverse rank of the integer part of the base 2
logarithm of the distance
called proximity belt (0 farthest, 255 closest, 256 self)
*/
func proximity(one, other Address) (ret int) {
	for i := 0; i < len(one); i++ {
		oxo := one[i] ^ other[i]
		for j := 0; j < 8; j++ {
			if (uint8(oxo)>>uint8(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}
	return len(one) * 8
}

// proxCmp compares the distances a->target and b->target.
// Returns -1 if a is closer to target, 1 if b is closer to target
// and 0 if they are equal.
func proxCmp(target, a, b Address) int {
	for i := range target {
		da := a[i] ^ target[i]
		db := b[i] ^ target[i]
		if da > db {
			return 1
		} else if da < db {
			return -1
		}
	}
	return 0
}

func (self *Kademlia) DB() [][]*NodeRecord {
	return self.nodeDB
}

// save persists all peers encountered
func (self *Kademlia) Save(path string) error {

	kad := kadDB{
		Address: self.addr,
		Nodes:   self.nodeDB,
	}

	for _, b := range kad.Nodes {
		for _, node := range b {
			node.setActive()
		}
	}

	data, err := json.MarshalIndent(&kad, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, os.ModePerm)
}

// loading the idle node record from disk
func (self *Kademlia) Load(path string) (err error) {
	var data []byte
	data, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}
	var kad kadDB
	err = json.Unmarshal(data, &kad)
	if err != nil {
		return
	}
	if self.addr != kad.Address {
		return fmt.Errorf("invalid kad db: address mismatch, expected %v, got %v", self.addr, kad.Address)
	}
	self.nodeDB = kad.Nodes
	return
}

// randomAddressAt(address, prox) generates a random address
// at proximity order prox relative to address
// if prox is negative a random address is generated
func RandomAddressAt(self Address, prox int) (addr Address) {
	addr = self
	var pos int
	if prox >= 0 {
		pos = prox / 8
		trans := prox % 8
		transbytea := byte(0)
		for j := 0; j <= trans; j++ {
			transbytea |= 1 << uint8(7-j)
		}
		flipbyte := byte(1 << uint8(7-trans))
		transbyteb := transbytea ^ byte(255)
		randbyte := byte(rand.Intn(255))
		addr[pos] = ((addr[pos] & transbytea) ^ flipbyte) | randbyte&transbyteb
	}
	for i := pos + 1; i < len(addr); i++ {
		addr[i] = byte(rand.Intn(255))
	}

	return
}

// randomAddressAt() generates a random address
func RandomAddress() Address {
	return RandomAddressAt(Address{}, -1)
}

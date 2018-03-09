// Copyright (c) 2018 Wolk Inc.  All rights reserved.

// The SWARMDB library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The SWARMDB library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
/*
 // Validator map reduce operations
 // SMASH
 cat data/farmerlog.txt data/buyerlog.txt | ./validator -mapred=smashmap | sort | ./validator  -mapred=smashred > data/storage-input.txt

 // STORAGE
 cat data/storage-input.txt | ./validator -mapred=storagemap | sort | ./validator -mapred=storagered > data/collation-storage.txt

 // SWAP
 cat data/swaplog.txt | ./validator -mapred=swapmap | sort | ./validator  -mapred=swapred > data/bandwidth-input.txt

 // BANDWIDTH
 cat data/bandwidth-input.txt | ./validator -mapred=bandwidthmap | sort | ./validator -mapred=bandwidthred > data/collation-bandwidth.txt

 // COLLATION
 cat data/collation-*.txt | go run validator.go -mapred=collationmap | sort | go run validator.go -mapred=collationred > data/submittx-input.txt
*/
package swarmdb

import (
	"bufio"
	"encoding/json"
	// "swarmdb"
	"github.com/ethereum/go-ethereum/crypto"

	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type SWARMDBNode struct {
	farmerID string
	ip       string
	port     int
}

type Validator struct {
	nodes         []SWARMDBNode
	BandwidthCost int
	StorageCost   int
}

// swapMapRed:      SwapLogEntry => BandwidthLogEntry
type SwapLogEntry struct {
	SwapID   string `json:"swapID"`
	LocalID  string `json:"localID"`
	RemoteID string `json:"remoteID"`
	B        int    `json:"b"`
	Sig      string `json:"sig"`
}

// bandwidthMapRed: BandwidthLogEntry => collationLogEntry
type BandwidthLogEntry struct {
	ID string `json:"ID"`
	B  int    `json:"b"`
}

// smashMapRed:     SmashLogEntry => storageLogEntry
// Buyer log entries are produced by swarmdb nodes and are storage requests
// {"buyer":"0xd80a4004350027f618107fe3240937d54e46c21b","chunkID":"1d4b6e4aa86d48c464c9adf83940d4e00df8affc","chunkBD":1515542149,"rep":3,"renewable":1,"sig":"s1","smash":"smash1"}

// Farmer log entries are produced by swarmdb nodes and are storage claims
// {"farmer":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","chunkID":"1d4b6e4aa86d48c464c9adf83940d4e00df8affc","chunkBD":1515542149,"rep":5,"renewable":1}
type SmashLogEntry struct {
	ChunkID   string `json:"chunkID"`
	Smash     string `json:"smash,omitempty"`
	FarmerID  string `json:"farmer,omitempty"`
	BuyerID   string `json:"buyer,omitempty"`
	ChunkBD   int    `json:"chunkBD,omitempty"`
	Rep       int    `json:"rep,omitempty"`
	Renewable int    `json:"renewable,omitempty"`
}

// storageLogEntries bring the above VALIDATED SmashLogEntry from multiple buyers and multiple farmers together
// {"id":"1d4b6e4aa86d48c464c9adf83940d4e00df8affc","buyers":["0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","0xd80a4004350027f618107fe3240937d54e46c21b"],"farmers":["0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","0xfd990c3c42446f6705dd66376bf5820cf2c09527"],"rep":3,"renewable":1,"chunkBD":1515542149,"smash":"smash1","sig":"s1"}
// storageMapRed:   storageLogEntry => collationLogEntry
type StorageLogEntry struct {
	ChunkID   string   `json:"chunkID"`
	Buyers    []string `json:"buyers,omitempty"`
	Farmers   []string `json:"farmers,omitempty"`
	Rep       int      `json:"rep,omitempty"`
	Renewable int      `json:"renewable,omitempty"`
	ChunkBD   int      `json:"chunkBD,omitempty"`
	Smash     string   `json:"smash,omitempty"`
}

// collationMapRed: CollationLogEntry => CollationSummaryEntry
type CollationLogEntry struct {
	ID string `json:"ID"`
	B  int    `json:"b,omitempty"`
	S  int    `json:"s,omitempty"`
	SB int    `json:"sb,omitempty"`
	SF int    `json:"sf,omitempty"`
}

type CollationSummaryEntry struct {
	ID string `json:"ID"`
	B  int    `json:"b,omitempty"`
	S  int    `json:"s,omitempty"`
	T  int    `json:"t,omitempty"`
}

/* LOG Downloading */
func (self *Validator) AddNode(farmerID string, ip string, port int) (err error) {
	var n SWARMDBNode
	n.farmerID = farmerID
	n.ip = ip
	n.port = port
	self.nodes = append(self.nodes, n)
	return nil
}

func (self *Validator) getSWARMDBLogs(logtype string, path string, epoch string) (err error) {
	// for each of the nodes, get the "smash" logs and put it in the path
	for _, n := range self.nodes {
		url := fmt.Sprintf("http://%s:%ip/%s/%s", n.ip, n.port, logtype, epoch)
		switch logtype {
		case "smash":
			url = "http://sourabh.wolk.com/validator/buyerlog-input.txt"
		case "storage":
			url = "http://sourabh.wolk.com/validator/farmerlog-input.txt"
		case "swap":
			url = "http://sourabh.wolk.com/validator/swap-input.txt"
		}

		resp, err := http.Get(url)
		if err != nil {
			// handle error
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
		}
		// save the smash log
		fn := fmt.Sprintf("%s/%s-%s.%s", path, epoch, n.farmerID, logtype)
		d1 := []byte(body)
		err = ioutil.WriteFile(fn, d1, 0644)
		fmt.Printf("SAVING %s (%d bytes)\n", fn, len(d1))
		if err != nil {
			//
		}

	}
	return nil
}

func (self *Validator) GetSWARMDBSmashLogs(path string, epoch string) (err error) {
	return self.getSWARMDBLogs("smash", path, epoch)
}

func (self *Validator) GetSWARMDBSwapLogs(path string, epoch string) (err error) {
	return self.getSWARMDBLogs("swap", path, epoch)
}

func (self *Validator) GetSWARMDBStorageLogs(path string, epoch string) (err error) {
	return self.getSWARMDBLogs("storage", path, epoch)
}

func fetchLogs() {
	var v Validator
	v.AddNode("farmer1", "0.0.0.0", 8501)

	path := "/tmp"
	epoch := "1"

	err := v.GetSWARMDBSmashLogs(path, epoch)
	if err != nil {
	}

	err = v.GetSWARMDBStorageLogs(path, epoch)
	if err != nil {
	}

	err = v.GetSWARMDBSwapLogs(path, epoch)
	if err != nil {
	}
}

/*
 SMASH map-reduce job takes storage requests and storage claims, validates them
Input:
{"buyer":"0xd80a4004350027f618107fe3240937d54e46c21b","chunkID":"1d4b6e4aa86d48c464c9adf83940d4e00df8affc","chunkBD":1515542149,"rep":3,"renewable":1,"sig":"s1","smash":"smash1"}
{"buyer":"0xd80a4004350027f618107fe3240937d54e46c21b","chunkID":"4b09668b93c718092a408c4222867968fcd3ad98","chunkBD":1515542150,"rep":4,"renewable":0,"sig":"s2","smash":"smash2"}
{"buyer":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","chunkID":"aeec6f5aca72f3a005af1b3420ab8c8c7009bac8","chunkBD":1515542151,"rep":5,"renewable":1,"sig":"s3","smash":"smash3"}
{"buyer":"0xd80a4004350027f618107fe3240937d54e46c21b","chunkID":"aeec6f5aca72f3a005af1b3420ab8c8c7009bac8","chunkBD":1515542152,"rep":4,"renewable":0,"sig":"s4","smash":"smash4"}
{"buyer":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","chunkID":"1d4b6e4aa86d48c464c9adf83940d4e00df8affc","chunkBD":1515542153,"rep":3,"renewable":1,"sig":"s5","smash":"smash5"}
{"buyer":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","chunkID":"d368b1c09e7ddfb6aff24e8e6f181ffeea905d31","chunkBD":1515542154,"rep":2,"renewable":0,"sig":"s6","smash":"smash6"}
{"farmer":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","chunkID":"1d4b6e4aa86d48c464c9adf83940d4e00df8affc","chunkBD":1515542149,"rep":5,"renewable":1}
{"farmer":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","chunkID":"4b09668b93c718092a408c4222867968fcd3ad98","chunkBD":1515542150,"rep":4,"renewable":0}
{"farmer":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","chunkID":"d368b1c09e7ddfb6aff24e8e6f181ffeea905d31","chunkBD":1515542151,"rep":3,"renewable":1}
{"farmer":"0xfd990c3c42446f6705dd66376bf5820cf2c09527","chunkID":"aeec6f5aca72f3a005af1b3420ab8c8c7009bac8","chunkBD":1515542152,"rep":2,"renewable":0}
{"farmer":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","chunkID":"aeec6f5aca72f3a005af1b3420ab8c8c7009bac8","chunkBD":1515542153,"rep":3,"renewable":1}
{"farmer":"0xfd990c3c42446f6705dd66376bf5820cf2c09527","chunkID":"1d4b6e4aa86d48c464c9adf83940d4e00df8affc","chunkBD":1515542154,"rep":4,"renewable":0}

Output:
{"id":"1d4b6e4aa86d48c464c9adf83940d4e00df8affc","buyers":["0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","0xd80a4004350027f618107fe3240937d54e46c21b"],"farmers":["0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","0xfd990c3c42446f6705dd66376bf5820cf2c09527"],"rep":3,"renewable":1,"chunkBD":1515542149,"smash":"smash1","sig":"s1"}
{"id":"4b09668b93c718092a408c4222867968fcd3ad98","buyers":["0xd80a4004350027f618107fe3240937d54e46c21b"],"farmers":["0xf6b55acbbc49f4524aa48d19281a9a77c54de10f"],"rep":4,"renewable":0,"chunkBD":1515542150,"smash":"smash2","sig":"s2"}
{"id":"aeec6f5aca72f3a005af1b3420ab8c8c7009bac8","buyers":["0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","0xd80a4004350027f618107fe3240937d54e46c21b"],"farmers":["0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","0xfd990c3c42446f6705dd66376bf5820cf2c09527"],"rep":4,"renewable":0,"chunkBD":1515542152,"smash":"smash4","sig":"s4"}
{"id":"d368b1c09e7ddfb6aff24e8e6f181ffeea905d31","buyers":["0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8"],"farmers":["0xf6b55acbbc49f4524aa48d19281a9a77c54de10f"],"rep":2,"renewable":0,"chunkBD":1515542154,"smash":"smash6","sig":"s6"}
*/

func (self *Validator) SmashMap() (err error) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		inp := scanner.Text()
		var e SmashLogEntry
		err := json.Unmarshal([]byte(inp), &e)
		if err != nil {
			fmt.Printf(" error %v\n", err)
		} else {
			if len(e.ChunkID) > 0 {
				fmt.Printf("%s\t%s\n", e.ChunkID, inp)

			}
		}
	}

	if scanner.Err() != nil {
		// handle error.
	}
	return nil
}

func (self *Validator) SmashRed() (err error) {
	scanner := bufio.NewScanner(os.Stdin)
	previd := ""

	var buyerLog []SmashLogEntry
	var farmerLog []SmashLogEntry
	for scanner.Scan() {
		line := scanner.Text()
		var sa []string
		sa = strings.Split(line, "\t")
		if len(sa) == 2 {
			chunkID := sa[0]
			if len(previd) > 0 && strings.Compare(previd, chunkID) != 0 {
				self.tallySmash(previd, buyerLog, farmerLog)
				buyerLog = nil
				farmerLog = nil
			}

			var e SmashLogEntry
			err := json.Unmarshal([]byte(sa[1]), &e)
			if err != nil {
				fmt.Printf(" error %v\n", err)
			} else {
				if len(e.FarmerID) > 0 {
					farmerLog = append(farmerLog, e)
				} else if len(e.BuyerID) > 0 {
					buyerLog = append(buyerLog, e)
				}
				previd = chunkID
			}
		}
	}
	if scanner.Err() != nil {
		// handle error.
	}
	return nil
}

func (self *Validator) GetFarmer(farmerID string) (f SWARMDBNode, err error) {
	// look up farmer
	f.ip = "0.0.0.0"
	f.port = 8501
	return f, nil
}

func (self *Validator) validateChunk(farmerID string, chunkID string, smash string) (ok bool, err error) {
	return true, nil
	// TODO: for review -- do a tcp open and relay a bunch of ash proof requests, but for now just do it via http

	// fetch the chunk from the node (within a specific timeout limit, up to N times)
	farmer, err := self.GetFarmer(farmerID)
	if err != nil {
		return false, err
	}
	url := fmt.Sprintf("http://%s:%s/ash/%s", farmer.ip, farmer.port, chunkID)

	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	if len(body) > 0 {
		// use respnse smash actual
		return true, nil
	}
	return false, nil
}

func (self *Validator) tallySmash(chunkID string, buyerLog []SmashLogEntry, farmerLog []SmashLogEntry) (err error) {
	var out StorageLogEntry

	if len(buyerLog) > 0 && len(farmerLog) > 0 {
		for _, b := range buyerLog {
			out.ChunkID = chunkID
			if len(out.Buyers) > 0 {
				out.Buyers = append(out.Buyers, b.BuyerID)
			} else {
				// first buyers info with valid signature gets control
				out.Buyers = append(out.Buyers, b.BuyerID)
				out.Rep = b.Rep
				out.Renewable = b.Renewable
				out.ChunkBD = b.ChunkBD
				out.Smash = b.Smash
			}
		}
		for _, f := range farmerLog {
			ok, err := self.validateChunk(f.FarmerID, f.ChunkID, out.Smash)
			if err != nil {
				// TODO: do something to skip this chunk
			}
			if ok {
				// add this farmer ONLY because farmer returned a valid SMASH proof
				out.Farmers = append(out.Farmers, f.FarmerID)
			} else {
				// TODO: farmer provided invalid proof
			}
		}
		o, err := json.Marshal(out)
		if err != nil {
			// fmt.Printf(" error %v\n", err)
		}
		fmt.Printf("%s\n", string(o))

	} else if len(buyerLog) > 0 {
		// buyer requested chunk storage but no farmers made claims
	} else if len(farmerLog) > 0 {
		// farmers made chunk claims but no buyer made requests for storage
	}
	return nil
}

/*
SWAP Map-reduce

Input:
{"id":"0xd80a4004350027f618107fe3240937d54e46c21b","remote":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","b":1234,"receipt":"r1"}
{"id":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","remote":"0xd80a4004350027f618107fe3240937d54e46c21b","b":-1234,"receipt":"r2"}
{"id":"0xd80a4004350027f618107fe3240937d54e46c21b","remote":"0xaeec6f5aca72f3a005af1b3420ab8c8c7009bac8","b":56,"receipt":"r3"}
{"id":"0xaeec6f5aca72f3a005af1b3420ab8c8c7009bac8","remote":"0xd80a4004350027f618107fe3240937d54e46c21b","b":-56,"receipt":"r4"}

Output:
{"id":"0xaeec6f5aca72f3a005af1b3420ab8c8c7009bac8","remote":"0xd80a4004350027f618107fe3240937d54e46c21b","b":-56}
{"id":"0xd80a4004350027f618107fe3240937d54e46c21b","remote":"0xaeec6f5aca72f3a005af1b3420ab8c8c7009bac8","b":56}
{"id":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","remote":"0xd80a4004350027f618107fe3240937d54e46c21b","b":-1234}
{"id":"0xd80a4004350027f618107fe3240937d54e46c21b","remote":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","b":1234}
*/

func (self *Validator) SwapMap() (err error) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		inp := scanner.Text()
		var e SwapLogEntry
		err := json.Unmarshal([]byte(inp), &e)
		if err != nil {
			fmt.Printf(" error %v\n", err)
		} else {
		}
		if len(e.LocalID) > 0 && len(e.RemoteID) > 0 {
			if e.B < 0 { // the person receiving the check is the author of this line
				if self.validateSwapSignature(e.SwapID, e.Sig, e.LocalID) {
					// TODO: check for valid signature here - sig should be the localID
					fmt.Printf("%s\t%s\t%s\t%d\n", e.SwapID, e.LocalID, e.RemoteID, e.B)
				} else {
					// TODO: provide feedback to localID
				}
			} else {
				if self.validateSwapSignature(e.SwapID, e.Sig, e.RemoteID) {
					// TODO: check for valid signature here - sig should be the remoteID
					fmt.Printf("%s\t%s\t%s\t%d\n", e.SwapID, e.LocalID, e.RemoteID, e.B)
				} else {
					// TODO: provide feedback to remoteID
				}
			}
		}
	}

	if scanner.Err() != nil {
		// handle error.
	}
	return nil
}

func (self *Validator) validateSwapSignature(swapID string, sig string, signer string) (ok bool) {
	// TODO: validate this
	return true
}

func (self *Validator) SwapRed() (err error) {
	scanner := bufio.NewScanner(os.Stdin)
	previd := ""
	beneficiary1 := ""
	sender1 := ""
	b1 := 0
	beneficiary2 := ""
	sender2 := ""
	b2 := 0
	for scanner.Scan() {
		line := scanner.Text()
		var sa []string
		sa = strings.Split(line, "\t")
		if len(sa) == 4 {
			swapID := sa[0]
			if len(previd) > 0 && (strings.Compare(previd, swapID) != 0) {
				self.tallySwap(previd, beneficiary1, sender1, b1, beneficiary2, sender2, b2)
				beneficiary1 = ""
				sender1 = ""
				beneficiary2 = ""
				sender2 = ""
				b1 = 0
				b2 = 0
			}
			bv := sa[3]
			b, err := strconv.Atoi(bv)
			if err != nil {
			} else {
				if b < 0 {
					b2 = b
					beneficiary2 = sa[1]
					sender2 = sa[2]
				} else {
					b1 = b
					beneficiary1 = sa[2]
					sender1 = sa[1]
				}
			}
			previd = sa[0]
		}
	}
	if len(previd) > 0 {
		self.tallySwap(previd, beneficiary1, sender1, b1, beneficiary2, sender2, b2)
	}
	return nil
}

func (self *Validator) tallySwap(swapID string, beneficiary1 string, sender1 string, b1 int, beneficiary2 string, sender2 string, b2 int) {
	var e BandwidthLogEntry
	if b2 < 0 && b1 > 0 && (b1+b2 < 1) {
		// e.SwapID = swapID
		e.ID = beneficiary1
		e.B = b1
		out, err := json.Marshal(e)
		if err != nil {
			// fmt.Printf(" error %v\n", err)
		}
		fmt.Printf("%s\n", string(out))

		e.ID = sender2
		e.B = b2
		out, err = json.Marshal(e)
		if err != nil {
			// fmt.Printf(" error %v\n", err)
		}
		fmt.Printf("%s\n", string(out))
	}
}

/*
STORAGE Map Reduce Job
Input: (this is the output of the SWAP map reduce job)
{"id":"1d4b6e4aa86d48c464c9adf83940d4e00df8affc","buyers":["0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","0xd80a4004350027f618107fe3240937d54e46c21b"],"farmers":["0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","0xfd990c3c42446f6705dd66376bf5820cf2c09527"],"rep":3,"renewable":1,"chunkBD":1515542149,"smash":"smash1","sig":"s1"}
{"id":"4b09668b93c718092a408c4222867968fcd3ad98","buyers":["0xd80a4004350027f618107fe3240937d54e46c21b"],"farmers":["0xf6b55acbbc49f4524aa48d19281a9a77c54de10f"],"rep":4,"renewable":0,"chunkBD":1515542150,"smash":"smash2","sig":"s2"}
{"id":"aeec6f5aca72f3a005af1b3420ab8c8c7009bac8","buyers":["0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","0xd80a4004350027f618107fe3240937d54e46c21b"],"farmers":["0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","0xfd990c3c42446f6705dd66376bf5820cf2c09527"],"rep":4,"renewable":0,"chunkBD":1515542152,"smash":"smash4","sig":"s4"}
{"id":"d368b1c09e7ddfb6aff24e8e6f181ffeea905d31","buyers":["0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8"],"farmers":["0xf6b55acbbc49f4524aa48d19281a9a77c54de10f"],"rep":2,"renewable":0,"chunkBD":1515542154,"smash":"smash6","sig":"s6"}

Output: (this is the input of the collation map reduce job)
{"id":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","sb":7,"sf":0,"s":-7}
{"id":"0xd80a4004350027f618107fe3240937d54e46c21b","sb":1,"sf":0,"s":-1}
{"id":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","sb":0,"sf":5,"s":5}
{"id":"0xfd990c3c42446f6705dd66376bf5820cf2c09527","sb":0,"sf":3,"s":3}
*/
func (self *Validator) StorageMap() (err error) {
	thresh := 100
	scanner := bufio.NewScanner(os.Stdin)
	var tallyBuyer map[string]int
	var tallyFarmer map[string]int

	tallyBuyer = make(map[string]int)
	tallyFarmer = make(map[string]int)
	c := 0
	for scanner.Scan() {
		inp := scanner.Text()
		var e StorageLogEntry
		err := json.Unmarshal([]byte(inp), &e)
		if err != nil {
			fmt.Printf(" error %v\n", err)
		}
		buyer, farmers := e.selectedBuyerAndFarmers()
		if len(farmers) < 3 {
			// TODO: let the buyer know we have NO farmers claiming chunks -- MAJOR PROBLEM, NEED IMMEDIATE action
		} else if len(farmers) < 3 {
			// TODO: let the buyer know we have TOO FEW farmers claiming chunks
		}
		if len(buyer) > 0 {
			tallyBuyer[buyer]++
		}
		for _, farmerID := range farmers {
			tallyFarmer[farmerID]++
		}
		if c == thresh {
			fmt.Printf("%v | %v\n", tallyBuyer, tallyFarmer)
			self.tallyStorage(tallyBuyer, tallyFarmer)
			tallyBuyer = make(map[string]int)
			tallyFarmer = make(map[string]int)
		} else {
			c++
		}
	}
	self.tallyStorage(tallyBuyer, tallyFarmer)

	return nil
}

func (self *Validator) tallyStorage(tallyBuyer map[string]int, tallyFarmer map[string]int) {
	for buyer, sb := range tallyBuyer {
		fmt.Printf("%s\t%d\t0\n", buyer, sb)
	}
	for farmer, sf := range tallyFarmer {
		fmt.Printf("%s\t0\t%d\n", farmer, sf)
	}
}

func (self *StorageLogEntry) selectedBuyerAndFarmers() (buyer string, farmers []string) {
	if len(self.Buyers) > 0 {
		buyer = self.Buyers[0]
	} else {
		buyer = ""
	}
	if len(self.Farmers) > 0 {
		for i := 0; i < self.Rep && i < len(self.Farmers); i++ {
			farmers = append(farmers, self.Farmers[i])
		}
	}
	return buyer, farmers
}

func (self *Validator) StorageRed() (err error) {
	scanner := bufio.NewScanner(os.Stdin)
	previd := ""
	tsb := 0
	tsf := 0
	for scanner.Scan() {
		line := scanner.Text()
		var sa []string
		sa = strings.Split(line, "\t")
		if len(sa) == 3 {
			id := sa[0]
			sb, _ := strconv.Atoi(sa[1])
			sf, _ := strconv.Atoi(sa[2])
			if strings.Compare(id, previd) != 0 && len(previd) > 0 {
				self.tallyStorageSummary(previd, tsb, tsf)
				tsb = 0
				tsf = 0
			}
			tsb += sb
			tsf += sf
			previd = id
		}
	}
	if tsb > 0 || tsf > 0 && len(previd) > 0 {
		self.tallyStorageSummary(previd, tsb, tsf)
	}

	if scanner.Err() != nil {
		// handle error.
	}
	return nil
}

func (self *Validator) tallyStorageSummary(id string, tsb int, tsf int) {
	var o CollationLogEntry
	o.ID = id
	o.SB = tsb
	o.SF = tsf
	o.S = o.SF - o.SB
	out, err := json.Marshal(o)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", out)
}

/*
bandwidth  Map reduce
Input: (this is the output of the "swap" job)
{"id":"0xaeec6f5aca72f3a005af1b3420ab8c8c7009bac8","remote":"0xd80a4004350027f618107fe3240937d54e46c21b","b":-56}
{"id":"0xd80a4004350027f618107fe3240937d54e46c21b","remote":"0xaeec6f5aca72f3a005af1b3420ab8c8c7009bac8","b":56}
{"id":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","remote":"0xd80a4004350027f618107fe3240937d54e46c21b","b":-1234}
{"id":"0xd80a4004350027f618107fe3240937d54e46c21b","remote":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","b":1234}

 Output: (used as input to "collation" map-reduce)
{"id":"0xaeec6f5aca72f3a005af1b3420ab8c8c7009bac8","b":-56}
{"id":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","b":-1234}
{"id":"0xd80a4004350027f618107fe3240937d54e46c21b","b":1290}
*/
func (self *Validator) BandwidthMap() (err error) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		inp := scanner.Text()
		var e BandwidthLogEntry
		err := json.Unmarshal([]byte(inp), &e)
		if err != nil {
			fmt.Printf(" error %v\n", err)
		}
		fmt.Printf("%s\t%d\n", e.ID, e.B)
	}
	return nil
}

func (self *Validator) BandwidthRed() (err error) {
	previd := ""
	tb := 0
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		var sa []string
		sa = strings.Split(line, "\t")
		if len(sa) == 2 {
			id := sa[0]
			b, err := strconv.Atoi(sa[1])
			if err != nil {
			} else {

			}
			if strings.Compare(previd, id) != 0 && len(previd) > 0 {
				self.tallyBandwidthSummary(previd, tb)
				tb = 0
			}
			tb += b
			previd = id
		}
	}
	if tb > 0 {
		self.tallyBandwidthSummary(previd, tb)
	}
	return nil
}

func (self *Validator) tallyBandwidthSummary(id string, b int) {
	var o CollationLogEntry
	o.ID = id
	o.B = b
	out, err := json.Marshal(o)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", out)
}

/*
COLLATION map-reduce

INPUT:  (from the output of "storage" and "bandwidth" map-reduce tasks
{"id":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","sb":7,"sf":0,"s":-7}
{"id":"0xd80a4004350027f618107fe3240937d54e46c21b","sb":1,"sf":0,"s":-1}
{"id":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","sb":0,"sf":5,"s":5}
{"id":"0xfd990c3c42446f6705dd66376bf5820cf2c09527","sb":0,"sf":3,"s":3}
{"id":"0xaeec6f5aca72f3a005af1b3420ab8c8c7009bac8","b":-56}
{"id":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","b":-1234}
{"id":"0xd80a4004350027f618107fe3240937d54e46c21b","b":1290}

OUTPUT:
{"id":"0xaeec6f5aca72f3a005af1b3420ab8c8c7009bac8","b":0,"s":-56,"t":-175.92918860104}
{"id":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","b":-7,"s":-1234,"t":-3895.7533073293}
{"id":"0xd80a4004350027f618107fe3240937d54e46c21b","b":-1,"s":1290,"t":4049.9362413026}
{"id":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","b":5,"s":0,"t":13.5914091423}
{"id":"0xfd990c3c42446f6705dd66376bf5820cf2c09527","b":3,"s":0,"t":8.15484548538}
*/

func (self *Validator) CollationMap() (err error) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		inp := scanner.Text()
		var e CollationLogEntry
		err := json.Unmarshal([]byte(inp), &e)
		if err != nil {
			fmt.Printf(" error %v\n", err)
		}
		fmt.Printf("%s\t%d\t%d\n", e.ID, e.B, e.S)
	}
	return nil
}

func (self *Validator) CollationRed() (err error) {

	previd := ""
	tb := 0
	ts := 0
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		var sa []string
		sa = strings.Split(line, "\t")
		if len(sa) == 3 {
			id := sa[0]
			b, _ := strconv.Atoi(sa[1])
			s, _ := strconv.Atoi(sa[2])

			if strings.Compare(previd, id) != 0 && len(previd) > 0 {
				self.tallyCollationSummary(previd, tb, ts)
				tb = 0
				ts = 0
			}
			tb += b
			ts += s
			previd = id
		}
	}
	if tb > 0 || ts > 0 {
		self.tallyCollationSummary(previd, tb, ts)
	}
	return nil
}

func (self *Validator) tallyCollationSummary(id string, b int, s int) {
	var e CollationSummaryEntry
	e.ID = id
	e.B = b
	e.S = s
	e.T = b*self.BandwidthCost + s*self.StorageCost
	out, err := json.Marshal(e)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", string(out))
}

func genID(i int, maxlen int) (out string) {
	o := fmt.Sprintf("%d", i)
	hash := crypto.Keccak256([]byte(o))
	out = fmt.Sprintf("%x", hash)
	if len(out) > maxlen {
		out1 := []rune(out)
		return (string(out1[0:maxlen]))
	}
	return out
}

func (self *Validator) GenerateLogs() (err error) {
	f0, err := os.Create("data/farmerlog.txt")
	if err != nil {
		return fmt.Errorf("Could not create buyerlog")
	}
	defer f0.Close()
	f1, err := os.Create("data/buyerlog.txt")
	if err != nil {
		return fmt.Errorf("Could not create farmerlog")
	}
	defer f1.Close()
	for i := 1; i < 100; i++ {
		// farmerlog
		var sm0 SmashLogEntry
		sm0.FarmerID = genID(3+i%8, 40)
		sm0.ChunkID = genID(100+i, 64)
		out0, _ := json.Marshal(sm0)
		f0.WriteString(fmt.Sprintf("%s\n", string(out0)))

		// buyerlog
		var sm1 SmashLogEntry
		sm1.ChunkID = genID(100+i, 64)
		sm1.ChunkBD = 1518508203
		sm1.Rep = 3
		sm1.Renewable = i % 2
		sm1.BuyerID = genID(i%3, 40)
		sm1.Smash = genID(1000+i, 64)
		out1, _ := json.Marshal(sm1)
		f1.WriteString(fmt.Sprintf("%s\n", string(out1)))
	}
	f3, err := os.Create("data/swaplog.txt")
	if err != nil {
		return fmt.Errorf("Could not create swaplog")
	}
	defer f3.Close()
	for i := 1; i < 100; i++ {
		// local is getting +B
		var sw0 SwapLogEntry
		sw0.SwapID = genID(i+1000, 64)
		sw0.LocalID = genID(3+i%8, 40)
		sw0.RemoteID = genID(i%3, 40)
		sw0.B = i%3 + 3
		sw0.Sig = "sig0" // km.SignMessage

		out, _ := json.Marshal(sw0)
		f3.WriteString(fmt.Sprintf("%s\n", string(out)))

		// local is getting -B
		var sw1 SwapLogEntry
		sw1.SwapID = genID(i+1000, 64)
		sw1.LocalID = sw0.RemoteID
		sw1.RemoteID = sw0.LocalID
		sw1.B = -sw0.B
		sw1.Sig = "sig1" // km.SignMessage
		out, _ = json.Marshal(sw1)
		f3.WriteString(fmt.Sprintf("%s\n", string(out)))
	}
	return nil
}

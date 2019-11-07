//func (k *Kademlia) MarshalJSON() ([]byte, error) {
//var (
//kademliaInfo = make(map[string]string)
//depth        = depthForPot(k.conns, k.NeighbourhoodSize, k.base)
//numConns     = k.conns.Size()
//numAddrs     = k.addrs.Size()

//connBinMap = make(map[string]string)
//)

//k.conns.EachBin(k.base, Pof, 0, func(bin *pot.Bin) bool {
//binMap := make(map[string]string)

//po := bin.ProximityOrder
//if po >= k.MaxProxDisplay {
//po = k.MaxProxDisplay - 1 // why???
//}
//size := bin.Size
//poStr := fmt.Sprintf("%d", po)
//binMap["po"] = poStr
//binMap["numConns"] = fmt.Sprintf("%d", size)

//var addresses []string
//bin.ValIterator(func(val pot.Val) bool {
//e := val.(*Peer)
//addresses = append(addresses, hex.EncodeToString(e.Address()[:16]))
//return true
//})

//binMap["addresses"] = strings.Join(addresses, ",")

//binString, _ := json.Marshal(binMap)
//connBinMap[poStr] = binString
//return true
//})

//k.addrs.EachBin(k.base, Pof, 0, func(bin *pot.Bin) bool {
//var rowlen int
//po := bin.ProximityOrder
//if po >= k.MaxProxDisplay {
//po = k.MaxProxDisplay - 1
//}
//size := bin.Size
//if size < 0 {
//panic("bin size shouldn't be less than zero")
//}
//row := []string{fmt.Sprintf("%2d", size)}
//// we are displaying live peers too
//bin.ValIterator(func(val pot.Val) bool {
//e := val.(*entry)
//row = append(row, Label(e))
//rowlen++
//return rowlen < 4
//})
//peersrows[po] = strings.Join(row, " ")
//return true
//})

//for i := 0; i < k.MaxProxDisplay; i++ {
//if i == depth {
//rows = append(rows, fmt.Sprintf("============ DEPTH: %d ==========================================", i))
//}
//left := liverows[i]
//right := peersrows[i]
//if len(left) == 0 {
//left = " 0                             "
//}
//if len(right) == 0 {
//right = " 0"
//}
//rows = append(rows, fmt.Sprintf("%03d %v | %v", i, left, right))
//}

//}



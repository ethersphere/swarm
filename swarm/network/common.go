package network

import (
	"fmt"
	"strings"
)

func LogAddrs(nns [][]byte) string {
	var nnsa []string
	for _, nn := range nns {
		nnsa = append(nnsa, fmt.Sprintf("%08x", nn[:4]))
	}
	return strings.Join(nnsa, ", ")
}

func logEmptyBins(ebs []int) string {
	var ebss []string
	for _, eb := range ebs {
		ebss = append(ebss, fmt.Sprintf("%d", eb))
	}
	return strings.Join(ebss, ", ")
}

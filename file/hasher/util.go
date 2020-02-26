package hasher

import (
	"math"
)

// TODO: level 0 should be SectionSize() not Branches()
// generates a dictionary of maximum span lengths per level represented by one SectionSize() of data
func generateSpanSizes(branches int, levels int) []int {
	spans := make([]int, levels)
	span := 1
	for i := 0; i < 9; i++ {
		spans[i] = span
		span *= branches
	}
	return spans
}

// TODO: use params instead of sectionSize, branches
// calculate the last level index which a particular data section count will result in.
// the returned level will be the level of the root hash
func getLevelsFromLength(l int, sectionSize int, branches int) int {
	if l == 0 {
		return 0
	} else if l <= sectionSize*branches {
		return 1
	}
	c := (l - 1) / (sectionSize)

	return int(math.Log(float64(c))/math.Log(float64(branches)) + 1)
}

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

// calculates the section index of the given byte size
func dataSizeToSectionIndex(length int, sectionSize int) int {
	return (length - 1) / sectionSize
}

// calculates the section count of the given byte size
func dataSizeToSectionCount(length int, sectionSize int) int {
	return dataSizeToSectionIndex(length, sectionSize) + 1
}

// calculates the corresponding level section for a data section
func dataSectionToLevelSection(p *treeParams, lvl int, sections int) int {
	span := p.Spans[lvl]
	return sections / span
}

// calculates the lower data section boundary of a level for which a data section is contained
// the higher level use is to determine whether the final data section written falls within
// a certain level's span
func dataSectionToLevelBoundary(p *treeParams, lvl int, section int) int {
	span := p.Spans[lvl+1]
	spans := section / span
	spanBytes := spans * span
	//log.Trace("levelboundary", "spans", spans, "section", section, "span", span)
	return spanBytes
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

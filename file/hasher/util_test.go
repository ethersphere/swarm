package hasher

import "testing"

// TestLevelsFromLength verifies getLevelsFromLength
func TestLevelsFromLength(t *testing.T) {

	sizes := []int{sectionSize, chunkSize, chunkSize + sectionSize, chunkSize * branches, chunkSize*branches + 1}
	expects := []int{1, 1, 2, 2, 3}

	for i, size := range sizes {
		lvl := getLevelsFromLength(size, sectionSize, branches)
		if expects[i] != lvl {
			t.Fatalf("size %d, expected %d, got %d", size, expects[i], lvl)
		}
	}
}

// TestDataSizeToSection verifies testDataSizeToSectionIndex
func TestDataSizeToSectionIndex(t *testing.T) {

	sizes := []int{chunkSize - 1, chunkSize, chunkSize + 1}
	expects := []int{branches - 1, branches - 1, branches}

	for j, size := range sizes {
		r := dataSizeToSectionIndex(size, sectionSize)
		expect := expects[j]
		if expect != r {
			t.Fatalf("size %d section %d: expected %d, got %d", size, sectionSize, expect, r)
		}
	}

}

// TestsDataSectionToLevelSection verifies dataSectionToLevelSection
func TestDataSectionToLevelSection(t *testing.T) {

	params := newTreeParams(dummyHashFunc)
	sections := []int{0, branches - 1, branches, branches + 1, branches * 2, branches*2 + 1, branches * branches}
	levels := []int{1, 2}
	expects := []int{
		0, 0, 1, 1, 2, 2, 128,
		0, 0, 0, 0, 0, 0, 1,
	}

	for i, lvl := range levels {
		for j, section := range sections {
			r := dataSectionToLevelSection(params, lvl, section)
			k := i*len(sections) + j
			expect := expects[k]
			if expect != r {
				t.Fatalf("levelsection size %d level %d: expected %d, got %d", section, lvl, expect, r)
			}
		}
	}
}

// TestDataSectionToLevelBoundary verifies dataSectionToLevelBoundary
func TestDataSectionToLevelBoundary(t *testing.T) {
	params := newTreeParams(dummyHashFunc)
	size := chunkSize*branches + chunkSize*2
	section := dataSizeToSectionIndex(size, sectionSize)
	lvl := 1
	expect := branches * branches

	r := dataSectionToLevelBoundary(params, lvl, section)
	if expect != r {
		t.Fatalf("levelboundary size %d level %d: expected %d, got %d", section, lvl, expect, r)
	}

	size = chunkSize*branches*branches + chunkSize*2
	section = dataSizeToSectionIndex(size, sectionSize)
	lvl = 1
	expect = branches * branches * branches

	r = dataSectionToLevelBoundary(params, lvl, section)
	if expect != r {
		t.Fatalf("levelboundary size %d level %d: expected %d, got %d", section, lvl, expect, r)
	}

	size = chunkSize*branches + chunkSize*2
	section = dataSizeToSectionIndex(size, sectionSize)
	lvl = 2
	expect = 0

	r = dataSectionToLevelBoundary(params, lvl, section)
	if expect != r {
		t.Fatalf("levelboundary size %d level %d: expected %d, got %d", section, lvl, expect, r)
	}
}

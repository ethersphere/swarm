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

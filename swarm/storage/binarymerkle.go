package storage

// This implementation does not take advantage of any paralellisms and uses
// far more memory than necessary, but it is easy to see that it is correct.
// It can be used for generating test cases for optimized implementations.

func BinaryMerkle(chunk []byte, hasher Hasher) []byte {
	hash := hasher()
	section := 2 * hash.Size()
	l := len(chunk)
	if l > section {
		n := l / section
		r := l - n*section
		hash.Write(chunk[0:r])
		next := hash.Sum(nil)
		for r < l {
			hash.Reset()
			hash.Write(chunk[r : r+section])
			next = hash.Sum(next)
			r += section
		}
		return BinaryMerkle(next, hasher)
	} else {
		hash.Write(chunk)
		return hash.Sum(nil)
	}
}

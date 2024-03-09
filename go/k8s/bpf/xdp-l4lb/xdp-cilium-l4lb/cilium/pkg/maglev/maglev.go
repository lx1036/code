package maglev

import (
	"runtime"
	"sync"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/murmur3"
)

const (
	DefaultTableSize = 16381

	// seed=$(head -c12 /dev/urandom | base64 -w0)
	DefaultHashSeed = "JLfvgnHc2kaSUFaI"
)

var (
	seedMurmur uint32

	SeedJhash0 uint32
	SeedJhash1 uint32

	// permutation is the slice containing the Maglev permutation calculations.
	permutation []uint64
)

// GetLookupTable returns the Maglev lookup table of the size "m" for the given
// backends. The lookup table contains the indices of the given backends.
func GetLookupTable(backends []string, m uint64) []int {
	if len(backends) == 0 {
		return nil
	}

	perm := getPermutation(backends, m, runtime.NumCPU())
	next := make([]int, len(backends))
	entry := make([]int, m)

	for j := uint64(0); j < m; j++ {
		entry[j] = -1
	}

	l := len(backends)

	for n := uint64(0); n < m; n++ {
		i := int(n) % l
		c := perm[i*int(m)+next[i]]
		for entry[c] >= 0 {
			next[i] += 1
			c = perm[i*int(m)+next[i]]
		}
		entry[c] = i
		next[i] += 1
	}

	return entry
}

func getPermutation(backends []string, m uint64, numCPU int) []uint64 {
	var wg sync.WaitGroup

	// The idea is to split the calculation into batches so that they can be
	// concurrently executed. We limit the number of concurrent goroutines to
	// the number of available CPU cores. This is because the calculation does
	// not block and is completely CPU-bound. Therefore, adding more goroutines
	// would result into an overhead (allocation of stackframes, stress on
	// scheduling, etc) instead of a performance gain.

	bCount := len(backends)
	if size := uint64(bCount) * m; size > uint64(len(permutation)) {
		// Reallocate slice so we don't have to allocate again on the next
		// call.
		permutation = make([]uint64, size)
	}

	batchSize := bCount / numCPU
	if batchSize == 0 {
		batchSize = bCount
	}

	for g := 0; g < bCount; g += batchSize {
		wg.Add(1)
		go func(from int) {
			to := from + batchSize
			if to > bCount {
				to = bCount
			}
			for i := from; i < to; i++ {
				offset, skip := getOffsetAndSkip(backends[i], m)
				permutation[i*int(m)] = offset % m
				for j := uint64(1); j < m; j++ {
					permutation[i*int(m)+int(j)] = (permutation[i*int(m)+int(j-1)] + skip) % m
				}
			}
			wg.Done()
		}(g)
	}
	wg.Wait()

	return permutation[:bCount*int(m)]
}

func getOffsetAndSkip(backend string, m uint64) (uint64, uint64) {
	h1, h2 := murmur3.Hash128([]byte(backend), seedMurmur)
	offset := h1 % m
	skip := (h2 % (m - 1)) + 1

	return offset, skip
}

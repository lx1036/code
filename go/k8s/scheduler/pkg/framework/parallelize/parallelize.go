package parallelize

import (
	"context"
	"math"

	"k8s.io/client-go/util/workqueue"
)

// DefaultParallelism is the default parallelism used in scheduler.
const DefaultParallelism int = 16

// Parallelizer holds the parallelism for scheduler.
type Parallelizer struct {
	parallelism int
}

// NewParallelizer returns an object holding the parallelism.
func NewParallelizer(p int) Parallelizer {
	return Parallelizer{parallelism: p}
}

func (p Parallelizer) Until(ctx context.Context, pieces int, doWorkPiece workqueue.DoWorkPieceFunc) {
	workqueue.ParallelizeUntil(ctx, p.parallelism, pieces, doWorkPiece, workqueue.WithChunkSize(chunkSizeFor(pieces, p.parallelism)))
}

// chunkSizeFor returns a chunk size for the given number of items to use for
// parallel work. The size aims to produce good CPU utilization.
// returns max(1, min(sqrt(n), n/Parallelism))
func chunkSizeFor(n, parallelism int) int {
	s := int(math.Sqrt(float64(n)))

	if r := n/parallelism + 1; s > r {
		s = r
	} else if s < 1 {
		s = 1
	}
	return s
}

package mvcc

import "testing"

func TestTreeIndexGet(test *testing.T) {
	treeIdx := newTreeIndex()
	treeIdx.Put([]byte("foo"), revision{main: 2})
	treeIdx.Put([]byte("foo"), revision{main: 4})
	treeIdx.Tombstone([]byte("foo"), revision{main: 6})

	fixtures := []struct {
		rev      int64
		wrev     revision
		wcreated revision
		wver     int64
		werr     error
	}{
		{0, revision{}, revision{}, 0, ErrRevisionNotFound},
		{1, revision{}, revision{}, 0, ErrRevisionNotFound},
		{2, revision{main: 2}, revision{main: 2}, 1, nil},
		{3, revision{main: 2}, revision{main: 2}, 1, nil},
		{4, revision{main: 4}, revision{main: 2}, 2, nil},
		{5, revision{main: 4}, revision{main: 2}, 2, nil},
		{6, revision{}, revision{}, 0, ErrRevisionNotFound},
	}

	for _, fixture := range fixtures {
		treeIdx.Get([]byte("foo"), fixture.rev)
	}
}

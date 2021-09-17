package mvcc

import (
	"reflect"
	"testing"
)

func TestTreeIndexGet(test *testing.T) {
	treeIdx := newTreeIndex()
	treeIdx.Put([]byte("foo"), revision{main: 2})
	treeIdx.Put([]byte("foo"), revision{main: 4})
	treeIdx.Tombstone([]byte("foo"), revision{main: 6})

	fixtures := []struct {
		desc      string
		rev       int64
		wantedRev revision
		created   revision
		wver      int64
		werr      error
	}{
		{"revision=0", 0, revision{}, revision{}, 0, ErrRevisionNotFound},
		{"revision=1", 1, revision{}, revision{}, 0, ErrRevisionNotFound},
		{"revision=2", 2, revision{main: 2}, revision{main: 2}, 1, nil},
		{"revision=3", 3, revision{main: 2}, revision{main: 2}, 1, nil},
		{"revision=4", 4, revision{main: 4}, revision{main: 2}, 2, nil},
		{"revision=5", 5, revision{main: 4}, revision{main: 2}, 2, nil},
		{"revision=6", 6, revision{}, revision{}, 0, ErrRevisionNotFound},
	}

	for i, fixture := range fixtures {
		test.Run(fixture.desc, func(t *testing.T) {
			rev, created, ver, err := treeIdx.Get([]byte("foo"), fixture.rev)
			if err != fixture.werr {
				test.Errorf("#%d: err = %v, want %v", i, err, fixture.werr)
			}
			if rev != fixture.wantedRev {
				test.Errorf("#%d: rev = %+v, want %+v", i, rev, fixture.wantedRev)
			}
			if created != fixture.created {
				test.Errorf("#%d: created = %+v, want %+v", i, created, fixture.created)
			}
			if ver != fixture.wver {
				test.Errorf("#%d: ver = %d, want %d", i, ver, fixture.wver)
			}
		})
	}
}

func TestTreeIndexRange(test *testing.T) {
	treeIdx := newTreeIndex()
	allKeys := [][]byte{[]byte("foo"), []byte("foo1"), []byte("foo2")}
	allRevs := []revision{{main: 1}, {main: 2}, {main: 3}}
	for i := range allKeys {
		treeIdx.Put(allKeys[i], allRevs[i])
	}

	fixtures := []struct {
		desc     string
		key, end []byte
		wkeys    [][]byte
		wrevs    []revision
	}{
		{
			"single key that not found", []byte("bar"), nil, nil, nil,
		},
		{
			"single key that found", []byte("foo"), nil, allKeys[:1], allRevs[:1],
		},
		{
			"range keys, return first member", []byte("foo"), []byte("foo1"), allKeys[:1], allRevs[:1],
		},
		{
			"range keys, return first two members", []byte("foo"), []byte("foo2"), allKeys[:2], allRevs[:2],
		},
		{
			"range keys, return all members", []byte("foo"), []byte("fop"), allKeys, allRevs,
		},
		{
			"range keys, return last two members", []byte("foo1"), []byte("fop"), allKeys[1:], allRevs[1:],
		},
		{
			"range keys, return last member", []byte("foo2"), []byte("fop"), allKeys[2:], allRevs[2:],
		},
		{
			"range keys, return nothing", []byte("foo3"), []byte("fop"), nil, nil,
		},
	}

	atRev := int64(3)
	for i, fixture := range fixtures {
		test.Run(fixture.desc, func(t *testing.T) {
			keys, revs := treeIdx.Range(fixture.key, fixture.end, atRev)
			if !reflect.DeepEqual(keys, fixture.wkeys) {
				t.Errorf("#%d: keys = %+v, want %+v", i, keys, fixture.wkeys)
			}
			if !reflect.DeepEqual(revs, fixture.wrevs) {
				t.Errorf("#%d: revs = %+v, want %+v", i, revs, fixture.wrevs)
			}
		})
	}
}

func TestTreeIndexRangeSince(test *testing.T) {
	treeIdx := newTreeIndex()
	allKeys := [][]byte{[]byte("foo"), []byte("foo1"), []byte("foo2"), []byte("foo2"), []byte("foo1"), []byte("foo")}
	allRevs := []revision{{main: 1}, {main: 2}, {main: 3}, {main: 4}, {main: 5}, {main: 6}}
	for i := range allKeys {
		treeIdx.Put(allKeys[i], allRevs[i])
	}

	fixtures := []struct {
		desc     string
		key, end []byte
		wrevs    []revision
	}{
		{
			"single key that not found", []byte("bar"), nil, nil,
		},
		{
			"single key that found", []byte("foo"), nil, []revision{{main: 1}, {main: 6}},
		},
		{
			"range keys, return first member", []byte("foo"), []byte("foo1"), []revision{{main: 1}, {main: 6}},
		},
		{
			"range keys, return first two members", []byte("foo"), []byte("foo2"), []revision{{main: 1}, {main: 2}, {main: 5}, {main: 6}},
		},
		{
			"range keys, return all members", []byte("foo"), []byte("fop"), allRevs,
		},
		{
			"range keys, return last two members", []byte("foo1"), []byte("fop"), []revision{{main: 2}, {main: 3}, {main: 4}, {main: 5}},
		},
		{
			"range keys, return last member", []byte("foo2"), []byte("fop"), []revision{{main: 3}, {main: 4}},
		},
		{
			"range keys, return nothing", []byte("foo3"), []byte("fop"), nil,
		},
	}

	atRev := int64(1)
	for i, fixture := range fixtures {
		test.Run(fixture.desc, func(t *testing.T) {
			revs := treeIdx.RangeSince(fixture.key, fixture.end, atRev)
			if !reflect.DeepEqual(revs, fixture.wrevs) {
				t.Errorf("#%d: revs = %+v, want %+v", i, revs, fixture.wrevs)
			}
		})
	}
}

func TestTreeIndexTombstone(test *testing.T) {
	treeIdx := newTreeIndex()
	treeIdx.Put([]byte("foo"), revision{main: 1})

	err := treeIdx.Tombstone([]byte("foo"), revision{main: 2})
	if err != nil {
		test.Errorf("tombstone error = %v, want nil", err)
	}

	_, _, _, err = treeIdx.Get([]byte("foo"), 2)
	if err != ErrRevisionNotFound {
		test.Errorf("get error = %v, want ErrRevisionNotFound", err)
	}

	err = treeIdx.Tombstone([]byte("foo"), revision{main: 3})
	if err != ErrRevisionNotFound {
		test.Errorf("tombstone error = %v, want %v", err, ErrRevisionNotFound)
	}
}

package raftlog

import (
	"reflect"
	"testing"

	"k8s-lx1036/k8s/storage/raft/proto"
	"k8s-lx1036/k8s/storage/raft/storage"
)

func TestMaybeLastIndex(test *testing.T) {
	fixtures := []struct {
		entries []*proto.Entry
		offset  uint64
		ok      bool
		index   uint64
	}{
		// last in entries
		{
			[]*proto.Entry{{Index: 5, Term: 1}}, 5, true, 5,
		},
		// empty unstable
		{
			[]*proto.Entry{}, 0, false, 0,
		},
	}

	for _, fixture := range fixtures {
		u := unstable{
			entries: fixture.entries,
			offset:  fixture.offset,
		}

		index, ok := u.maybeLastIndex()
		if ok != fixture.ok {
			test.Errorf("want %v, got %v", fixture.ok, ok)
		}

		if index != fixture.index {
			test.Errorf("want %v, got %v", fixture.index, index)
		}
	}
}

// RaftLog
func TestAppend(test *testing.T) {
	previousEntries := []*proto.Entry{{Index: 1, Term: 1}, {Index: 2, Term: 2}}
	fixtures := []struct {
		ents      []*proto.Entry
		windex    uint64
		wents     []*proto.Entry
		wunstable uint64
	}{
		{
			[]*proto.Entry{},
			2,
			[]*proto.Entry{{Index: 1, Term: 1}, {Index: 2, Term: 2}},
			3,
		},
		{
			[]*proto.Entry{{Index: 3, Term: 2}},
			3,
			[]*proto.Entry{{Index: 1, Term: 1}, {Index: 2, Term: 2}, {Index: 3, Term: 2}},
			3,
		},
		// conflicts with index 1
		// TODO 这里为啥是 {Index: 1, Term: 2} , {Index: 2, Term: 2} 被truncate了么？
		{
			[]*proto.Entry{{Index: 1, Term: 2}},
			1,
			[]*proto.Entry{{Index: 1, Term: 2}},
			1,
		},
		// conflicts with index 2
		{
			[]*proto.Entry{{Index: 2, Term: 3}, {Index: 3, Term: 3}},
			3,
			[]*proto.Entry{{Index: 1, Term: 1}, {Index: 2, Term: 3}, {Index: 3, Term: 3}},
			2,
		},
	}

	for i, fixture := range fixtures {
		s := storage.DefaultMemoryStorage()
		s.StoreEntries(previousEntries)

		log, err := newRaftLog(s)
		if err != nil {
			test.Fatalf("#%d: unexpected error %v", i, err)
		}
		index := log.append(fixture.ents...)
		if index != fixture.windex {
			test.Errorf("want %v, got %v", fixture.windex, index)
		}

		e, err := log.entries(1, noLimit)
		if err != nil {
			test.Fatalf("#%d: unexpected error %v", i, err)
		}
		if !reflect.DeepEqual(e, fixture.wents) {
			test.Errorf("want %v, got %v", fixture.wents, e)
		}

		if log.unstable.offset != fixture.wunstable {
			test.Errorf("want %v, got %v", fixture.wunstable, log.unstable.offset)
		}
	}
}

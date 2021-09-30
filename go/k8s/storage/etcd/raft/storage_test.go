package raft

import (
	"reflect"
	"testing"

	pb "go.etcd.io/etcd/raft/v3/raftpb"
)

func TestStorage(t *testing.T) {
	entries := []pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 4}, {Index: 5, Term: 5}}
	fixtures := []struct {
		desc          string
		entries       []pb.Entry
		wantedEntries []pb.Entry
		lastIndex     uint64
	}{
		{
			desc:          "truncate compacted entries",
			entries:       []pb.Entry{{Index: 1, Term: 1}, {Index: 2, Term: 2}},
			wantedEntries: []pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 4}, {Index: 5, Term: 5}},
			lastIndex:     5,
		},
		{
			desc:          "truncate entries",
			entries:       []pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 6}, {Index: 5, Term: 6}},
			wantedEntries: []pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 6}, {Index: 5, Term: 6}},
			lastIndex:     5,
		},
		{
			desc:          "truncate incoming entries, truncate the existing entries and append",
			entries:       []pb.Entry{{Index: 2, Term: 3}, {Index: 3, Term: 3}, {Index: 4, Term: 5}},
			wantedEntries: []pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 5}},
			lastIndex:     4,
		},
		{
			desc:          "truncate the existing entries and append",
			entries:       []pb.Entry{{Index: 4, Term: 5}},
			wantedEntries: []pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 5}},
			lastIndex:     4,
		},
		{
			desc:          "direct append",
			entries:       []pb.Entry{{Index: 6, Term: 5}},
			wantedEntries: []pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 4}, {Index: 5, Term: 5}, {Index: 6, Term: 5}},
			lastIndex:     6,
		},
	}

	for index, fixture := range fixtures {
		t.Run(fixture.desc, func(t *testing.T) {
			storage := &MemoryStorage{entries: entries}
			storage.Append(fixture.entries)
			if !reflect.DeepEqual(storage.entries, fixture.wantedEntries) {
				t.Errorf("#%d: entries = %v, want %v", index, storage.entries, fixture.wantedEntries)
			}
			last := storage.LastIndex()
			if last != fixture.lastIndex {
				t.Errorf("last = %d, want %d", last, 5)
			}
		})
	}
}

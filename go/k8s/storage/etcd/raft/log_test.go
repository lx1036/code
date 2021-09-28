package raft

import (
	"reflect"
	"testing"

	pb "go.etcd.io/etcd/raft/v3/raftpb"
)

func TestLogAppend(test *testing.T) {
	fixtures := []struct {
		desc          string
		entries       []pb.Entry
		lastIndex     uint64
		wantedEntries []pb.Entry
		offset        uint64
	}{
		{
			desc:          "empty entries",
			entries:       []pb.Entry{},
			lastIndex:     2,
			wantedEntries: []pb.Entry{{Index: 1, Term: 1}, {Index: 2, Term: 2}},
			offset:        3,
		},
		{
			desc:          "append one entry",
			entries:       []pb.Entry{{Index: 3, Term: 2}},
			lastIndex:     3,
			wantedEntries: []pb.Entry{{Index: 1, Term: 1}, {Index: 2, Term: 2}, {Index: 3, Term: 2}},
			offset:        3,
		},
		{
			desc:          "conflicts with index 1",
			entries:       []pb.Entry{{Index: 1, Term: 2}},
			lastIndex:     1,
			wantedEntries: []pb.Entry{{Index: 1, Term: 2}},
			offset:        1,
		},
		{
			// INFO: 自动会剔除掉过期的 entry
			desc:          "conflicts with index 2",
			entries:       []pb.Entry{{Index: 2, Term: 3}, {Index: 3, Term: 3}},
			lastIndex:     3,
			wantedEntries: []pb.Entry{{Index: 1, Term: 1}, {Index: 2, Term: 3}, {Index: 3, Term: 3}},
			offset:        2,
		},
	}

	previousEnts := []pb.Entry{{Index: 1, Term: 1}, {Index: 2, Term: 2}}
	for i, fixture := range fixtures {
		test.Run(fixture.desc, func(t *testing.T) {
			storage := NewMemoryStorage()
			storage.Append(previousEnts)
			log := newLog(storage)
			lastIndex := log.append(fixture.entries...)
			if lastIndex != fixture.lastIndex {
				test.Errorf("#%d: lastIndex = %d, want %d", i, lastIndex, fixture.lastIndex)
			}
			entries, err := log.entries(1, noLimit)
			if err != nil {
				t.Fatalf("#%d: unexpected error %v", i, err)
			}
			if !reflect.DeepEqual(entries, fixture.wantedEntries) {
				t.Errorf("#%d: logEnts = %+v, want %+v", i, entries, fixture.wantedEntries)
			}
			if offset := log.unstable.offset; offset != fixture.offset {
				t.Errorf("#%d: unstable = %d, want %d", i, offset, fixture.offset)
			}
		})
	}
}

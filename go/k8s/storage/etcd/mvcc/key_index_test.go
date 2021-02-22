package mvcc

import (
	"k8s.io/klog/v2"
	"testing"
)

// 查看KeyIndex，即generation是否为空
func TestKeyIndexIsEmpty(test *testing.T) {
	fixtures := []struct {
		keyIdx *keyIndex
		w      bool
	}{
		{
			&keyIndex{
				key:         []byte("foo"),
				generations: []generation{{}},
			},
			true,
		},
		{
			&keyIndex{
				key:      []byte("foo"),
				modified: revision{2, 0},
				generations: []generation{
					{created: revision{1, 0}, version: 2, revisions: []revision{{main: 2}}},
				},
			},
			false,
		},
	}

	for i, fixture := range fixtures {
		g := fixture.keyIdx.isEmpty()
		if g != fixture.w {
			test.Errorf("#%d: isEmpty = %v, want %v", i, g, fixture.w)
		}
	}
}

// 测试findGeneration()
func TestKeyIndexFindGeneration(test *testing.T) {
	keyIdx := newTestKeyIndex()

	fixtures := []struct {
		rev int64
		wg  *generation
	}{
		{0, nil},
		{1, nil},
		{2, &keyIdx.generations[0]},
		{3, &keyIdx.generations[0]},
		{4, &keyIdx.generations[0]},
		{5, &keyIdx.generations[0]},
		{6, nil},
		{7, nil},
		{8, &keyIdx.generations[1]},
		{9, &keyIdx.generations[1]},
		{10, &keyIdx.generations[1]},
		{11, &keyIdx.generations[1]},
		{12, nil},
		{13, nil},
	}

	for i, fixture := range fixtures {
		g := keyIdx.findGeneration(fixture.rev)
		if g != fixture.wg {
			test.Errorf("#%d: generation = %+v, want %+v", i, g, fixture.wg)
		}
	}
}

func TestName(test *testing.T) {
	newTestKeyIndex()
}

func newTestKeyIndex() *keyIndex {
	// key: "foo"
	// rev: 16
	// generations:
	//    {empty}
	//    {{14, 0}[1], {14, 1}[2], {16, 0}(t)[3]}
	//    {{8, 0}[1], {10, 0}[2], {12, 0}(t)[3]}
	//    {{2, 0}[1], {4, 0}[2], {6, 0}(t)[3]}

	keyIdx := &keyIndex{key: []byte("foo")}
	keyIdx.put(2, 0)
	keyIdx.put(4, 0)
	keyIdx.tombstone(6, 0)
	keyIdx.put(8, 0)
	keyIdx.put(10, 0)
	keyIdx.tombstone(12, 0)
	keyIdx.put(14, 0)
	keyIdx.put(14, 1)
	keyIdx.tombstone(16, 0)

	klog.Infof("modified revision main:%d sub:%d", keyIdx.modified.main, keyIdx.modified.sub)

	return keyIdx
}

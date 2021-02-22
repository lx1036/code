package mvcc

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/google/btree"
	"k8s.io/klog/v2"
)

var (
	ErrRevisionNotFound = errors.New("mvcc: revision not found")
)

type generation struct {
	version int64 // 表示此key的修改次数

	created revision // keyIndex创建时版本号，只需要赋值一次

	revisions []revision // 每次修改key时的revision追加到此数组
}

func (g *generation) isEmpty() bool {
	return g == nil || len(g.revisions) == 0
}

type keyIndex struct {
	key []byte

	modified revision // 最后一次修改key时的版本号revision

	generations []generation
}

func (keyIdx *keyIndex) Less(than btree.Item) bool {
	return bytes.Compare(keyIdx.key, than.(*keyIndex).key) == -1 // check if keyIdx.key < than.(*keyIndex).key
}

// findGeneration finds out the generation of the keyIndex that the
// given rev belongs to. If the given rev is at the gap of two generations,
// which means that the key does not exist at the given rev, it returns nil.
func (keyIdx *keyIndex) findGeneration(rev int64) *generation {
	//lastg := len(keyIdx.generations) - 1
	//cg := lastg

	return nil
}

func (keyIdx *keyIndex) get(rev int64) (modified, created revision, ver int64, err error) {
	if keyIdx.isEmpty() {
		klog.Errorf("'get' got an unexpected empty keyIndex key %s", string(keyIdx.key))
	}

	//g := keyIdx.findGeneration(rev)

	return revision{}, revision{}, 0, err
}

// put puts a revision to the keyIndex.
func (keyIdx *keyIndex) put(main, sub int64) {
	rev := revision{main: main, sub: sub}
	if !rev.GreaterThan(keyIdx.modified) {
		klog.Errorf("'put' with an unexpected smaller revision, given-revision-main %d, given-revision-sub %d, modified-revision-main %d, modified-revision-sub %d",
			rev.main, rev.sub, keyIdx.modified.main, keyIdx.modified.sub)
	}

	// 更新 modified
	keyIdx.modified = rev

	// 更新generations
	if len(keyIdx.generations) == 0 {
		keyIdx.generations = append(keyIdx.generations, generation{})
	}
	g := &keyIdx.generations[len(keyIdx.generations)-1] // 这里取指针，因为下面要修改这个generation值
	if len(g.revisions) == 0 {
		g.created = rev // keyIndex创建时版本号，只需要赋值一次
	}
	g.version++
	g.revisions = append(g.revisions, rev) // 每次修改key时的revision追加到此数组
}

func (keyIdx *keyIndex) isEmpty() bool {
	return len(keyIdx.generations) == 1 && keyIdx.generations[0].isEmpty()
}

// tombstone puts a revision, pointing to a tombstone, to the keyIndex.
// It also creates a new empty generation in the keyIndex.
// It returns ErrRevisionNotFound when tombstone on an empty generation.
func (keyIdx *keyIndex) tombstone(main, sub int64) error {
	if keyIdx.isEmpty() {
		errMsg := fmt.Sprintf("'tombstone' got an unexpected empty keyIndex %s", string(keyIdx.key))
		klog.Errorf(errMsg)
		return errors.New(errMsg)
	}

	if keyIdx.generations[len(keyIdx.generations)-1].isEmpty() { // last generation isEmpty
		return ErrRevisionNotFound
	}

	keyIdx.put(main, sub)
	keyIdx.generations = append(keyIdx.generations, generation{}) // append一个空generation{}对象，表示删除这个keyIndex
	return nil
}

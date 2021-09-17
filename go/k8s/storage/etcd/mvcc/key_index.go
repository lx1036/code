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

// INFO: 从队尾开始查找，找到revision.main小于等于当前的revision的第一个revision，返回其index
func (g *generation) walk(f func(rev revision) bool) int {
	l := len(g.revisions)
	for i := range g.revisions {
		index := l - 1 - i // 从队尾开始
		ok := f(g.revisions[index])
		if !ok {
			return index
		}
	}

	return -1
}

type keyIndex struct {
	key []byte

	modified revision // 最后一次修改key时的版本号revision

	generations []generation
}

func (keyIdx *keyIndex) Less(than btree.Item) bool {
	return bytes.Compare(keyIdx.key, than.(*keyIndex).key) == -1 // check if keyIdx.key < than.(*keyIndex).key
}

// INFO: 根据指定atRev从generations里查找
func (keyIdx *keyIndex) findGeneration(atRev int64) *generation {
	lastg := len(keyIdx.generations) - 1
	cg := lastg

	for cg >= 0 {
		// 如果最后一个generation是dummy generation，这种情况一般是最后一个操作是直接tombstone，跳过
		if len(keyIdx.generations[cg].revisions) == 0 {
			cg--
			continue
		}

		// INFO: 检查atRev是不是在，一个generation里最高和最低revision之间
		g := keyIdx.generations[cg]
		if cg != lastg {
			// 最后删除操作的事务版本号，比atRev还小，返回空
			if tomb := g.revisions[len(g.revisions)-1].main; tomb <= atRev {
				return nil
			}
		}
		if g.revisions[0].main <= atRev {
			return &g
		}

		cg--
	}

	return nil
}

// INFO: 获取该key的最新修改版本，同时还有创建时版本
func (keyIdx *keyIndex) get(atRev int64) (modified, created revision, ver int64, err error) {
	if keyIdx.isEmpty() {
		klog.Errorf("[get]got an unexpected empty keyIndex key %s", string(keyIdx.key))
	}

	g := keyIdx.findGeneration(atRev)
	if g.isEmpty() {
		return revision{}, revision{}, 0, ErrRevisionNotFound
	}

	n := g.walk(func(rev revision) bool { return rev.main > atRev })
	if n != -1 {
		return g.revisions[n], g.created, g.version - int64(len(g.revisions)-n-1), nil
	}

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
	// 这里取指针，因为下面要修改这个generation值
	g := &keyIdx.generations[len(keyIdx.generations)-1]
	if len(g.revisions) == 0 {
		g.created = rev // keyIndex创建时版本号，只需要赋值一次
	}
	g.version++
	g.revisions = append(g.revisions, rev) // 每次修改key时的revision追加到此数组
}

func (keyIdx *keyIndex) isEmpty() bool {
	return len(keyIdx.generations) == 1 && keyIdx.generations[0].isEmpty()
}

// INFO: 删除revision,会创建tombstone revision
func (keyIdx *keyIndex) tombstone(main, sub int64) error {
	if keyIdx.isEmpty() {
		errMsg := fmt.Sprintf("[tombstone]got an unexpected empty keyIndex %s", string(keyIdx.key))
		klog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	if keyIdx.generations[len(keyIdx.generations)-1].isEmpty() { // last generation isEmpty
		return ErrRevisionNotFound
	}

	keyIdx.put(main, sub)
	keyIdx.generations = append(keyIdx.generations, generation{}) // append一个空generation{}对象，表示删除这个keyIndex
	return nil
}

func (keyIdx *keyIndex) since(rev int64) []revision {
	if keyIdx.isEmpty() {
		klog.Errorf("[since]got an unexpected empty keyIndex key %s", string(keyIdx.key))
	}

	since := revision{rev, 0}
	var gi int
	// INFO: 找到revisio比rev大，且最大的那一个generation
	for gi = len(keyIdx.generations) - 1; gi > 0; gi-- {
		g := keyIdx.generations[gi]
		if g.isEmpty() {
			continue
		}
		if since.GreaterThan(g.created) {
			break
		}
	}

	var revs []revision
	var last int64
	for ; gi < len(keyIdx.generations); gi++ {
		for _, r := range keyIdx.generations[gi].revisions {
			if since.GreaterThan(r) {
				continue
			}
			// INFO: 只取第一个 revision
			if r.main == last {
				// replace the revision with a new one that has higher sub value,
				// because the original one should not be seen by external
				revs[len(revs)-1] = r
				continue
			}
			revs = append(revs, r)
			last = r.main
		}
	}

	return revs
}

package radix_tree

import "testing"

/**
https://github.com/trustfeed/radix-tree-go
 */
func TestTrie(test *testing.T) {
	t0 := New()
	if t0.Search([]byte{0}) != nil {
		test.Errorf("Trie contains data for missing key")
	}

	t1 := t0.Insert([]byte{0}, []byte("test"))
}

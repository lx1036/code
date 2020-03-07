package kube_gin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInsertOnePattern(test *testing.T) {
	root := &node{}
	pattern := "/people/:id/accounts"
	parts := parsePattern(pattern) // []string{"people", ":id", "accounts"}
	root.insert(pattern, parts, 0)

	assert.Equal(test, "", root.pattern)
	assert.Equal(test, "", root.part)
	assert.Equal(test, 1, len(root.children))
	assert.False(test, root.isWild)

	peopleNode := root.children[0]
	assert.Equal(test, "", peopleNode.pattern)
	assert.Equal(test, "people", peopleNode.part)
	assert.Equal(test, 1, len(peopleNode.children))
	assert.False(test, peopleNode.isWild)

	idNode := peopleNode.children[0]
	assert.Equal(test, "", idNode.pattern)
	assert.Equal(test, ":id", idNode.part)
	assert.Equal(test, 1, len(idNode.children))
	assert.True(test, idNode.isWild)

	accountNode := idNode.children[0]
	assert.Equal(test, pattern, accountNode.pattern)
	assert.Equal(test, "accounts", accountNode.part)
	assert.Equal(test, 0, len(accountNode.children))
	assert.False(test, accountNode.isWild)
}

func TestInsertMultiplePatterns(test *testing.T) {
	root := &node{}
	pattern1 := "/people/:id/accounts"
	parts1 := parsePattern(pattern1) // []string{"people", ":id", "accounts"}
	root.insert(pattern1, parts1, 0)

	pattern2 := "/people/:id/houses"
	parts2 := parsePattern(pattern2) // []string{"people", ":id", "houses"}
	root.insert(pattern2, parts2, 0)

	assert.Equal(test, "", root.pattern)
	assert.Equal(test, "", root.part)
	assert.Equal(test, 1, len(root.children))
	assert.False(test, root.isWild)

	peopleNode := root.children[0]
	assert.Equal(test, "", peopleNode.pattern)
	assert.Equal(test, "people", peopleNode.part)
	assert.Equal(test, 1, len(peopleNode.children))
	assert.False(test, peopleNode.isWild)

	idNode := peopleNode.children[0]
	assert.Equal(test, "", idNode.pattern)
	assert.Equal(test, ":id", idNode.part)
	assert.Equal(test, 2, len(idNode.children))
	assert.True(test, idNode.isWild)

	accountNode := idNode.children[0]
	assert.Equal(test, pattern1, accountNode.pattern)
	assert.Equal(test, "accounts", accountNode.part)
	assert.Equal(test, 0, len(accountNode.children))
	assert.False(test, accountNode.isWild)

	houseNode := idNode.children[1]
	assert.Equal(test, pattern2, houseNode.pattern)
	assert.Equal(test, "houses", houseNode.part)
	assert.Equal(test, 0, len(houseNode.children))
	assert.False(test, houseNode.isWild)
}

func TestSearch(test *testing.T) {
	root := &node{}
	pattern1 := "/people/:id/accounts"
	parts1 := parsePattern(pattern1) // []string{"people", ":id", "accounts"}
	root.insert(pattern1, parts1, 0)

	pattern2 := "/people/:id/houses"
	parts2 := parsePattern(pattern2) // []string{"people", ":id", "houses"}
	root.insert(pattern2, parts2, 0)

	houseNode := root.search(parts2, 0)
	assert.Equal(test, pattern2, houseNode.pattern)
	assert.Equal(test, "houses", houseNode.part)
	assert.Equal(test, 0, len(houseNode.children))
	assert.False(test, houseNode.isWild)

	pattern3 := "/banks/*id/accounts"
	parts3 := parsePattern(pattern3)
	root.insert(pattern3, parts3, 0)
	bankNode := root.search(parsePattern("/banks/123"), 0)
	assert.Equal(test, pattern3, bankNode.pattern)
	assert.Equal(test, "*id", bankNode.part)
	assert.Equal(test, 0, len(bankNode.children))
	assert.True(test, bankNode.isWild)
}

package _95_max_area_island

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func maxAreaOfIsland(grid [][]int) int {
	ans := 0
	for i := 0; i < len(grid); i++ {
		for j := 0; j < len(grid[0]); j++ {
			ans = int(math.Max(float64(ans), float64(dfs(grid, i, j))))
		}
	}

	return ans
}

// 深度优先 https://leetcode-cn.com/problems/max-area-of-island/solution/dao-yu-de-zui-da-mian-ji-by-leetcode-solution/
// 时间复杂度：O(m*n)，空间复杂度：O(m*n)
func dfs(grid [][]int, i, j int) int {
	if i >= 0 && j >= 0 && i < len(grid) && j < len(grid[0]) { // 边界条件
		if grid[i][j] == 0 { // 边界条件
			return 0
		} else {
			grid[i][j] = 0
			return 1 + dfs(grid, i-1, j) + dfs(grid, i+1, j) + dfs(grid, i, j-1) + dfs(grid, i, j+1)
		}
	}

	return 0
}

func TestDFS(test *testing.T) {
	islands := [][]int{
		{0,0,1,0,0,0,0,1,0,0,0,0,0},
		{0,0,0,0,0,0,0,1,1,1,0,0,0},
		{0,1,1,0,1,0,0,0,0,0,0,0,0},
		{0,1,0,0,1,1,0,0,1,0,1,0,0},
		{0,1,0,0,1,1,0,0,1,1,1,0,0},
		{0,0,0,0,0,0,0,0,0,0,1,0,0},
		{0,0,0,0,0,0,0,1,1,1,0,0,0},
		{0,0,0,0,0,0,0,1,1,0,0,0,0},
	}

	area := maxAreaOfIsland(islands)
	assert.Equal(test, 6, area)

	islands = [][]int{
		{0,0,0,0,0,0,0,0},
	}

	area = maxAreaOfIsland(islands)
	assert.Equal(test, 0, area)
}

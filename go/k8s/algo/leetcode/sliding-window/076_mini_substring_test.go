package sliding_window

// 字节面试题

// https://leetcode-cn.com/problems/minimum-window-substring/

func minWindow(s string, t string) string {

	tHashTable, sHashTable := map[byte]int{}, map[byte]int{}
	for i := 0; i < len(t); i++ {
		tHashTable[t[i]]++
	}

	//l, r := 0, 0
	for r := 0; r < len(s); r++ {
		if tHashTable[s[r]] > 0 {
			sHashTable[s[r]]++
		}

	}

	return ""
}

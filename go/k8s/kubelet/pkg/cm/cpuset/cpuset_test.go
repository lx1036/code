package cpuset

import "testing"

func TestCPUSetString(t *testing.T) {
	testCases := []struct {
		set      CPUSet
		expected string
	}{
		{NewCPUSet(), ""},
		{NewCPUSet(5), "5"},
		{NewCPUSet(1, 2, 3, 4, 5), "1-5"},
		{NewCPUSet(1, 2, 3, 5, 6, 8), "1-3,5-6,8"},
	}

	for _, c := range testCases {
		result := c.set.String()
		if result != c.expected {
			t.Fatalf("expected set as string to be %s (got %s), s: [%v]", c.expected, result, c.set)
		}
	}
}

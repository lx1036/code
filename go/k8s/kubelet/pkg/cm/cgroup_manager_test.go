package cm

import (
	"path"
	"reflect"
	"testing"
)

func TestCgroupNameToCgroupfs(t *testing.T) {
	testCases := []struct {
		input    CgroupName
		expected string
	}{
		{
			input:    RootCgroupName,
			expected: "/",
		},
		{
			input:    NewCgroupName(RootCgroupName, "Burstable"),
			expected: "/Burstable",
		},
	}

	for _, testCase := range testCases {
		if actual := testCase.input.ToCgroupfs(); actual != testCase.expected {
			t.Errorf("Unexpected result, input: %v, expected: %v, actual: %v", testCase.input, testCase.expected, actual)
		}
	}
}

func TestCgroupNameToSystemd(t *testing.T) {
	testCases := []struct {
		input    CgroupName
		expected string
	}{
		{
			input:    RootCgroupName,
			expected: "/",
		},
		{
			input:    NewCgroupName(RootCgroupName, "Burstable"),
			expected: "/Burstable.slice",
		},
		{
			input:    NewCgroupName(RootCgroupName, "Burstable", "pod-123"),
			expected: "/Burstable.slice/Burstable-pod_123.slice",
		},
		{
			input:    NewCgroupName(RootCgroupName, "BestEffort", "pod-6c1a4e95-6bb6-11e6-bc26-28d2444e470d"),
			expected: "/BestEffort.slice/BestEffort-pod_6c1a4e95_6bb6_11e6_bc26_28d2444e470d.slice",
		},
		{
			input:    NewCgroupName(RootCgroupName, "kubepods"),
			expected: "/kubepods.slice",
		},
	}
	for _, testCase := range testCases {
		if actual := testCase.input.ToSystemd(); actual != testCase.expected {
			t.Errorf("Unexpected result, input: %v, expected: %v, actual: %v", testCase.input, testCase.expected, actual)
		}
	}
}

func TestCgroupNameToSystemdBasename(t *testing.T) {
	testCases := []struct {
		input    CgroupName
		expected string
	}{
		{
			input:    RootCgroupName,
			expected: "/",
		},
		{
			input:    NewCgroupName(RootCgroupName, "system"),
			expected: "system.slice",
		},
		{
			input:    NewCgroupName(RootCgroupName, "system", "Burstable"),
			expected: "system-Burstable.slice",
		},
		{
			input:    NewCgroupName(RootCgroupName, "Burstable", "pod-123"),
			expected: "Burstable-pod_123.slice",
		},
		{
			input:    NewCgroupName(RootCgroupName, "test", "a", "b"),
			expected: "test-a-b.slice",
		},
		{
			input:    NewCgroupName(RootCgroupName, "test", "a", "b", "Burstable"),
			expected: "test-a-b-Burstable.slice",
		},
		{
			input:    NewCgroupName(RootCgroupName, "Burstable"),
			expected: "Burstable.slice",
		},
		{
			input:    NewCgroupName(RootCgroupName, "BestEffort", "pod-6c1a4e95-6bb6-11e6-bc26-28d2444e470d"),
			expected: "BestEffort-pod_6c1a4e95_6bb6_11e6_bc26_28d2444e470d.slice",
		},
	}
	for _, testCase := range testCases {
		if actual := path.Base(testCase.input.ToSystemd()); actual != testCase.expected {
			t.Errorf("Unexpected result, input: %v, expected: %v, actual: %v", testCase.input, testCase.expected, actual)
		}
	}
}

func TestNewCgroupName(t *testing.T) {
	a := ParseCgroupfsToCgroupName("/a/")
	ab := NewCgroupName(a, "b")
	expectedAB := CgroupName([]string{"a", "", "b"})
	if !reflect.DeepEqual(ab, expectedAB) {
		t.Errorf("Expected %d%+v; got %d%+v", len(expectedAB), expectedAB, len(ab), ab)
	}

	abc := NewCgroupName(ab, "c")
	expectedABC := CgroupName([]string{"a", "", "b", "c"})
	if !reflect.DeepEqual(abc, expectedABC) {
		t.Errorf("Expected %d%+v; got %d%+v", len(expectedABC), expectedABC, len(abc), abc)
	}

	abd := NewCgroupName(ab, "d")
	expectedABD := CgroupName([]string{"a", "", "b", "d"})
	if !reflect.DeepEqual(abd, expectedABD) {
		t.Errorf("Expected %d%+v; got %d%+v", len(expectedABC), expectedABC, len(abc), abc)
	}
}

func TestParseSystemdToCgroupName(t *testing.T) {
	testCases := []struct {
		input    string
		expected CgroupName
	}{
		{
			input:    "/test",
			expected: []string{"test"},
		},
		{
			input:    "/test.slice",
			expected: []string{"test"},
		},
	}

	for _, testCase := range testCases {
		if actual := ParseSystemdToCgroupName(testCase.input); !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("Unexpected result, input: %v, expected: %v, actual: %v", testCase.input, testCase.expected, actual)
		}
	}
}

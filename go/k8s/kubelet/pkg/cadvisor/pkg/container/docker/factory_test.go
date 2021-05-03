package docker

import "testing"

func TestIsContainerName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{
			name:     "/system.slice/var-lib-docker-overlay-9f086b233ab7c786bf8b40b164680b658a8f00e94323868e288d6ce20bc92193-merged.mount",
			expected: false,
		},
		{
			name:     "/system.slice/docker-72e5a5ff5eef3c4222a6551b992b9360a99122f77d2229783f0ee0946dfd800e.scope",
			expected: true,
		},
	}
	for _, test := range tests {
		actual := isContainerName(test.name)
		if actual != test.expected {
			t.Errorf("%s: expected: %v, actual: %v", test.name, test.expected, actual)
		}
	}
}

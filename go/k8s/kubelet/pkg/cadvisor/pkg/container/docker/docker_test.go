package docker

import (
	"reflect"
	"regexp"
	"testing"
)

func TestParseDockerAPIVersion(t *testing.T) {
	tests := []struct {
		version       string
		regex         *regexp.Regexp
		length        int
		expected      []int
		expectedError string
	}{
		{"17.03.0", versionRe, 3, []int{17, 03, 0}, ""},
		{"17.a3.0", versionRe, 3, []int{}, `version string "17.a3.0" doesn't match expected regular expression: "(\d+)\.(\d+)\.(\d+)"`},
		{"1.20", apiVersionRe, 2, []int{1, 20}, ""},
		{"1.a", apiVersionRe, 2, []int{}, `version string "1.a" doesn't match expected regular expression: "(\d+)\.(\d+)"`},
	}

	for _, test := range tests {
		actual, err := parseVersion(test.version, test.regex, test.length)
		if err != nil {
			if len(test.expectedError) == 0 {
				t.Errorf("%s: expected no error, got %v", test.version, err)
			} else if err.Error() != test.expectedError {
				t.Errorf("%s: expected error %v, got %v", test.version, test.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("%s: expected array %v, got %v", test.version, test.expected, actual)
			}
		}
	}
}

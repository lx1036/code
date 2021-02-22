package leaderelection

import "testing"

func Test_sanitizeName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			"requires no change",
			"test-driver",
			"test-driver",
		},
		{
			"has characters that should be replaced",
			"test!driver/foo",
			"test-driver-foo",
		},
		{
			"has trailing space",
			"driver\\",
			"driver-X",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := sanitizeName(test.input)
			if output != test.output {
				t.Logf("expected name: %q", test.output)
				t.Logf("actual name: %q", output)
				t.Errorf("unexpected santized name")
			}
		})
	}
}

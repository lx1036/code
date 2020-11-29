package flags

import (
	"gotest.tools/assert"
	"net/url"
	"testing"
)

func TestUriString(test *testing.T) {
	tests := []struct {
		uri  Uri
		want string
	}{
		{
			uri: Uri{
				Key:   "abc",
				Value: url.URL{},
			},
			want: "abc",
		},
		{
			uri: Uri{
				Key: "kubernetes",
				Value: url.URL{
					Scheme:   "http",
					Host:     "localhost:8080",
					RawQuery: "key1=value1&key2=value2",
				},
			},
			want: "kubernetes:http://localhost:8080?key1=value1&key2=value2",
		},
	}

	for _, t := range tests {
		assert.Equal(test, t.want, t.uri.String())
	}
}

func TestUriSet(test *testing.T) {

}

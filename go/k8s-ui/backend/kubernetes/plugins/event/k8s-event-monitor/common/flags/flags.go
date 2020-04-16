package flags

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strings"
)


type Uri struct {
	Key string
	Value url.URL
}

func (uri *Uri)Set(value string) error {
	str := strings.SplitN(value, ":", 2)
	if str[0] == "" {
		return fmt.Errorf("missing key in {%s}", value)
	}
	uri.Key = str[0]
	if len(str) > 1 && str[1] != "" {
		u, err := url.Parse(os.ExpandEnv(str[1]))
		if err != nil {
			return err
		}
		uri.Value = *u
	}

	return nil
}

func (uri *Uri) String() string {
	value := uri.Value.String()
	if value == "" {
		return fmt.Sprintf("%s", uri.Key)
	}
	return fmt.Sprintf("%s:%s", uri.Key, value)
}

type Uris []Uri

func (uris *Uris) Set(value string) error {
	uri := Uri{}
	if err := uri.Set(value); err != nil {
		return err
	}

	*uris = append(*uris, uri)
	return nil
}

func (uris *Uris)String() string {
	var buffer bytes.Buffer
	buffer.WriteString("[")
	for i, uri := range *uris {
		if i > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString(uri.String())
	}
	buffer.WriteString("]")

	return buffer.String()
}

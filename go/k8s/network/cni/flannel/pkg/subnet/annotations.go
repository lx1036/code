package subnet

import (
	"errors"
	"regexp"
	"strings"
)

type annotations struct {
	SubnetKubeManaged          string
	BackendData                string
	BackendV6Data              string
	BackendType                string
	BackendPublicIP            string
	BackendPublicIPv6          string
	BackendPublicIPOverwrite   string
	BackendPublicIPv6Overwrite string
}

func newAnnotations(prefix string) (annotations, error) {
	slashCnt := strings.Count(prefix, "/")
	if slashCnt > 1 {
		return annotations{}, errors.New("subnet/kube: prefix can contain at most single slash")
	}
	if slashCnt == 0 {
		prefix += "/"
	}
	if !strings.HasSuffix(prefix, "/") && !strings.HasSuffix(prefix, "-") {
		prefix += "-"
	}

	// matches is a regexp matching the format used by the kubernetes for
	// annotations. Following rules apply:
	//
	//	- must start with FQDN - must contain at most one slash "/"
	//	- must contain only lowercase letters, nubers, underscores,
	//	  hyphens, dots and slash
	matches, err := regexp.MatchString(`(?:[a-z0-9_-]+\.)+[a-z0-9_-]+/(?:[a-z0-9_-]+-)?$`, prefix)
	if err != nil {
		panic(err)
	}
	if !matches {
		return annotations{}, errors.New("subnet/kube: prefix must be in a format: fqdn/[0-9a-z-_]*")
	}

	a := annotations{
		SubnetKubeManaged:          prefix + "kube-subnet-manager",
		BackendData:                prefix + "backend-data",
		BackendV6Data:              prefix + "backend-v6-data",
		BackendType:                prefix + "backend-type",
		BackendPublicIP:            prefix + "public-ip",
		BackendPublicIPOverwrite:   prefix + "public-ip-overwrite",
		BackendPublicIPv6:          prefix + "public-ipv6",
		BackendPublicIPv6Overwrite: prefix + "public-ipv6-overwrite",
	}

	return a, nil
}

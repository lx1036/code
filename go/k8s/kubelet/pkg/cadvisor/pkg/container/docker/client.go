package docker

import (
	"net/http"
	"sync"

	dclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
)

var (
	dockerClient     *dclient.Client
	dockerClientErr  error
	dockerClientOnce sync.Once
)

// Client creates a Docker API client based on the given Docker flags
func Client() (*dclient.Client, error) {
	dockerClientOnce.Do(func() {
		var client *http.Client
		if *ArgDockerTLS {
			client = &http.Client{}
			options := tlsconfig.Options{
				CAFile:             *ArgDockerCA,
				CertFile:           *ArgDockerCert,
				KeyFile:            *ArgDockerKey,
				InsecureSkipVerify: false,
			}
			tlsc, err := tlsconfig.Client(options)
			if err != nil {
				dockerClientErr = err
				return
			}
			client.Transport = &http.Transport{
				TLSClientConfig: tlsc,
			}
		}
		dockerClient, dockerClientErr = dclient.NewClientWithOpts(
			dclient.WithHost(*ArgDockerEndpoint),
			dclient.WithHTTPClient(client),
			dclient.WithAPIVersionNegotiation())
	})

	return dockerClient, dockerClientErr
}

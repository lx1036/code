package libdocker

// newKubeDockerClient creates an kubeDockerClient from an existing docker client. If requestTimeout is 0,
// defaultTimeout will be applied.
func newKubeDockerClient(dockerClient *dockerapi.Client, requestTimeout, imagePullProgressDeadline time.Duration) Interface {
	if requestTimeout == 0 {
		requestTimeout = defaultTimeout
	}

}

package cadvisor

// imageFsInfoProvider knows how to translate the configured runtime
// to its file system label for images.
type imageFsInfoProvider struct {
	runtime         string
	runtimeEndpoint string
}

// NewImageFsInfoProvider returns a provider for the specified runtime configuration.
func NewImageFsInfoProvider(runtime, runtimeEndpoint string) ImageFsInfoProvider {
	return &imageFsInfoProvider{runtime: runtime, runtimeEndpoint: runtimeEndpoint}
}





```go

// cmd/kubelet/app/server.go
if kubeDeps.CAdvisorInterface == nil {
	// s.ContainerRuntime="docker", s.RemoteRuntimeEndpoint="unix:///var/run/dockershim.sock"
    imageFsInfoProvider := cadvisor.NewImageFsInfoProvider(s.ContainerRuntime, s.RemoteRuntimeEndpoint)
    kubeDeps.CAdvisorInterface, err = cadvisor.New(imageFsInfoProvider, s.RootDirectory, cgroupRoots, cadvisor.UsingLegacyCadvisorStats(s.ContainerRuntime, s.RemoteRuntimeEndpoint))
    if err != nil {
        return err
    }
}

```

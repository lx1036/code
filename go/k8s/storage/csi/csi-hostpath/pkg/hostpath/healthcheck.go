package hostpath

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type MountPointInfo struct {
	Target              string           `json:"target"`
	Source              string           `json:"source"`
	FsType              string           `json:"fstype"`
	Options             string           `json:"options"`
	ContainerFileSystem []MountPointInfo `json:"children,omitempty"`
}

type ContainerFileSystem struct {
	Children []MountPointInfo `json:"children"`
}

type FileSystems struct {
	Filsystem []ContainerFileSystem `json:"filesystems"`
}

func locateCommandPath(commandName string) string {
	// default to root
	binary := filepath.Join("/", commandName)
	for _, path := range []string{"/bin", "/usr/sbin", "/usr/bin"} {
		binPath := filepath.Join(path, binary)
		if _, err := os.Stat(binPath); err != nil {
			continue
		}

		return binPath
	}

	return ""
}

func parseMountInfo(originalMountInfo []byte) ([]MountPointInfo, error) {
	fs := FileSystems{
		Filsystem: make([]ContainerFileSystem, 0),
	}

	if err := json.Unmarshal(originalMountInfo, &fs); err != nil {
		return nil, err
	}

	if len(fs.Filsystem) <= 0 {
		return nil, fmt.Errorf("failed to get mount info")
	}

	return fs.Filsystem[0].Children, nil
}

package cgroups

import (
	"github.com/docker/go-units"
	"io/ioutil"
	"strings"
)

var HugePageSizeUnitList = []string{"B", "KB", "MB", "GB", "TB", "PB"}

type Mount struct {
	Mountpoint string
	Root       string
	Subsystems []string
}

func GetHugePageSize() ([]string, error) {
	// INFO: 这里mock使用本地目录文件, linux上是 /sys/kernel/mm/hugepages
	files, err := ioutil.ReadDir("./mock/sys/kernel/mm/hugepages")
	if err != nil {
		return []string{}, err
	}
	var fileNames []string
	for _, st := range files {
		fileNames = append(fileNames, st.Name())
	}
	return getHugePageSizeFromFilenames(fileNames)
}

func getHugePageSizeFromFilenames(fileNames []string) ([]string, error) {
	var pageSizes []string
	for _, fileName := range fileNames {
		nameArray := strings.Split(fileName, "-")
		pageSize, err := units.RAMInBytes(nameArray[1])
		if err != nil {
			return []string{}, err
		}
		sizeString := units.CustomSize("%g%s", float64(pageSize), 1024.0, HugePageSizeUnitList)
		pageSizes = append(pageSizes, sizeString)
	}

	return pageSizes, nil
}

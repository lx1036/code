package cgroups

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func getCgroupMountsV1(all bool) ([]Mount, error) {
	path, err := filepath.Abs("./fixtures/proc/self/mountinfo")
	if err != nil {
		panic(err)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	path, err = filepath.Abs("./fixtures/proc/self/cgroup")
	if err != nil {
		panic(err)
	}
	allSubsystems, err := ParseCgroupFile(path)
	if err != nil {
		return nil, err
	}

	allMap := make(map[string]bool)
	for s := range allSubsystems {
		allMap[s] = false
	}
	return getCgroupMountsHelper(allMap, f, all)
}

func getCgroupMountsHelper(ss map[string]bool, mi io.Reader, all bool) ([]Mount, error) {
	res := make([]Mount, 0, len(ss))
	scanner := bufio.NewScanner(mi)
	numFound := 0
	for scanner.Scan() && numFound < len(ss) {
		txt := scanner.Text()
		sepIdx := strings.Index(txt, " - ")
		if sepIdx == -1 {
			return nil, fmt.Errorf("invalid mountinfo format")
		}
		if txt[sepIdx+3:sepIdx+10] == "cgroup2" || txt[sepIdx+3:sepIdx+9] != "cgroup" {
			continue
		}
		fields := strings.Split(txt, " ")
		m := Mount{
			Mountpoint: fields[4],
			Root:       fields[3],
		}
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			seen, known := ss[opt]
			if !known || (!all && seen) {
				continue
			}
			ss[opt] = true
			opt = strings.TrimPrefix(opt, CgroupNamePrefix)
			m.Subsystems = append(m.Subsystems, opt)
			numFound++
		}
		if len(m.Subsystems) > 0 || all {
			res = append(res, m)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

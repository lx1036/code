package myraft

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
)

const SegMeta = ".SM."

type Segment struct {
	first int64 // first index, static
	last  int64 // last index, dymanic
	path  string
}

// RecoverSegment recover segment from existed file
func RecoverSegment(name, path string) (*Segment, error) {
	index, err := strconv.Atoi(name[:20])
	if err != nil {
		return nil, err
	}
	filename := filepath.Join(path, name)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if string(content) != SegMeta {
		return nil, fmt.Errorf("invalid segment meta content")
	}

	segment := &Segment{
		first: int64(index),
		path:  path,
	}

	return segment, nil
}

package s3

import (
	"strconv"
	"strings"

	"github.com/google/btree"
)

type RelationType int

const (
	RelationNone RelationType = iota
	RelationInterset
	RelationContain
)

type IOVector struct {
	Ranges []*Range
}

func (i *IOVector) String() string {
	ss := make([]string, 0)
	for _, rg := range i.Ranges {
		ss = append(ss, rg.String())
	}
	vec := strings.Join(ss, ",")
	return vec
}

func (i *IOVector) Len() int {
	length := 0
	for _, rg := range i.Ranges {
		length += int(rg.Length)
	}
	return length
}

type Range struct {
	Offset int64
	Length int64
}

func (r *Range) Less(than btree.Item) bool {
	td, _ := than.(*Range)
	return r.Offset < td.Offset
}

func (r *Range) Copy() btree.Item {
	return &Range{
		Offset: r.Offset,
		Length: r.Length,
	}
}

func (r *Range) Equal(rg *Range) bool {
	return r.Offset == rg.Offset && r.Length == rg.Length
}

func (r *Range) String() string {
	off := strconv.FormatInt(r.Offset, 10)
	len := strconv.FormatInt(r.Length, 10)
	return off + "-" + len
}

func (r *Range) Relation(rg *Range) RelationType {
	end := r.Offset + r.Length
	rgEnd := rg.Offset + rg.Length

	if end <= rg.Offset || rgEnd <= r.Offset {
		return RelationNone
	} else if r.Offset <= rg.Offset && rgEnd <= end {
		return RelationContain
	}
	return RelationInterset
}

func (r *Range) Merge(rg *Range) (*Range, bool) {
	if r.Length <= 0 {
		return r, false
	}

	end := r.Offset + r.Length
	rgEnd := rg.Offset + rg.Length
	if rgEnd < r.Offset || end < rg.Offset {
		return r, false
	} else if rgEnd <= end {
		if rg.Offset < r.Offset {
			r.Length += (r.Offset - rg.Offset)
			r.Offset = rg.Offset
		}
	} else {
		if rg.Offset < r.Offset {
			r.Length += (r.Offset - rg.Offset)
			r.Offset = rg.Offset
		}
		r.Length += (rgEnd - end)
	}
	return r, true
}

func (r *Range) Sub(rg *Range) (*Range, bool) {
	if r.Length <= 0 {
		return nil, false
	}

	end := r.Offset + r.Length
	rgEnd := rg.Offset + rg.Length
	if rgEnd <= r.Offset || end <= rg.Offset {
		return nil, false
	} else if rgEnd <= end {
		if rg.Offset <= r.Offset {
			r.Length -= (rgEnd - r.Offset)
			r.Offset = rgEnd
			return nil, true
		} else {
			tmp := &Range{}
			tmp.Offset = r.Offset
			tmp.Length = rg.Offset - r.Offset
			r.Length -= (rgEnd - r.Offset)
			r.Offset = rgEnd
			return tmp, true
		}
	} else {
		if rg.Offset <= r.Offset {
			r.Offset = end
			r.Length = 0
			return nil, false
		} else {
			tmp := &Range{}
			tmp.Offset = r.Offset
			tmp.Length = rg.Offset - r.Offset
			r.Offset = end
			r.Length = 0
			return tmp, true
		}
	}
}

package cpuset

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// CPUSet is a thread-safe, immutable set-like data structure for CPU IDs.
type CPUSet struct {
	elems map[int]struct{}
}

func (s CPUSet) IsEmpty() bool {
	return s.Size() == 0
}

func (s CPUSet) Size() int {
	return len(s.elems)
}

// ToSlice returns a slice of integers that contains all elements from
// this set.
func (s CPUSet) ToSlice() []int {
	var result []int
	for cpu := range s.elems {
		result = append(result, cpu)
	}
	sort.Ints(result)
	return result
}

// String returns a new string representation of the elements in this CPU set
// in canonical linux CPU list format.
//
// See: http://man7.org/linux/man-pages/man7/cpuset.7.html#FORMATS
func (s CPUSet) String() string {
	if s.IsEmpty() {
		return ""
	}

	elems := s.ToSlice()

	type rng struct {
		start int
		end   int
	}

	ranges := []rng{{elems[0], elems[0]}}

	for i := 1; i < len(elems); i++ {
		lastRange := &ranges[len(ranges)-1]
		// if this element is adjacent to the high end of the last range
		if elems[i] == lastRange.end+1 {
			// then extend the last range to include this element
			lastRange.end = elems[i]
			continue
		}
		// otherwise, start a new range beginning with this element
		ranges = append(ranges, rng{elems[i], elems[i]})
	}

	// construct string from ranges
	var result bytes.Buffer
	for _, r := range ranges {
		if r.start == r.end {
			result.WriteString(strconv.Itoa(r.start))
		} else {
			result.WriteString(fmt.Sprintf("%d-%d", r.start, r.end))
		}
		result.WriteString(",")
	}

	return strings.TrimRight(result.String(), ",")
}

// Clone returns a copy of this CPU set.
func (s CPUSet) Clone() CPUSet {
	b := NewBuilder()
	for elem := range s.elems {
		b.Add(elem)
	}
	return b.Result()
}

func (s CPUSet) Contains(cpu int) bool {
	_, found := s.elems[cpu]
	return found
}

func (s CPUSet) Equals(s2 CPUSet) bool {
	return reflect.DeepEqual(s.elems, s2.elems)
}

// 求交集
func (s CPUSet) Intersection(s2 CPUSet) CPUSet {
	return s.Filter(func(cpu int) bool { return s2.Contains(cpu) })
}

// 求并集
func (s CPUSet) UnionAll(s2 []CPUSet) CPUSet {
	b := NewBuilder()
	for cpu := range s.elems {
		b.Add(cpu)
	}
	for _, cs := range s2 {
		for cpu := range cs.elems {
			b.Add(cpu)
		}
	}
	return b.Result()
}

func (s CPUSet) Union(s2 CPUSet) CPUSet {
	b := NewBuilder()
	for cpu := range s.elems {
		b.Add(cpu)
	}
	for cpu := range s2.elems {
		b.Add(cpu)
	}
	return b.Result()
}

// s - s2
func (s CPUSet) Difference(s2 CPUSet) CPUSet {
	return s.FilterNot(func(cpu int) bool { return s2.Contains(cpu) })
}

func (s CPUSet) FilterNot(predicate func(int) bool) CPUSet {
	b := NewBuilder()
	for cpu := range s.elems {
		if !predicate(cpu) {
			b.Add(cpu)
		}
	}
	return b.Result()
}

// builder 设计模式
func (s CPUSet) Filter(predicate func(int) bool) CPUSet {
	b := NewBuilder()
	for cpu := range s.elems {
		if predicate(cpu) {
			b.Add(cpu)
		}
	}
	return b.Result()
}

// ToSliceNoSort returns a slice of integers that contains all elements from
// this set.
func (s CPUSet) ToSliceNoSort() []int {
	result := []int{}
	for cpu := range s.elems {
		result = append(result, cpu)
	}
	return result
}

func NewCPUSet(cpus ...int) CPUSet {
	b := NewBuilder()
	for _, c := range cpus {
		b.Add(c)
	}
	return b.Result()
}

// Builder is a mutable builder for CPUSet. Functions that mutate instances
// of this type are not thread-safe.
type Builder struct {
	result CPUSet
	done   bool
}

// NewBuilder returns a mutable CPUSet builder.
func NewBuilder() Builder {
	return Builder{
		result: CPUSet{
			elems: map[int]struct{}{},
		},
	}
}
func (b Builder) Add(elems ...int) {
	if b.done {
		return
	}
	for _, elem := range elems {
		b.result.elems[elem] = struct{}{}
	}
}

func (b Builder) Result() CPUSet {
	b.done = true
	return b.result
}

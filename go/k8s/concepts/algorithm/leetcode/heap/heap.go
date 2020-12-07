package heap

import "sort"

type Interface interface {
	sort.Interface

	Push(x interface{})
	Pop() interface{}
}

func Init(h Interface) {

}

func Push(h Interface, x interface{}) {

}

func Pop(h Interface) interface{} {

	return nil
}

func Fix(h Interface, i int) {

}

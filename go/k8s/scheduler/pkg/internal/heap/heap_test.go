package heap

import "testing"

type testHeapObject struct {
	name string
	val  interface{}
}

func mkHeapObj(name string, val interface{}) testHeapObject {
	return testHeapObject{name: name, val: val}
}

func keyFunc(obj interface{}) (string, error) {
	return obj.(testHeapObject).name, nil
}
func compareFunc(val1 interface{}, val2 interface{}) bool {
	first := val1.(testHeapObject).val.(int)
	second := val2.(testHeapObject).val.(int)
	return first < second
}

func TestHeapBasic(test *testing.T) {
	h := New(keyFunc, compareFunc)

	const amount = 500
	var i int
	for i = amount; i > 0; i-- {
		h.Add(mkHeapObj(string([]rune{'a', rune(i)}), i))
	}

	// Make sure that the numbers are popped in ascending order.
	prevNum := 0
	for i := 0; i < amount; i++ {
		obj, err := h.Pop()
		num := obj.(testHeapObject).val.(int)
		// All the items must be sorted.
		if err != nil || prevNum > num {
			test.Errorf("got %v out of order, last was %v", obj, prevNum)
		}

		prevNum = num
	}
}

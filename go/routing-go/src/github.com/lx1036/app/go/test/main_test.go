package test

import "testing"

func TestFib(t *testing.T) {
	in  := 7
	expected := 13
	actual := Fib(in)

	if actual != expected{
		t.Errorf("Fib(%d) = %d; expected=%d", in, actual, expected)
	}
}

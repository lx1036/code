package main

import (
	"fmt"
	"math"
	"testing"
)

func Perimeter(rectangle Rectangle) float64 {
	return (rectangle.width + rectangle.height) * 2
}

/*func Area(rectangle Rectangle) float64 {
	return rectangle.width * rectangle.height
}*/

type Rectangle struct {
	width  float64
	height float64
}

func (rectangle Rectangle) Area() float64 {
	return rectangle.width * rectangle.height
}

func TestPerimeter(t *testing.T) {
	rectangle := Rectangle{10.0, 10.0}
	got := Perimeter(rectangle)
	want := 40.0
	if got != want {
		t.Errorf("got %.2f want %.2f", got, want) // .2 表示输出两位小数, f 表示 float64
	}
}

type Circle struct {
	radius float64
}

func (circle Circle) Area() float64 {
	return math.Pi * circle.radius * circle.radius
}

type Shape interface {
	Area() float64
}

type Triangle struct {
	base float64
	height float64
}

func (triangle Triangle) Area() float64  {
	return triangle.base * triangle.height * 0.5
}

func TestArea(t *testing.T) {
	tableDrivenTests := []struct{
		shape Shape
		want float64
	}{
		{shape: Rectangle{width: 10.0, height: 10.0}, want: 100.0},
		{Circle{10.0}, 314.1592653589793},
		{Triangle{12.0, 6}, 36.0},
	}

	for key, tableDrivenTest := range tableDrivenTests {
		fmt.Println(key)
		got := tableDrivenTest.shape.Area()
		if got != tableDrivenTest.want {
			t.Errorf("%#v got %.2f want %.2f", tableDrivenTest.shape, got, tableDrivenTest.want)
		}
	}
}

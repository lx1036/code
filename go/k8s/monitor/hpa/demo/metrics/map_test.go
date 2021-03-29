package main

import (
	"fmt"
	"testing"
)

func TestMap(test *testing.T) {
	metrics := map[string]int{
		"cpu":    1,
		"memory": 2,
	}
	add(metrics) // map 传的是指针
	fmt.Println(metrics)
	// map[cpu:2 memory:2]

	metricsSlice := []map[string]int{
		{
			"cpu":    1,
			"memory": 2,
		},
	}
	addSlice(metricsSlice) // slice 传的是指针
	fmt.Println(metricsSlice)
	// [map[cpu:2 memory:2]]

	a := 1
	addInt(a) // 标量scalar传的是值
	fmt.Println(a)
	// 1
}

func add(metrics map[string]int) {
	metrics["cpu"] += 1
}

func addSlice(metrics []map[string]int) {
	for _, metric := range metrics {
		metric["cpu"] += 1
	}
}

func addInt(a int) {
	a += 1
}

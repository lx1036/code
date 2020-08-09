package main

import (
	"flag"
	"fmt"
	"math"
)

var (
	min   = flag.Float64("min", 10, "min")
	max   = flag.Float64("max", 200, "max")
	count = flag.Int("count", 20, "count")
)

func main() {
	flag.Parse()
	init_bucket_range(*min, *max, *count)
}
func init_bucket_range(minVal float64, maxVal float64, bucket_count int) {
	var ranges map[int]float64
	ranges = make(map[int]float64)
	log_max := math.Log(maxVal)
	bucket_index := 1
	current := minVal
	ranges[bucket_index] = current
	run := true
	for run == true {
		bucket_index++
		if bucket_count < bucket_index {
			run = false
			continue
		}
		log_current := math.Log(current)
		last_count := bucket_count - bucket_index
		log_ratio := (log_max - log_current) / float64(last_count)
		log_next := log_current + log_ratio
		next := math.Floor(math.Exp(log_next) + 0.5)
		if next > current {
			current = next
		} else {
			current++
		}
		ranges[bucket_index] = current
	}
	for i := 1; i <= bucket_count; i++ {
		if bucket_count == i {
			fmt.Print(ranges[i])
		} else {
			fmt.Print(ranges[i], ",")
		}
	}
}

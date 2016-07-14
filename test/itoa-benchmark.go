package main

import (
	"time"
	"math/rand"
	"fmt"
	"strconv"
)

func main() {
	N := 10000000
	rand.Seed(time.Now().Unix())
	src :=make([]int, N, 2 * N)
	for i := 0; i < N; i++ {
		src[i] = rand.Int()
	}

	start := time.Now()
	for i := 0; i < N; i++ {
		fmt.Sprintf("%d", src[i])
	}
	elapse := time.Now().Sub(start).Nanoseconds()
	fmt.Printf("total time: %d\n", elapse)

	start = time.Now()
	for i := 0; i < N; i++ {
		strconv.Itoa(src[i])
	}
	elapse = time.Now().Sub(start).Nanoseconds()
	fmt.Printf("total time: %d\n", elapse)
}
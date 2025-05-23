package main

import (
	"fmt"
	"os"
	"strconv"
)

func add(a, b int) int {
	return a + b
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run add.go <int1> <int2>")
		return
	}

	a, err1 := strconv.Atoi(os.Args[1])
	b, err2 := strconv.Atoi(os.Args[2])

	if err1 != nil || err2 != nil {
		fmt.Println("Both arguments must be integers")
		return
	}

	sum := add(a, b)
	fmt.Println("Sum:", sum)
}

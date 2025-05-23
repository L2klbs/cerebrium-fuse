package main

import (
	"fmt"
	"os"
	"strconv"
)

func divide(a, b int) float64 {
	return float64(a) / float64(b)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run divide.go <int1> <int2>")
		return
	}

	a, err1 := strconv.Atoi(os.Args[1])
	b, err2 := strconv.Atoi(os.Args[2])

	if err1 != nil || err2 != nil {
		fmt.Println("Both arguments must be integers")
		return
	}

	if b == 0 {
		fmt.Println("Can't divide by 0")
		return
	}

	quotient := divide(a, b)
	fmt.Printf("Quotient: %.2f\n", quotient)
}

// Listing 3: declarations, assignments and type rules in Go.
// Run with: go run ./scratch/expressions
package main

import "fmt"

func main() {
	count, total := 2, 12.5 // := declares two variables and infers int and float64
	const maxItems = 100    // a constant
	a, b := 1, 2
	a, b = b, a                   // multiple assignment: swap
	avg := total / float64(count) // explicit conversion; total / count would not compile

	fmt.Println("count:", count, "total:", total, "maxItems:", maxItems)
	fmt.Println("after swap a:", a, "b:", b)
	fmt.Println("average:", avg)
}

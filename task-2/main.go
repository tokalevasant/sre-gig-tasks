package main

import (
	"fmt"
	"math/rand"
	"os"
)

func main() {
	// loop to test multiple random numbers
	for i := 0; i < 25; i++ {
		testRandomNumber(nextRandom())
	}
}

func testRandomNumber(i int) {
	fmt.Fprintln(os.Stdout, "Random number: ", i)

	if i > 50 {
		if i%2 == 0 {
			fmt.Fprintln(os.Stdout, "It's closer to 100, and it's even")
		} else {
			fmt.Fprintln(os.Stdout, "It's closer to 100")
		}
	} else if i == 50 {
		fmt.Fprintln(os.Stdout, "It's 50")
	} else {
		fmt.Fprintln(os.Stdout, "It's closer to 0")
	}
    fmt.Println()
}

func nextRandom() int {
	return rand.Intn(100)
}

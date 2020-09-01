package main

import "fmt"

func main() {
	var s = []int{1, 2, 3}
	s2 := append(s[:], 4)
	s2[0] = 100
	fmt.Println(s2)
	fmt.Println(s)
}

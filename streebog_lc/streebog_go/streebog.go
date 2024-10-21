package main

import (
	"C"
	"fmt"
)

//export helloWorld
func helloWorld() {
	fmt.Println("Hello, World!")
}

func main() {}

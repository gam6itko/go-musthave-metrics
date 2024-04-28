package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello World")
	os.Exit(1) // want "no need to exit"
	//goland:noinspection GoUnreachableCode
	fmt.Println("Bye World")
}

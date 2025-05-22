package example

import "fmt"

import "strings"

func ThisFunctionWillFailBecauseNoSeparator() {
	fmt.Println("Hello world")
	strings.HasPrefix("He	", "llo")
}

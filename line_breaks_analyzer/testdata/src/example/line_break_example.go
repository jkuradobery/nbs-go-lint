package example

import (
	"fmt"
	"math/rand"
	"os"
)

// DO NOT FORMAT THIS FILE WITH GOFMT
type Example struct {
	a int
}

// This is valid
func example() {
	file, err := os.Open("hello_world")
	if err != nil {
		panic(err)
	}
	defer func() {
		file.Close()
	}()

	data := make([]byte, 100)
	_, err = file.Read(data)
	if err != nil {
		panic(err)
	}

	innerFunc := func() {
		a := 24
		if rand.Int() > a {
			fmt.Println("Something")
		}

		switch rand.Int() % 3 {
		case 0:
			fmt.Println("Case 0")
		case 1:
			fmt.Println("Case 1")
		default:
			fmt.Println("Case 2")
		}

		for i := 0; i < 10; i++ {
			if i > 5 {
				fmt.Println("Something")
			}

			fmt.Println("Nothing")
		}
	}

	fmt.Println(Example{a: 24})
	fmt.Println(
		Example{a: 24})
	fmt.Println(
		Example{
			a: 24,
		},
	)
	innerFunc()
	_, err = file.Read(data)
	if err != nil {
		panic(err)
	}
}

// This is invalid
func brokenExample() {
	file, err := os.Open("hello_world")
	if err != nil {
		panic(err)
	}

	defer func() { // want "Line break before 'defer' statement is not allowed."
		err = file.Close()
		if err != nil {
			panic(err)
		}

	}() // want "Line break before closing } is not allowed."

	data := make([]byte, 100)
	_, err = file.Read(data)
	if err != nil {
		panic(err)
	}

	innerFunc := func() {
		a := 24
		if rand.Int() > a {
			fmt.Println("Something")
		} // want "Line break after closing } is required."
		switch rand.Int() % 3 {
		case 0:
			fmt.Println("Case 0")
		case 1:
			fmt.Println("Case 1")
		default:
			fmt.Println("Case 2")
		} // want "Line break after closing } is required."
		for i := 0; i < 10; i++ {
			if i > 5 {
				fmt.Println("Something")
			} // want "Line break after closing } is required."
			fmt.Println("Nothing")

		} // want "Line break before closing } is not allowed."
	} // want "Line break after closing } is required."
	innerFunc()
	_, err = file.Read(data)
	if err != nil {
		panic(err)
	}
}

// this is valid
func doBracketedStuff() {
	func() {
		fmt.Println("Hello world")
	}()
	fmt.Println("HW")
	b := func(aaaaaaaaaaaaaaa int, f func(int) int) int {
		fmt.Println(aaaaaaaaaaaaaaa * f(10))
		return 3
	}

	eax := func(
		ebx int,
		ecx uint,
		edx float32,
	) int {

		defer func() {}()
		return ebx * 2
	}

	eax(1, 3, 4.5)
	c := b(
		10,
		func(i int) int {
			return i * 2
		},
	)
	d := b(10, func(i int) int {
		return i * 2
	})
	fmt.Println(c * d)
}

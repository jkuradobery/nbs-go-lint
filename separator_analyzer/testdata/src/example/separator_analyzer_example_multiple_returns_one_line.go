package example

////////////////////////////////////////////////////////////////////////////////

func doSomething() (int, error) {
	return 0, nil
}

////////////////////////////////////////////////////////////////////////////////

type severance struct {
}

func SampleF() int { // want `Function which is not a constructor for struct 'SampleF' is not allowed in the same group as struct 'severance'`
	for i, j := range []int{1, 2, 3} {
		if i == j {
			return i
		}
	}

	i, k := doSomething()
	j, _ := doSomething()
	print(i, j, k)
	return 0
}

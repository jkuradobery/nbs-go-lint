package example

import "fmt"

////////////////////////////////////////////////////////////////////////////////

type Alpha interface {
	AlphaMethod()
}

////////////////////////////////////////////////////////////////////////////////

type Beta struct {
	Name string
}

func (b *Beta) AlphaMethod() {
	fmt.Printf("%s!\n", b.Name)
}

func NewBeta(name string) (Alpha, error) {
	return &Beta{Name: name}, nil
}

////////////////////////////////////////////////////////////////////////////////

type Dzeta struct {
	Name string
}

func (d *Dzeta) AlphaMethod() {
	fmt.Printf("Dzeta says hello, %s!\n", d.Name)
}

func NewDzeta(name string) Alpha {
	fst := &Dzeta{Name: name}
	return fst
}

////////////////////////////////////////////////////////////////////////////////

type Gamma struct {
	Name string
}

func (g *Gamma) AlphaMethod() {
	fmt.Printf("Gamma says hello, %s!\n", g.Name)
}

func NewGamma(name string) Alpha {
	var g *Gamma
	g = &Gamma{Name: name}
	return g
}

package goal

import (
	"fmt"

	"github.com/dvaumoron/foresee/base"
)

type guessed0[T0 any, T1 any] interface {
	Add(T0) T1
}

type adder[T base.Addable] struct {
	value T
}

func NewAdder[T base.Addable](a T) *adder[T] {
	return &adder[T]{value: a}
}

func (a *adder[T]) Add(b T) T {
	return a.value + b
}

type adderInt struct {
	value int
}

func MakeAdderInt(a int) adderInt {
	return adderInt{value: a}
}

func (a adderInt) Add(b int) int {
	return a.value + b
}

func Addition[T0 any, T1 any](a guessed0[T0, T1], b T0) T1 {
	return a.Add(b)
}

func UseAdder() {
	var a float64
	a = 1
	b := float64(2)

	fmt.Println("Result 1 is", Addition(NewAdder(a), b))

	c, d := 1, int(2)

	fmt.Println("Result 2 is", Addition(MakeAdderInt(c), d))
}

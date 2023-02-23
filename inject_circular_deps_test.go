package inject

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type A struct {
	b *B
}

type B struct {
	a *A
}

func newA(b *B) *A {
	return &A{b}
}

func newB(a *A) *B {
	return &B{a}
}

func TestCircularDependencies(t *testing.T) {
	mod := NewModule()
	mod.BindSingletonConstructor(newA)
	mod.BindSingletonConstructor(newB)
	inj, err := NewInjector(mod)
	require.Error(t, err)
	require.Nil(t, inj)

	fmt.Println(err)
}

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

func newAWithInjector(inj Injector) (*A, error) {
	b, err := inj.Call(newB)
	if err != nil {
		return nil, err
	}
	return &A{b[0].(*B)}, nil
}

// TestCircularDependencies tests that circular dependencies are detected at injector creation time.
func TestCircularDependencies(t *testing.T) {
	mod := NewModule()
	mod.BindSingletonConstructor(newA)
	mod.BindSingletonConstructor(newB)
	inj, err := NewInjector(mod)
	require.Error(t, err)
	require.Nil(t, inj)

	fmt.Println(err)
}

// TestRuntimeCircularDependencies tests that circular dependencies are detected at runtime when injecting the injector
// itself and getting values from it. Requires that detection is enabled.
func TestRuntimeCircularDependencies(t *testing.T) {
	tcd := DetectCircularDependencies
	DetectCircularDependencies = true
	defer func() { DetectCircularDependencies = tcd }()

	mod := NewModule()
	mod.BindSingletonConstructor(newAWithInjector)
	mod.BindSingletonConstructor(newB)
	inj, err := NewInjector(mod)
	require.NoError(t, err)

	a, err := inj.Get((*A)(nil))
	require.Error(t, err)
	require.Nil(t, a)

	fmt.Println(err)
}

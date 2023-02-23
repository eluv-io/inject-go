package inject_test

import (
	"fmt"

	"github.com/eluv-io/inject-go"
)

type A struct {
	b B
	e E
}

func newA(b B, e E) A {
	return A{b, e}
}

////////////////////////////////////////////////////////////////////////////////

type B struct {
	c C
	d D
}

func newB(c C, d D) B {
	return B{
		c: c,
		d: d,
	}
}

////////////////////////////////////////////////////////////////////////////////

type C struct{}
type D struct{}
type E struct{}

////////////////////////////////////////////////////////////////////////////////

func ExampleInjector_DependencyTree() {
	root := createInjector()
	child := createChildInjector(root)

	for _, inj := range []inject.Injector{root, child} {
		fmt.Println("injector name: ", inj.String())

		tree, _ := inj.DependencyTree()
		fmt.Println("dependency tree:")
		fmt.Println(tree)
	}

	// Output:
	//
	// injector name:  injector{my injector}
	// dependency tree:
	// root : my injector
	// ├── {type:*inject.Injector} : my injector
	// ├── {type:inject_test.A} : <github.com/eluv-io/inject-go_test.newA(inject_test.B, inject_test.E) inject_test.A>
	// │   ├── {type:inject_test.B} : <github.com/eluv-io/inject-go_test.newB(inject_test.C, inject_test.D) inject_test.B>
	// │   │   ├── {type:inject_test.C} : singleton inject_test.C
	// │   │   └── {type:inject_test.D} : singleton inject_test.D
	// │   └── {type:inject_test.E} : singleton inject_test.E
	// ├── {type:inject_test.B} : <github.com/eluv-io/inject-go_test.newB(inject_test.C, inject_test.D) inject_test.B>
	// │   ├── {type:inject_test.C} : singleton inject_test.C
	// │   └── {type:inject_test.D} : singleton inject_test.D
	// ├── {type:inject_test.C} : singleton inject_test.C
	// ├── {type:inject_test.D} : singleton inject_test.D
	// └── {type:inject_test.E} : singleton inject_test.E
	//
	// injector name:  injector{my child injector}, parent injector{my injector}
	// dependency tree:
	// child : my child injector
	// ├── {type:*inject.Injector} : my child injector
	// └── {type:string} : singleton string
}

func createInjector() inject.Injector {
	mod := inject.NewModule()
	mod.BindSingletonConstructor(newA)
	mod.BindSingletonConstructor(newB)
	mod.BindSingleton(C{})
	mod.BindSingleton(D{})
	mod.BindSingleton(E{})
	inj, _ := inject.NewNamedInjector("my injector", mod)
	return inj
}

func createChildInjector(inj inject.Injector) inject.Injector {
	mod := inject.NewModule()
	mod.BindSingleton("just a string")
	child, _ := inj.NewNamedChildInjector("my child injector", nil, mod)
	return child
}

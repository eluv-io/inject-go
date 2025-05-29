package inject_test

import (
	"fmt"
	"strings"

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

func (c *C) String() string {
	return "I'm C!"
}

type D struct{}
type E struct{}

type F struct {
	c C
}

func newF(c C) F {
	return F{c}
}

////////////////////////////////////////////////////////////////////////////////

func ExampleInjector_DependencyTree() {
	root := createInjector()
	child := createChildInjector(root)

	rootTree, _ := root.DependencyTree()
	childTree, _ := child.DependencyTree()

	if rootTree.String() != childTree.String() {
		fmt.Println("dependency tree mismatch!")
	}
	fmt.Println("dependency tree:")
	fmt.Println(rootTree)

	if rootTree.Dot() != childTree.Dot() {
		fmt.Println("dependency tree mismatch!")
	}
	fmt.Println("dependency graph:")
	fmt.Println(rootTree.Dot())

	// Output:
	//
	// dependency tree:
	// root-injector : app injector
	// ├── child-injector : scoped injector
	// │   ├── {type:*inject.Injector} : scoped injector
	// │   ├── {type:inject_test.F} : <github.com/eluv-io/inject-go_test.newF(inject_test.C) inject_test.F>
	// │   │   └── {type:inject_test.C} : singleton inject_test.C
	// │   └── {type:string} : singleton string
	// ├── {type:*inject.Injector} : app injector
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
	// dependency graph:
	// digraph G  {
	// 	"app injector" [tooltip="app injector"];
	// 	"inject-go_test.newA" [tooltip="<github.com/eluv-io/inject-go_test.newA(inject_test.B, inject_test.E) inject_test.A>"];
	// 	"inject-go_test.newB" [tooltip="<github.com/eluv-io/inject-go_test.newB(inject_test.C, inject_test.D) inject_test.B>"];
	// 	"inject-go_test.newF" [tooltip="<github.com/eluv-io/inject-go_test.newF(inject_test.C) inject_test.F>"];
	// 	"scoped injector" [tooltip="scoped injector"];
	// 	"singleton inject_test.C" [tooltip="singleton inject_test.C"];
	// 	"singleton inject_test.D" [tooltip="singleton inject_test.D"];
	// 	"singleton inject_test.E" [tooltip="singleton inject_test.E"];
	// 	"singleton string" [tooltip="singleton string"];
	//
	// 	"app injector" -> "scoped injector";
	// 	"app injector" -> "app injector";
	// 	"app injector" -> "inject-go_test.newA";
	// 	"app injector" -> "inject-go_test.newB";
	// 	"app injector" -> "singleton inject_test.C";
	// 	"app injector" -> "singleton inject_test.D";
	// 	"app injector" -> "singleton inject_test.E";
	// 	"inject-go_test.newA" -> "inject-go_test.newB";
	// 	"inject-go_test.newA" -> "singleton inject_test.E";
	// 	"inject-go_test.newB" -> "singleton inject_test.C";
	// 	"inject-go_test.newB" -> "singleton inject_test.D";
	// 	"inject-go_test.newF" -> "singleton inject_test.C";
	// 	"scoped injector" -> "scoped injector";
	// 	"scoped injector" -> "inject-go_test.newF";
	// 	"scoped injector" -> "singleton string";
	// }
}

func ExampleDependencyTree_Filter() {
	root := createInjector()
	_ = createChildInjector(root)

	rootTree, _ := root.DependencyTree()
	filteredTree := rootTree.Filter(func(bindingKey, resolvedBinding string) bool {
		return strings.Contains(bindingKey, "inject_test.D")
	})

	fmt.Println("filtered dependency tree:")
	fmt.Println(filteredTree.String())

	fmt.Println("filtered dependency graph:")
	fmt.Println(filteredTree.Dot("\tnode [shape=box,width=0, height=0,fixedsize=false]\n\tedge [arrowsize=0.5]\n"))

	// Output:
	//
	// filtered dependency tree:
	// root-injector : app injector
	// └── {type:inject_test.D} : singleton inject_test.D
	//
	// filtered dependency graph:
	// digraph G  {
	// 	node [shape=box,width=0, height=0,fixedsize=false]
	// 	edge [arrowsize=0.5]
	// 	"app injector" [tooltip="app injector"];
	// 	"singleton inject_test.D" [tooltip="singleton inject_test.D"];
	//
	// 	"app injector" -> "singleton inject_test.D";
	// }
}

func createInjector() inject.Injector {
	mod := inject.NewModule()
	mod.BindSingletonConstructor(newA)
	mod.BindSingletonConstructor(newB)
	mod.BindSingleton(C{})
	mod.BindSingleton(D{})
	mod.BindSingleton(E{})
	inj, _ := inject.NewNamedInjector("app injector", mod)
	return inj
}

func createChildInjector(inj inject.Injector) inject.Injector {
	mod := inject.NewModule()
	mod.BindSingleton("just a string")
	mod.BindSingletonConstructor(newF)
	child, _ := inj.NewNamedChildInjector("scoped injector", nil, mod)
	return child
}

func ExampleInjector_Obtain() {
	mod := inject.NewModule()
	mod.BindSingleton(&C{})
	mod.BindSingletonConstructor(func(c *C) fmt.Stringer { return c })
	injector, err := inject.NewInjector(mod)
	if err != nil {
		fmt.Println("failed to create injector", err)
		return
	}

	{
		var c *C
		err := injector.Obtain(&c)
		fmt.Println("no error", err == nil)
	}
	{
		var stringer fmt.Stringer
		err := injector.Obtain(&stringer)
		fmt.Println("no error", err == nil)
	}

	// Output:
	//
	// no error true
	// no error true
}

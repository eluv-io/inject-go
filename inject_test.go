package inject

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	goRoutineIterations = 100
)

// ***** simple bind tests *****

type SimpleInterface interface {
	Foo() string
}

type SimpleStruct struct {
	foo string
}

func (s SimpleStruct) Foo() string {
	return s.foo
}

func createSimpleInterface() SimpleInterface {
	return &SimpleStruct{foo: "default"}
}

type SimplePtrStruct struct {
	foo string
}

func (s *SimplePtrStruct) Foo() string {
	return s.foo
}

func createSimplePtrInterface() (SimpleInterface, error) {
	return &SimplePtrStruct{foo: "default"}, nil
}

func TestSimpleStructSingletonDirect(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil), SimpleStruct{}).ToSingleton(SimpleStruct{"hello"})

	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface := object.(SimpleInterface)
			require.Equal(t, "hello", simpleInterface.Foo())
			object, err = injector.Get(SimpleStruct{})
			require.NoError(t, err)
			simpleStruct := object.(SimpleStruct)
			require.Equal(t, "hello", simpleStruct.Foo())
		})
	}
}

func TestSimpleStructSingletonDirectSucceedsWhenPtr(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimpleStruct{"hello"})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface := object.(SimpleInterface)
			require.Equal(t, "hello", simpleInterface.Foo())
		})
	}
}

func TestSimplePtrStructSingletonDirect(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface := object.(SimpleInterface)
			require.Equal(t, "hello", simpleInterface.Foo())
		})
	}
}

func TestSimplePtrStructSingletonDirectFailsWhenNotPtr(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(SimplePtrStruct{"hello"})
	_, err := NewInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), injectErrorTypeNotAssignable)
}

func TestSimpleStructSingletonIndirect(t *testing.T) {
	module := NewModule()
	module.BindInterface((*SimpleInterface)(nil)).To(SimpleStruct{})
	module.Bind(SimpleStruct{}).ToSingleton(SimpleStruct{"hello"})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface := object.(SimpleInterface)
			require.Equal(t, "hello", simpleInterface.Foo())
		})
	}
}

func TestSimpleStructSingletonIndirectSucceedsWhenPtr(t *testing.T) {
	module := NewModule()
	module.BindInterface((*SimpleInterface)(nil)).To(&SimpleStruct{})
	module.Bind(&SimpleStruct{}).ToSingleton(&SimpleStruct{"hello"})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface := object.(SimpleInterface)
			require.Equal(t, "hello", simpleInterface.Foo())
		})
	}
}

func TestSimplePtrStructSingletonIndirect(t *testing.T) {
	module := NewModule()
	module.BindInterface((*SimpleInterface)(nil)).To(&SimplePtrStruct{})
	module.Bind(&SimplePtrStruct{}).ToSingleton(&SimplePtrStruct{"hello"})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface := object.(SimpleInterface)
			require.Equal(t, "hello", simpleInterface.Foo())
		})
	}
}

func TestSimplePtrStructSingletonIndirectFailsWhenNotPtr(t *testing.T) {
	module := NewModule()
	module.BindInterface((*SimpleInterface)(nil)).To(SimplePtrStruct{})
	_, err := NewInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), injectErrorTypeNotAssignable)
}

func TestSimplePtrStructIndirectFailsWhenNoFinalBinding(t *testing.T) {
	module := NewModule()
	module.BindInterface((*SimpleInterface)(nil)).To(&SimplePtrStruct{})
	_, err := NewInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), injectErrorTypeNoFinalBinding)
}

// ***** simple BindTagged tests *****

func TestTaggedTagEmpty(t *testing.T) {
	module := NewModule()
	module.BindTagged("", (*SimpleInterface)(nil)).ToSingleton(SimpleStruct{"hello"})
	_, err := NewInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), injectErrorTypeTagEmpty)
}

func TestTaggedSimpleStructSingletonDirect(t *testing.T) {
	module := NewModule()
	module.BindTagged("tagOne", (*SimpleInterface)(nil)).ToSingleton(SimpleStruct{"hello"})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.GetTagged("tagOne", (*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface := object.(SimpleInterface)
			require.Equal(t, "hello", simpleInterface.Foo())
			_, err = injector.Get((*SimpleInterface)(nil))
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)
			_, err = injector.GetTagged("tagTwo", (*SimpleInterface)(nil))
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)
		})
	}
}

func TestTaggedSimpleStructSingletonDirectSucceedsWhenPtr(t *testing.T) {
	module := NewModule()
	module.BindTagged("tagOne", (*SimpleInterface)(nil)).ToSingleton(&SimpleStruct{"hello"})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.GetTagged("tagOne", (*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface := object.(SimpleInterface)
			require.Equal(t, "hello", simpleInterface.Foo())
			_, err = injector.Get((*SimpleInterface)(nil))
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)
			_, err = injector.GetTagged("tagTwo", (*SimpleInterface)(nil))
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)
		})
	}
}

func TestTaggedSimplePtrStructSingletonDirect(t *testing.T) {
	module := NewModule()
	module.BindTagged("tagOne", (*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.GetTagged("tagOne", (*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface := object.(SimpleInterface)
			require.Equal(t, "hello", simpleInterface.Foo())
			_, err = injector.Get((*SimpleInterface)(nil))
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)
			_, err = injector.GetTagged("tagTwo", (*SimpleInterface)(nil))
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)
		})
	}
}

func TestTaggedSimplePtrStructSingletonDirectFailsWhenNotPtr(t *testing.T) {
	module := NewModule()
	module.BindTagged("tagOne", (*SimpleInterface)(nil)).ToSingleton(SimplePtrStruct{"hello"})
	_, err := NewInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), injectErrorTypeNotAssignable)
}

func TestTaggedSimpleStructSingletonIndirect(t *testing.T) {
	module := NewModule()
	module.BindTaggedInterface("tagOne", (*SimpleInterface)(nil)).To(SimpleStruct{})
	module.Bind(SimpleStruct{}).ToSingleton(SimpleStruct{"hello"})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.GetTagged("tagOne", (*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface := object.(SimpleInterface)
			require.Equal(t, "hello", simpleInterface.Foo())
			_, err = injector.Get((*SimpleInterface)(nil))
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)
			_, err = injector.GetTagged("tagTwo", (*SimpleInterface)(nil))
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)
		})
	}
}

func TestTaggedSimpleStructSingletonIndirectSucceedsWhenPtr(t *testing.T) {
	module := NewModule()
	module.BindTaggedInterface("tagOne", (*SimpleInterface)(nil)).To(&SimpleStruct{})
	module.Bind(&SimpleStruct{}).ToSingleton(&SimpleStruct{"hello"})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.GetTagged("tagOne", (*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface := object.(SimpleInterface)
			require.Equal(t, "hello", simpleInterface.Foo())
			_, err = injector.Get((*SimpleInterface)(nil))
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)
			_, err = injector.GetTagged("tagTwo", (*SimpleInterface)(nil))
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)
		})
	}
}

func TestTaggedSimplePtrStructSingletonIndirect(t *testing.T) {
	module := NewModule()
	module.BindTaggedInterface("tagOne", (*SimpleInterface)(nil)).To(&SimplePtrStruct{})
	module.Bind(&SimplePtrStruct{}).ToSingleton(&SimplePtrStruct{"hello"})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.GetTagged("tagOne", (*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface := object.(SimpleInterface)
			require.Equal(t, "hello", simpleInterface.Foo())
			_, err = injector.Get((*SimpleInterface)(nil))
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)
			_, err = injector.GetTagged("tagTwo", (*SimpleInterface)(nil))
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)
		})
	}
}

func TestTaggedSimplePtrStructSingletonIndirectFailsWhenNotPtr(t *testing.T) {
	module := NewModule()
	module.BindTaggedInterface("tagOne", (*SimpleInterface)(nil)).To(SimplePtrStruct{})
	_, err := NewInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), injectErrorTypeNotAssignable)
}

func TestTaggedSimplePtrStructIndirectFailsWhenNoFinalBinding(t *testing.T) {
	module := NewModule()
	module.BindTaggedInterface("tagOne", (*SimpleInterface)(nil)).To(&SimplePtrStruct{})
	_, err := NewInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), injectErrorTypeNoFinalBinding)
}

// ***** additional simple tagged tests *****

func TestTaggedSimpleStructSingletonDirectTwoBindings(t *testing.T) {
	module := NewModule()
	module.BindTagged("tagOne", (*SimpleInterface)(nil)).ToSingleton(SimpleStruct{"hello"})
	module.BindTagged("tagTwo", (*SimpleInterface)(nil)).ToSingleton(SimpleStruct{"good day"})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.GetTagged("tagOne", (*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface := object.(SimpleInterface)
			require.Equal(t, "hello", simpleInterface.Foo())
			object, err = injector.GetTagged("tagTwo", (*SimpleInterface)(nil))
			require.NoError(t, err)
			simpleInterface = object.(SimpleInterface)
			require.Equal(t, "good day", simpleInterface.Foo())
		})
	}
}

// ***** simple constructor tests *****

type BarInterface interface {
	Bar() int
}

type BarStruct struct {
	bar int
}

func (b BarStruct) Bar() int {
	return b.bar
}

type BarPtrStruct struct {
	bar int
}

func (b *BarPtrStruct) Bar() int {
	return b.bar
}

type SecondInterface interface {
	Foo() SimpleInterface
	Bar() BarInterface
}

type SecondPtrStruct struct {
	foo SimpleInterface
	bar BarInterface
}

func (s *SecondPtrStruct) Foo() SimpleInterface {
	return s.foo
}

func (s *SecondPtrStruct) Bar() BarInterface {
	return s.bar
}

type UnboundInterface interface {
	Baz() string
}

func createSecondInterface(s SimpleInterface, b BarInterface) (SecondInterface, error) {
	return &SecondPtrStruct{s, b}, nil
}

func createSecondInterfaceErr(s SimpleInterface, b BarInterface) (SecondInterface, error) {
	return nil, errors.New("XYZ")
}

func createSecondInterfaceErrNoBinding(s SimpleInterface, b BarInterface, u UnboundInterface) (SecondInterface, error) {
	return &SecondPtrStruct{s, b}, nil
}

type BarInterfaceError struct {
	BarInterface
	err error
}

var evilCounter int32

func createEvilBarInterface() (BarInterface, error) {
	value := atomic.AddInt32(&evilCounter, 1)
	return &BarPtrStruct{int(value)}, nil
}

func createEvilBarInterfaceErr() (BarInterface, error) {
	value := atomic.AddInt32(&evilCounter, 1)
	return nil, fmt.Errorf("XYZ %v", value)
}

func TestMultipleBindingErrors(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(SimplePtrStruct{"hello"})
	module.Bind((*BarInterface)(nil)).ToSingleton(BarPtrStruct{1})
	_, err := NewInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), "1:inject:")
	require.Contains(t, err.Error(), "2:inject:")
	require.Contains(t, err.Error(), injectErrorTypeNotAssignable)
}

func TestConstructorSimple(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	module.Bind((*SecondInterface)(nil)).ToConstructor(createSecondInterface)
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SecondInterface)(nil))
			require.NoError(t, err)
			secondInterface := object.(SecondInterface)
			require.Equal(t, "hello", secondInterface.Foo().Foo())
			require.Equal(t, 1, secondInterface.Bar().Bar())
		})
	}
}

func TestConstructorErrReturned(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	module.Bind((*SecondInterface)(nil)).ToConstructor(createSecondInterfaceErr)
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			_, err := injector.Get((*SecondInterface)(nil))
			require.Error(t, err)
			require.Contains(t, err.Error(), "XYZ")
		})
	}
}

func TestConstructorErrNoBindingReturned(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	module.Bind((*SecondInterface)(nil)).ToConstructor(createSecondInterfaceErrNoBinding)
	_, err := NewInjector(module)
	require.Error(t, err)
	fmt.Println(err)
	require.Contains(t, err.Error(), injectErrorTypeNoBinding)
	require.Contains(t, err.Error(), "inject.UnboundInterface")
}

func TestConstructorWithEvilCounter(t *testing.T) {
	module := NewModule()
	module.Bind((*BarInterface)(nil)).ToConstructor(createEvilBarInterface)
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {

			evilCounter = int32(0)
			evilChan := make(chan BarInterfaceError)

			// TODO(pedge): i know this is a terrible way to do concurrency testing
			for i := 0; i < goRoutineIterations; i++ {
				go func() {
					barInterface, err := injector.Get((*BarInterface)(nil))
					if err != nil {
						evilChan <- BarInterfaceError{nil, err}
					} else {
						evilChan <- BarInterfaceError{barInterface.(BarInterface), nil}
					}
				}()
			}
			count := 0
			for i := 0; i < goRoutineIterations; i++ {
				barInterfaceErr := <-evilChan
				require.Nil(t, barInterfaceErr.err, "%v", barInterfaceErr.err)
				bar := barInterfaceErr.Bar()
				count += bar
			}
			// cute
			require.Equal(t, (goRoutineIterations*(goRoutineIterations+1))/2, count)

			close(evilChan)
		})
	}
}

func TestSingletonConstructorWithEvilCounter(t *testing.T) {
	module := NewModule()
	module.Bind((*BarInterface)(nil)).ToSingletonConstructor(createEvilBarInterface)
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {

			evilCounter = int32(0)
			evilChan := make(chan BarInterfaceError)

			// TODO(pedge): i know this is a terrible way to do concurrency testing
			for i := 0; i < goRoutineIterations; i++ {
				go func() {
					barInterface, err := injector.Get((*BarInterface)(nil))
					if err != nil {
						evilChan <- BarInterfaceError{nil, err}
					} else {
						evilChan <- BarInterfaceError{barInterface.(BarInterface), nil}
					}
				}()
			}
			for i := 0; i < goRoutineIterations; i++ {
				barInterfaceErr := <-evilChan
				require.Nil(t, barInterfaceErr.err, "%v", barInterfaceErr.err)
				require.Equal(t, 1, barInterfaceErr.Bar())
			}

			close(evilChan)
		})
	}
}

func TestSingletonConstructorWithEvilCounterErr(t *testing.T) {
	module := NewModule()
	module.Bind((*BarInterface)(nil)).ToSingletonConstructor(createEvilBarInterfaceErr)
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {

			evilCounter = int32(0)
			evilChan := make(chan BarInterfaceError)

			// TODO(pedge): i know this is a terrible way to do concurrency testing
			for i := 0; i < goRoutineIterations; i++ {
				go func() {
					barInterface, err := injector.Get((*BarInterface)(nil))
					if err != nil {
						evilChan <- BarInterfaceError{nil, err}
					} else {
						evilChan <- BarInterfaceError{barInterface.(BarInterface), nil}
					}
				}()
			}
			for i := 0; i < goRoutineIterations; i++ {
				barInterfaceErr := <-evilChan
				require.Error(t, barInterfaceErr.err)
				require.Contains(t, barInterfaceErr.err.Error(), "XYZ 1")
				if i == 0 {
					fmt.Println(barInterfaceErr.err)
				}
			}

			close(evilChan)
		})
	}
}

func TestSingletonConstructorWithEvilCounterMultipleInjectors(t *testing.T) {
	module := NewModule()
	module.Bind((*BarInterface)(nil)).ToSingletonConstructor(createEvilBarInterface)
	injector1, err := NewInjector(module)
	require.NoError(t, err)
	injector2, err := NewInjector(module)
	require.NoError(t, err)
	injector3, err := NewInjector(module)
	require.NoError(t, err)

	evilCounter = int32(0)

	barInterface1, err := injector1.Get((*BarInterface)(nil))
	require.NoError(t, err)
	barInterface2, err := injector2.Get((*BarInterface)(nil))
	require.NoError(t, err)
	barInterface3, err := injector3.Get((*BarInterface)(nil))
	require.NoError(t, err)
	require.Equal(t, 1, barInterface1.(BarInterface).Bar())
	require.Equal(t, 2, barInterface2.(BarInterface).Bar())
	require.Equal(t, 3, barInterface3.(BarInterface).Bar())
	barInterface1, err = injector1.Get((*BarInterface)(nil))
	require.NoError(t, err)
	barInterface2, err = injector2.Get((*BarInterface)(nil))
	require.NoError(t, err)
	barInterface3, err = injector3.Get((*BarInterface)(nil))
	require.NoError(t, err)
	require.Equal(t, 1, barInterface1.(BarInterface).Bar())
	require.Equal(t, 2, barInterface2.(BarInterface).Bar())
	require.Equal(t, 3, barInterface3.(BarInterface).Bar())
}

// ***** tagged constructors

func createSecondInterfaceTaggedOne(str struct {
	S SimpleInterface `inject:"tagOne"`
	B BarInterface    `inject:"tagTwo"`
}) (SecondInterface, error) {
	return &SecondPtrStruct{str.S, str.B}, nil
}

func createSecondInterfaceTaggedOneHasNoTag(str struct {
	S SimpleInterface `inject:"tagOne"`
	B BarInterface
}) (SecondInterface, error) {
	return &SecondPtrStruct{str.S, str.B}, nil
}

func createSecondInterfaceTaggedNoTags(str struct {
	S SimpleInterface
	B BarInterface
}) (SecondInterface, error) {
	return &SecondPtrStruct{str.S, str.B}, nil
}

func TestTaggedConstructorSimple(t *testing.T) {
	module := NewModule()
	module.BindTagged("tagOne", (*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	module.BindTagged("tagTwo", (*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	module.Bind((*SecondInterface)(nil)).ToTaggedConstructor(createSecondInterfaceTaggedOne)
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SecondInterface)(nil))
			require.NoError(t, err)
			secondInterface := object.(SecondInterface)
			require.Equal(t, "hello", secondInterface.Foo().Foo())
			require.Equal(t, 1, secondInterface.Bar().Bar())
		})
	}
}

func TestTaggedConstructorOneHasNoTag(t *testing.T) {
	module := NewModule()
	module.BindTagged("tagOne", (*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	module.Bind((*SecondInterface)(nil)).ToTaggedConstructor(createSecondInterfaceTaggedOneHasNoTag)
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SecondInterface)(nil))
			require.NoError(t, err)
			secondInterface := object.(SecondInterface)
			require.Equal(t, "hello", secondInterface.Foo().Foo())
			require.Equal(t, 1, secondInterface.Bar().Bar())
		})
	}
}

func TestTaggedConstructorNoTags(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	module.Bind((*SecondInterface)(nil)).ToTaggedConstructor(createSecondInterfaceTaggedNoTags)
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SecondInterface)(nil))
			require.NoError(t, err)
			secondInterface := object.(SecondInterface)
			require.Equal(t, "hello", secondInterface.Foo().Foo())
			require.Equal(t, 1, secondInterface.Bar().Bar())
		})
	}
}

func TestTaggedConstructorOneHasNoTagMultipleBindings(t *testing.T) {
	module := NewModule()
	module.BindTagged("tagOne", (*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	module.BindTagged("tagTwo", (*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"goodbye"})
	module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"another"})
	module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	module.Bind((*SecondInterface)(nil)).ToTaggedConstructor(createSecondInterfaceTaggedOneHasNoTag)
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SecondInterface)(nil))
			require.NoError(t, err)
			secondInterface := object.(SecondInterface)
			require.Equal(t, "hello", secondInterface.Foo().Foo())
			require.Equal(t, 1, secondInterface.Bar().Bar())
			object, err = injector.GetTagged("tagTwo", (*SimpleInterface)(nil))
			require.NoError(t, err)
			require.Equal(t, "goodbye", object.(SimpleInterface).Foo())
			object, err = injector.Get((*SimpleInterface)(nil))
			require.NoError(t, err)
			require.Equal(t, "another", object.(SimpleInterface).Foo())
		})
	}
}

func newModuleForTheTestBelow(t *testing.T) Module {
	module := NewModule()
	module.BindTagged("tagOne", (*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	module.BindTagged("tagTwo", (*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"goodbye"})
	module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"another"})
	module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	module.Bind((*SecondInterface)(nil)).ToTaggedConstructor(createSecondInterfaceTaggedOneHasNoTag)
	return module
}

// had a situation where I thought I might have a pointer issue, keeping this test anyways for now
func TestTaggedConstructorOneHasNoTagMultipleBindingsModuleFromFunction(t *testing.T) {
	for _, injector := range createInjectors(t, newModuleForTheTestBelow(t)) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SecondInterface)(nil))
			require.NoError(t, err)
			secondInterface := object.(SecondInterface)
			require.Equal(t, "hello", secondInterface.Foo().Foo())
			require.Equal(t, 1, secondInterface.Bar().Bar())
			object, err = injector.GetTagged("tagTwo", (*SimpleInterface)(nil))
			require.NoError(t, err)
			require.Equal(t, "goodbye", object.(SimpleInterface).Foo())
			object, err = injector.Get((*SimpleInterface)(nil))
			require.NoError(t, err)
			require.Equal(t, "another", object.(SimpleInterface).Foo())
		})
	}
}

// ***** Call, CallTagged tests *****

func getSecondInterface(s SimpleInterface, b BarInterface) (SecondInterface, string, error) {
	return &SecondPtrStruct{s, b}, "hello", nil
}

func getSecondInterfaceErrNoBinding(s SimpleInterface, b BarInterface, u UnboundInterface) (SecondInterface, string, error) {
	return &SecondPtrStruct{s, b}, "hello", nil
}

func getSecondInterfaceTaggedOne(str struct {
	S SimpleInterface `inject:"tagOne"`
	B BarInterface    `inject:"tagTwo"`
}) (SecondInterface, string, error) {
	return &SecondPtrStruct{str.S, str.B}, "hello", nil
}

func getSecondInterfaceTaggedOneHasNoTag(str struct {
	S SimpleInterface `inject:"tagOne"`
	B BarInterface
}) (SecondInterface, string, error) {
	return &SecondPtrStruct{str.S, str.B}, "hello", nil
}

func getSecondInterfaceTaggedNoTags(str struct {
	S SimpleInterface
	B BarInterface
}) (SecondInterface, string, error) {
	return &SecondPtrStruct{str.S, str.B}, "hello", nil
}

func getSecondInterfaceTaggedNoTagsAndError(str struct {
	S SimpleInterface
	B BarInterface
}) (SecondInterface, string, error) {
	return &SecondPtrStruct{str.S, str.B}, "hello", errors.New("an error")
}

func TestCallAndCallTaggedSimple(t *testing.T) {
	module := NewModule()
	module.BindTagged("tagOne", (*SimpleInterface)(nil)).ToSingleton(SimpleStruct{"hello"})
	module.Bind((*SimpleInterface)(nil)).ToSingleton(SimpleStruct{"another"})
	module.BindTagged("tagTwo", (*BarInterface)(nil)).ToSingleton(BarStruct{1})
	module.Bind((*BarInterface)(nil)).ToSingleton(BarStruct{2})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {

			values, err := injector.Call(getSecondInterface)
			require.NoError(t, err)
			secondPtrStruct := values[0].(*SecondPtrStruct)
			str := values[1].(string)
			require.Equal(t, SecondPtrStruct{SimpleStruct{"another"}, BarStruct{2}}, *secondPtrStruct)
			require.Equal(t, "hello", str)
			require.Nil(t, values[2])

			values, err = injector.Call(getSecondInterfaceErrNoBinding)
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)

			values, err = injector.CallTagged(getSecondInterfaceTaggedOne)
			require.NoError(t, err)
			secondPtrStruct = values[0].(*SecondPtrStruct)
			str = values[1].(string)
			require.Equal(t, SecondPtrStruct{SimpleStruct{"hello"}, BarStruct{1}}, *secondPtrStruct)
			require.Equal(t, "hello", str)
			require.Nil(t, values[2])

			values, err = injector.CallTagged(getSecondInterfaceTaggedOneHasNoTag)
			require.NoError(t, err)
			secondPtrStruct = values[0].(*SecondPtrStruct)
			str = values[1].(string)
			require.Equal(t, SecondPtrStruct{SimpleStruct{"hello"}, BarStruct{2}}, *secondPtrStruct)
			require.Equal(t, "hello", str)
			require.Nil(t, values[2])

			values, err = injector.CallTagged(getSecondInterfaceTaggedNoTags)
			require.NoError(t, err)
			secondPtrStruct = values[0].(*SecondPtrStruct)
			str = values[1].(string)
			require.Equal(t, SecondPtrStruct{SimpleStruct{"another"}, BarStruct{2}}, *secondPtrStruct)
			require.Equal(t, "hello", str)
			require.Nil(t, values[2])

			values, err = injector.CallTagged(getSecondInterfaceTaggedNoTagsAndError)
			require.NoError(t, err)
			secondPtrStruct = values[0].(*SecondPtrStruct)
			str = values[1].(string)
			require.Equal(t, SecondPtrStruct{SimpleStruct{"another"}, BarStruct{2}}, *secondPtrStruct)
			require.Equal(t, "hello", str)
			err = values[2].(error)
			require.Equal(t, errors.New("an error"), err)
		})
	}
}

// ***** Populate tests *****

type PopulateStructTwoTags struct {
	S SimpleInterface `inject:"tagOne"`
	B BarInterface    `inject:"tagTwo"`
}

type PopulateStructNoBinding struct {
	S SimpleInterface `inject:"tagOne"`
	B BarInterface    `inject:"tagTwo"`
	U UnboundInterface
}

type PopulateStructOneTag struct {
	S SimpleInterface `inject:"tagOne"`
	B BarInterface
}

type PopulateStructNoTags struct {
	S SimpleInterface
	B BarInterface
}

func TestPopulateSimple(t *testing.T) {
	module := NewModule()
	module.BindTagged("tagOne", (*SimpleInterface)(nil)).ToSingleton(SimpleStruct{"hello"})
	module.Bind((*SimpleInterface)(nil)).ToSingleton(SimpleStruct{"another"})
	module.BindTagged("tagTwo", (*BarInterface)(nil)).ToSingleton(BarStruct{1})
	module.Bind((*BarInterface)(nil)).ToSingleton(BarStruct{2})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {

			populateStructTwoTags := PopulateStructTwoTags{}
			err := injector.Populate(&populateStructTwoTags)
			require.NoError(t, err)
			require.Equal(t, PopulateStructTwoTags{SimpleStruct{"hello"}, BarStruct{1}}, populateStructTwoTags)

			populateStructNoBinding := PopulateStructNoBinding{}
			err = injector.Populate(&populateStructNoBinding)
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNoBinding)

			populateStructOneTag := PopulateStructOneTag{}
			err = injector.Populate(&populateStructOneTag)
			require.NoError(t, err)
			require.Equal(t, PopulateStructOneTag{SimpleStruct{"hello"}, BarStruct{2}}, populateStructOneTag)

			populateStructNoTags := PopulateStructNoTags{}
			err = injector.Populate(&populateStructNoTags)
			require.NoError(t, err)
			require.Equal(t, PopulateStructNoTags{SimpleStruct{"another"}, BarStruct{2}}, populateStructNoTags)
		})
	}
}

// ***** BindTaggedConstant tests *****

type PopulateStructOneTagWithInt struct {
	S SimpleInterface `inject:"tagOne"`
	B BarInterface
	I int `inject:"intTag"`
}

type PopulateStructOneTagWithIntErr struct {
	S SimpleInterface `inject:"tagOne"`
	B BarInterface
	I *int
}

func TestBindTaggedConstantSimple(t *testing.T) {
	module := NewModule()
	module.BindTaggedBool("boolTrue").ToSingleton(true)
	module.BindTaggedBool("boolFalse").ToSingleton(false)
	module.BindTaggedInt("int10").ToConstructor(func() (int, error) { return 10, nil })
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			boolTrue, err := injector.GetTaggedBool("boolTrue")
			require.NoError(t, err)
			require.Equal(t, true, boolTrue)
			boolFalse, err := injector.GetTaggedBool("boolFalse")
			require.NoError(t, err)
			require.Equal(t, false, boolFalse)
		})
	}
}

func TestBindTaggedConstantWrongType(t *testing.T) {
	module := NewModule()
	module.BindTaggedInt8("tagOne").ToSingleton(1)
	_, err := NewInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), injectErrorTypeNotAssignable)
}

func TestBindTaggedConstantPopulate(t *testing.T) {
	module := NewModule()
	module.BindTagged("tagOne", (*SimpleInterface)(nil)).ToSingleton(SimpleStruct{"hello"})
	module.Bind((*BarInterface)(nil)).ToSingleton(BarStruct{2})
	module.BindTaggedInt("intTag").ToSingleton(10)
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			populateStructOneTagWithInt := PopulateStructOneTagWithInt{}
			err := injector.Populate(&populateStructOneTagWithInt)
			require.NoError(t, err)
			require.Equal(t, PopulateStructOneTagWithInt{SimpleStruct{"hello"}, BarStruct{2}, 10}, populateStructOneTagWithInt)

			populateStructOneTagWithIntErr := PopulateStructOneTagWithIntErr{}
			err = injector.Populate(&populateStructOneTagWithIntErr)
			require.Error(t, err)
			require.Contains(t, err.Error(), injectErrorTypeNotSupportedYet)
		})
	}
}

func newSimpleStructFromTaggedStringAnonymousStruct(p struct {
	Time time.Time
	Val  string `inject:"tag_one"`
}) *SimpleStruct {
	return &SimpleStruct{foo: fmt.Sprintf("%s-%s", p.Val, p.Time)}
}

func TestInjectTaggedStringInConstructorWithAnonymousStruct(t *testing.T) {
	module := NewModule()
	now := time.Now()
	module.Bind(time.Time{}).ToSingleton(now)
	module.BindTaggedString("tag_one").ToSingleton("Value of tag_one")
	module.Bind((*SimpleStruct)(nil)).ToTaggedSingletonConstructor(newSimpleStructFromTaggedStringAnonymousStruct)
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			v, err := injector.GetTaggedString("tag_one")
			require.NoError(t, err)
			require.Equal(t, v, "Value of tag_one")

			s, err := injector.Get((*SimpleStruct)(nil))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprintf("%s-%s", "Value of tag_one", now), s.(*SimpleStruct).foo)
		})
	}
}

func TestOverride(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	module.Bind((*SecondInterface)(nil)).ToConstructor(createSecondInterface)

	override1 := NewModule()
	override1.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"yo, yo!"})

	override2 := NewModule()
	override2.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{3})

	for _, injector := range createInjectors(t, Override(module).With(override1, override2)) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SecondInterface)(nil))
			require.NoError(t, err)
			secondInterface := object.(SecondInterface)
			require.Equal(t, "yo, yo!", secondInterface.Foo().Foo())
			require.Equal(t, 3, secondInterface.Bar().Bar())
		})
	}
}

func TestOverrideErrorsInSource(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(SimplePtrStruct{"hello"})
	module.Bind((*BarInterface)(nil)).ToSingleton(BarPtrStruct{1})

	override := NewModule()
	override.Bind((*SimpleInterface)(nil)).ToSingleton(SimpleStruct{"hello"})
	override.Bind((*BarInterface)(nil)).ToSingleton(BarStruct{1})

	_, err := NewInjector(Override(module).With(override))
	require.Error(t, err)
	require.Contains(t, err.Error(), "1:inject:")
	require.Contains(t, err.Error(), "2:inject:")
	require.Contains(t, err.Error(), injectErrorTypeNotAssignable)
}

func TestOverrideErrorsInOverride(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(SimpleStruct{"hello"})
	module.Bind((*BarInterface)(nil)).ToSingleton(BarStruct{1})

	override := NewModule()
	override.Bind((*SimpleInterface)(nil)).ToSingleton(SimplePtrStruct{"hello"})
	override.Bind((*BarInterface)(nil)).ToSingleton(BarPtrStruct{1})

	_, err := NewInjector(Override(module).With(override))
	require.Error(t, err)
	require.Contains(t, err.Error(), "1:inject:")
	require.Contains(t, err.Error(), "2:inject:")
	require.Contains(t, err.Error(), injectErrorTypeNotAssignable)
}

func TestOverrideErrorsInBoth(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(SimpleStruct{"hello"})
	module.Bind((*BarInterface)(nil)).ToSingleton(BarPtrStruct{1})

	override := NewModule()
	override.Bind((*BarInterface)(nil)).ToSingleton(BarStruct{1})
	override.Bind((*SimpleInterface)(nil)).ToSingleton(SimplePtrStruct{"hello"})

	_, err := NewInjector(Override(module).With(override))
	require.Error(t, err)
	require.Contains(t, err.Error(), "1:inject:")
	require.Contains(t, err.Error(), "2:inject:")
	require.Contains(t, err.Error(), injectErrorTypeNotAssignable)
}

func TestInstall(t *testing.T) {
	sub1 := NewModule()
	sub1.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})

	sub2 := NewModule()
	sub2.Bind((*SecondInterface)(nil)).ToConstructor(createSecondInterface)

	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	module.Install(sub1)
	module.Install(sub2)

	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*SecondInterface)(nil))
			require.NoError(t, err)
			secondInterface := object.(SecondInterface)
			require.Equal(t, "hello", secondInterface.Foo().Foo())
			require.Equal(t, 1, secondInterface.Bar().Bar())
		})
	}
}

func TestInstallDuplicateBinding(t *testing.T) {
	sub1 := NewModule()
	sub1.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"}) // duplicate binding

	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})

	module.Install(sub1)

	_, err := NewInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), "1:inject:")
	require.Contains(t, err.Error(), injectErrorTypeAlreadyBound)
}

func TestInstallErrorsInSource(t *testing.T) {
	sub1 := NewModule()
	sub1.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	sub1.Bind((*SecondInterface)(nil)).ToConstructor(createSecondInterface)

	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(SimplePtrStruct{"hello"}) // invalid binding

	module.Install(sub1)

	_, err := NewInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), "1:inject:")
	require.Contains(t, err.Error(), injectErrorTypeNotAssignable)
}

func TestInstallErrorsInTarget(t *testing.T) {
	sub := NewModule()
	sub.Bind((*BarInterface)(nil)).ToSingleton(BarPtrStruct{1}) // invalid binding
	sub.Bind((*SecondInterface)(nil)).ToConstructor(createSecondInterface)

	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})

	module.Install(sub)

	_, err := NewInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), "1:inject:")
	require.Contains(t, err.Error(), injectErrorTypeNotAssignable)
}

func TestInstallErrorsInBoth(t *testing.T) {
	sub := NewModule()
	sub.Bind((*BarInterface)(nil)).ToSingleton(BarPtrStruct{1}) // invalid binding
	sub.Bind((*SecondInterface)(nil)).ToConstructor(createSecondInterface)

	module := NewModule()
	module.Bind((*SimpleInterface)(nil)).ToSingleton(SimplePtrStruct{"hello"}) // invalid binding

	module.Install(sub)

	_, err := NewInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), "1:inject:")
	require.Contains(t, err.Error(), "2:inject:")
	require.Contains(t, err.Error(), injectErrorTypeNotAssignable)
}

func callWithSecondInterface(s SecondInterface) SecondInterface {
	return s
}

func TestBindConstructor(t *testing.T) {
	tests := []struct {
		name        string
		constructor interface{}
		singleton   bool
	}{
		{"BindAutoWithSimple", createSimpleInterface, false},
		{"BindAutoSingletonWithSimple", createSimpleInterface, true},
		{"BindAutoWithSimplePrt", createSimplePtrInterface, false},
		{"BindAutoSingletonWithSimplePtr", createSimplePtrInterface, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doBindConstructor(t, tt.constructor, tt.singleton)
		})
	}
}

func doBindConstructor(t *testing.T, constructor interface{}, singleton bool) {
	module := NewModule()
	if singleton {
		module.BindSingletonConstructor(constructor)
	} else {
		module.BindConstructor(constructor)
	}
	module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	module.BindConstructor(createSecondInterface)

	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			res, err := injector.Call(callWithSecondInterface)
			require.NoError(t, err)
			require.Equal(t, 1, len(res))
			require.NotNil(t, res[0])
			secondInterfaceFromCall, ok := res[0].(SecondInterface)
			require.True(t, ok)

			object, err := injector.Get((*SecondInterface)(nil))
			require.NoError(t, err)
			secondInterface := object.(SecondInterface)
			require.Equal(t, "default", secondInterface.Foo().Foo())
			require.Equal(t, 1, secondInterface.Bar().Bar())

			require.Exactly(t, object, res[0])
			require.Equal(t, singleton, secondInterface.Foo() == secondInterfaceFromCall.Foo())
		})
	}
}

func TestGetInjector(t *testing.T) {
	module := NewModule()
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			object, err := injector.Get((*Injector)(nil))
			require.NoError(t, err)
			i := object.(Injector)
			require.Exactly(t, injector.Injector, i)
		})
	}
}

func callWithInjector(injector Injector) Injector {
	return injector
}

func TestInjectInjector(t *testing.T) {
	module := NewModule()
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			res, err := injector.Call(callWithInjector)
			require.NoError(t, err)
			require.Equal(t, 1, len(res))
			i := res[0].(Injector)
			require.Exactly(t, injector.Injector, i)
		})
	}
}

var simpleInterfaceCreateCount = 0

func createSimpleInterfaceAndCount() SimpleInterface {
	simpleInterfaceCreateCount++
	return &SimpleStruct{foo: "default"}
}

var callAndIncrementCount = 0
var callAndIncrementSimple SimpleInterface

func callAndIncrement(s SimpleInterface) {
	callAndIncrementCount++
	callAndIncrementSimple = s
}

type moduleType int

const (
	regular moduleType = iota
	installed
	overridden
)

func TestEagerSingletons(t *testing.T) {
	tests := []struct {
		name       string
		modCreator func() Module
		eagerly    bool
	}{
		{"BindSingletonConstructor.Eagerly", func() Module {
			module := NewModule()
			singletonBuilder := module.BindSingletonConstructor(createSimpleInterfaceAndCount)
			singletonBuilder.Eagerly()
			return module
		}, true},
		{"BindSingletonConstructor.EagerlyAndCall", func() Module {
			module := NewModule()
			singletonBuilder := module.BindSingletonConstructor(createSimpleInterfaceAndCount)
			singletonBuilder.EagerlyAndCall(callAndIncrement)
			return module
		}, false},
		{"Bind.ToSingletonConstructor.Eagerly", func() Module {
			module := NewModule()
			singletonBuilder := module.Bind((*SimpleInterface)(nil)).ToSingletonConstructor(createSimpleInterfaceAndCount)
			singletonBuilder.Eagerly()
			return module
		}, true},
		{"Bind.ToSingletonConstructor.EagerlyAndCall", func() Module {
			module := NewModule()
			singletonBuilder := module.Bind((*SimpleInterface)(nil)).ToSingletonConstructor(createSimpleInterfaceAndCount)
			singletonBuilder.EagerlyAndCall(callAndIncrement)
			return module
		}, false},
		{"CallEagerly", func() Module {
			module := NewModule()
			module.BindSingletonConstructor(createSimpleInterfaceAndCount)
			module.CallEagerly(callAndIncrement)
			return module
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doEagerSingletons(t, tt.modCreator, tt.eagerly, regular)
		})
		t.Run(tt.name+"_installed_module", func(t *testing.T) {
			doEagerSingletons(t, tt.modCreator, tt.eagerly, installed)
		})
		t.Run(tt.name+"_overriden_module", func(t *testing.T) {
			doEagerSingletons(t, tt.modCreator, tt.eagerly, overridden)
		})
	}
}

func doEagerSingletons(t *testing.T, creator func() Module, eagerly bool, modType moduleType) {
	simpleInterfaceCreateCount = 0
	callAndIncrementCount = 0
	var module Module

	switch modType {
	case regular:
		module = creator()
	case installed:
		module = NewModule()
		module.Install(creator())
	case overridden:
		module = Override(creator()).With(NewModule())
	}
	_, err := NewInjector(module)
	require.NoError(t, err)

	require.Equal(t, 1, simpleInterfaceCreateCount)
	if eagerly {
		require.Equal(t, 0, callAndIncrementCount)
	} else {
		require.Equal(t, 1, callAndIncrementCount)
	}
}

func TestBindBasicTypes(t *testing.T) {
	module := NewModule()

	module.Bind("").ToSingleton("singleton")
	module.Bind(false).ToSingleton(true)
	module.Bind(0).ToSingleton(-99)
	module.Bind(int8(0)).ToSingleton(int8(-8))
	module.Bind(int16(0)).ToSingleton(int16(-16))
	module.Bind(int32(0)).ToSingleton(int32(-32))
	module.Bind(int64(0)).ToSingleton(int64(-64))
	module.Bind(uint(0)).ToSingleton(uint(99))
	module.Bind(uint8(0)).ToSingleton(uint8(8))
	module.Bind(uint16(0)).ToSingleton(uint16(16))
	module.Bind(uint32(0)).ToSingleton(uint32(32))
	module.Bind(uint64(0)).ToSingleton(uint64(64))

	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			var o interface{}
			var err error

			o, err = injector.Get(true)
			require.NoError(t, err)
			require.Equal(t, o, true)

			o, err = injector.Get("")
			require.NoError(t, err)
			require.Equal(t, o, "singleton")

			o, err = injector.Get(0)
			require.NoError(t, err)
			require.Equal(t, o, -99)

			o, err = injector.Get(int8(0))
			require.NoError(t, err)
			require.Equal(t, o, int8(-8))

			o, err = injector.Get(int16(0))
			require.NoError(t, err)
			require.Equal(t, o, int16(-16))

			o, err = injector.Get(int32(0))
			require.NoError(t, err)
			require.Equal(t, o, int32(-32))

			o, err = injector.Get(int64(0))
			require.NoError(t, err)
			require.Equal(t, o, int64(-64))

			o, err = injector.Get(uint(0))
			require.NoError(t, err)
			require.Equal(t, o, uint(99))

			o, err = injector.Get(uint8(0))
			require.NoError(t, err)
			require.Equal(t, o, uint8(8))

			o, err = injector.Get(uint16(0))
			require.NoError(t, err)
			require.Equal(t, o, uint16(16))

			o, err = injector.Get(uint32(0))
			require.NoError(t, err)
			require.Equal(t, o, uint32(32))

			o, err = injector.Get(uint64(0))
			require.NoError(t, err)
			require.Equal(t, o, uint64(64))
		})
	}
}

func TestBindSlices(t *testing.T) {
	module := NewModule()

	type custom struct {
		name  string
		value int
	}
	type customSlice []custom
	cs := customSlice{
		custom{"one", 1},
		custom{"two", 2},
		custom{"three", 3},
	}
	customFn := func() customSlice {
		return cs
	}

	module.Bind([]byte{}).ToSingleton([]byte("singleton"))
	module.Bind([]int{}).ToSingleton([]int{0, 1, 2})
	module.BindSingletonConstructor(customFn)

	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			var o interface{}
			var err error

			o, err = injector.Get([]byte{})
			require.NoError(t, err)
			require.Equal(t, o, []byte("singleton"))

			o, err = injector.Get([]int{})
			require.NoError(t, err)
			require.Equal(t, o, []int{0, 1, 2})

			o, err = injector.Get(customSlice{})
			require.NoError(t, err)
			require.Equal(t, o, cs)
		})
	}
}

func TestBindMaps(t *testing.T) {
	module := NewModule()

	msi := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	mib := map[int]bool{
		1: false,
		2: true,
		3: false,
	}

	type custom struct {
		name  string
		value int
	}
	type customMap map[string]*custom
	msc := customMap{
		"one":   &custom{"one", 1},
		"two":   &custom{"two", 2},
		"three": &custom{"three", 3},
	}
	customFn := func() customMap {
		return msc
	}

	module.Bind(map[string]int{}).ToSingleton(msi)
	module.Bind(map[int]bool{}).ToSingleton(mib)
	module.BindSingletonConstructor(customFn)

	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			var o interface{}
			var err error

			o, err = injector.Get(map[string]int{})
			require.NoError(t, err)
			require.Equal(t, o, msi)

			o, err = injector.Get(map[int]bool{})
			require.NoError(t, err)
			require.Equal(t, o, mib)

			o, err = injector.Get(customMap{})
			require.NoError(t, err)
			require.Equal(t, o, msc)

		})
	}
}

func TestBindChannels(t *testing.T) {
	module := NewModule()

	strChan := make(chan string)
	var rec chan<- string = strChan
	var snd <-chan string = strChan

	type custom struct {
		name  string
		value int
	}
	type customChannel chan custom
	cs := make(customChannel)
	customFn := func() customChannel {
		return cs
	}

	module.Bind(strChan).ToSingleton(strChan)
	module.Bind((chan<- string)(nil)).ToSingleton(rec)
	module.Bind((<-chan string)(nil)).ToSingleton(snd)
	module.BindSingletonConstructor(customFn)

	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			var o interface{}
			var err error

			o, err = injector.Get((chan string)(nil))
			require.NoError(t, err)
			require.Equal(t, o, strChan)

			o, err = injector.Get((chan<- string)(nil))
			require.NoError(t, err)
			require.Equal(t, o, rec)

			o, err = injector.Get((<-chan string)(nil))
			require.NoError(t, err)
			require.Equal(t, o, snd)

			o, err = injector.Get((customChannel)(nil))
			require.NoError(t, err)
			require.Equal(t, o, cs)

		})
	}
}

type collection struct {
	s string
	i int
}

type stringProvider func() string

func provideString(c *collection) stringProvider {
	return func() string {
		return c.s
	}
}

type intProvider func() int

func provideInt(c *collection) intProvider {
	return func() int {
		return c.i
	}
}

func TestBindFunctions(t *testing.T) {
	module := NewModule()
	coll := &collection{
		s: "a string",
		i: 99,
	}
	module.BindSingleton(coll)
	module.BindSingletonConstructor(provideString)
	module.BindSingletonConstructor(provideInt)
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			_, err := injector.Call(func(sp stringProvider, ip intProvider) {
				require.Equal(t, coll.s, sp())
				require.Equal(t, coll.i, ip())
			})
			require.NoError(t, err)
		})
	}
}

func TestInjectorString(t *testing.T) {
	module := NewModule()
	module.Bind((*SimpleInterface)(nil), SimpleStruct{}).ToSingleton(SimpleStruct{"hello"})
	for _, injector := range createInjectors(t, module) {
		t.Run(injector.name, func(t *testing.T) {
			require.Regexp(t, "injector\\{.*github.com/eluv-io/inject-go.createInjectors}", injector.String(), "injector{")
		})
	}
}

// createInjectors creates three equivalent injectors:
// * a regular injector based on the given module
// * a child injector where all bindings are in the parent and
// * a child injector where all bindings are in the child
func createInjectors(t *testing.T, module Module) []*namedInjector {
	var err error
	var inj Injector
	var injectors []*namedInjector

	// regular injector
	inj, err = NewInjector(module)
	require.NoError(t, err)
	injectors = append(injectors, &namedInjector{inj, "regular"})

	// empty child injector
	{
		inj, err = NewInjector(module)
		require.NoError(t, err)
		inj, err = inj.NewChildInjector(nil)
		require.NoError(t, err)
		injectors = append(injectors, &namedInjector{inj, "empty child"})
	}

	// child injector with empty parent injector
	{
		inj, err = NewInjector()
		require.NoError(t, err)
		inj, err = inj.NewChildInjector(nil, module)
		require.NoError(t, err)
		injectors = append(injectors, &namedInjector{inj, "empty parent"})
	}

	return injectors
}

type namedInjector struct {
	Injector
	name string
}

package inject

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
)

const (
	goRoutineIterations = 100
)

type SimpleInterface interface {
	Foo() string
}

type SimpleStruct struct {
	foo string
}

func (this SimpleStruct) Foo() string {
	return this.foo
}

type SimplePtrStruct struct {
	foo string
}

func (this *SimplePtrStruct) Foo() string {
	return this.foo
}

// ***** simple Bind tests *****

func TestSimpleStructSingletonDirect(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*SimpleInterface)(nil)).ToSingleton(SimpleStruct{"hello"})
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.Get((*SimpleInterface)(nil))
	require.NoError(t, err)
	simpleInterface := object.(SimpleInterface)
	require.Equal(t, "hello", simpleInterface.Foo())
}

func TestSimpleStructSingletonDirectSucceedsWhenPtr(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimpleStruct{"hello"})
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.Get((*SimpleInterface)(nil))
	require.NoError(t, err)
	simpleInterface := object.(SimpleInterface)
	require.Equal(t, "hello", simpleInterface.Foo())
}

func TestSimplePtrStructSingletonDirect(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.Get((*SimpleInterface)(nil))
	require.NoError(t, err)
	simpleInterface := object.(SimpleInterface)
	require.Equal(t, "hello", simpleInterface.Foo())
}

func TestSimplePtrStructSingletonDirectFailsWhenNotPtr(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*SimpleInterface)(nil)).ToSingleton(SimplePtrStruct{"hello"})
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeDoesNotImplement)
}

func TestSimpleStructSingletonIndirect(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*SimpleInterface)(nil)).To(SimpleStruct{})
	require.NoError(t, err)
	err = module.Bind(SimpleStruct{}).ToSingleton(SimpleStruct{"hello"})
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.Get((*SimpleInterface)(nil))
	require.NoError(t, err)
	simpleInterface := object.(SimpleInterface)
	require.Equal(t, "hello", simpleInterface.Foo())
}

func TestSimpleStructSingletonIndirectSucceedsWhenPtr(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*SimpleInterface)(nil)).To(&SimpleStruct{})
	require.NoError(t, err)
	err = module.Bind(&SimpleStruct{}).ToSingleton(&SimpleStruct{"hello"})
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.Get((*SimpleInterface)(nil))
	require.NoError(t, err)
	simpleInterface := object.(SimpleInterface)
	require.Equal(t, "hello", simpleInterface.Foo())
}

func TestSimplePtrStructSingletonIndirect(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*SimpleInterface)(nil)).To(&SimplePtrStruct{})
	require.NoError(t, err)
	err = module.Bind(&SimplePtrStruct{}).ToSingleton(&SimplePtrStruct{"hello"})
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.Get((*SimpleInterface)(nil))
	require.NoError(t, err)
	simpleInterface := object.(SimpleInterface)
	require.Equal(t, "hello", simpleInterface.Foo())
}

func TestSimplePtrStructSingletonIndirectFailsWhenNotPtr(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*SimpleInterface)(nil)).To(SimplePtrStruct{})
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeDoesNotImplement)
}

func TestSimplePtrStructIndirectFailsWhenNoFinalBinding(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*SimpleInterface)(nil)).To(&SimplePtrStruct{})
	require.NoError(t, err)
	_, err = CreateInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoFinalBinding)
}

// ***** simple BindTagged tests *****

func TestTaggedTagEmpty(t *testing.T) {
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "").ToSingleton(SimpleStruct{"hello"})
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeTagEmpty)
}

func TestTaggedSimpleStructSingletonDirect(t *testing.T) {
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "tagOne").ToSingleton(SimpleStruct{"hello"})
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.GetTagged((*SimpleInterface)(nil), "tagOne")
	require.NoError(t, err)
	simpleInterface := object.(SimpleInterface)
	require.Equal(t, "hello", simpleInterface.Foo())
	_, err = injector.Get((*SimpleInterface)(nil))
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoBinding)
	_, err = injector.GetTagged((*SimpleInterface)(nil), "tagTwo")
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoBinding)
}

func TestTaggedSimpleStructSingletonDirectSucceedsWhenPtr(t *testing.T) {
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "tagOne").ToSingleton(&SimpleStruct{"hello"})
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.GetTagged((*SimpleInterface)(nil), "tagOne")
	require.NoError(t, err)
	simpleInterface := object.(SimpleInterface)
	require.Equal(t, "hello", simpleInterface.Foo())
	_, err = injector.Get((*SimpleInterface)(nil))
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoBinding)
	_, err = injector.GetTagged((*SimpleInterface)(nil), "tagTwo")
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoBinding)
}

func TestTaggedSimplePtrStructSingletonDirect(t *testing.T) {
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "tagOne").ToSingleton(&SimplePtrStruct{"hello"})
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.GetTagged((*SimpleInterface)(nil), "tagOne")
	require.NoError(t, err)
	simpleInterface := object.(SimpleInterface)
	require.Equal(t, "hello", simpleInterface.Foo())
	_, err = injector.Get((*SimpleInterface)(nil))
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoBinding)
	_, err = injector.GetTagged((*SimpleInterface)(nil), "tagTwo")
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoBinding)
}

func TestTaggedSimplePtrStructSingletonDirectFailsWhenNotPtr(t *testing.T) {
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "tagOne").ToSingleton(SimplePtrStruct{"hello"})
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeDoesNotImplement)
}

func TestTaggedSimpleStructSingletonIndirect(t *testing.T) {
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "tagOne").To(SimpleStruct{})
	require.NoError(t, err)
	err = module.Bind(SimpleStruct{}).ToSingleton(SimpleStruct{"hello"})
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.GetTagged((*SimpleInterface)(nil), "tagOne")
	require.NoError(t, err)
	simpleInterface := object.(SimpleInterface)
	require.Equal(t, "hello", simpleInterface.Foo())
	_, err = injector.Get((*SimpleInterface)(nil))
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoBinding)
	_, err = injector.GetTagged((*SimpleInterface)(nil), "tagTwo")
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoBinding)
}

func TestTaggedSimpleStructSingletonIndirectSucceedsWhenPtr(t *testing.T) {
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "tagOne").To(&SimpleStruct{})
	require.NoError(t, err)
	err = module.Bind(&SimpleStruct{}).ToSingleton(&SimpleStruct{"hello"})
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.GetTagged((*SimpleInterface)(nil), "tagOne")
	require.NoError(t, err)
	simpleInterface := object.(SimpleInterface)
	require.Equal(t, "hello", simpleInterface.Foo())
	_, err = injector.Get((*SimpleInterface)(nil))
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoBinding)
	_, err = injector.GetTagged((*SimpleInterface)(nil), "tagTwo")
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoBinding)
}

func TestTaggedSimplePtrStructSingletonIndirect(t *testing.T) {
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "tagOne").To(&SimplePtrStruct{})
	require.NoError(t, err)
	err = module.Bind(&SimplePtrStruct{}).ToSingleton(&SimplePtrStruct{"hello"})
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.GetTagged((*SimpleInterface)(nil), "tagOne")
	require.NoError(t, err)
	simpleInterface := object.(SimpleInterface)
	require.Equal(t, "hello", simpleInterface.Foo())
	_, err = injector.Get((*SimpleInterface)(nil))
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoBinding)
	_, err = injector.GetTagged((*SimpleInterface)(nil), "tagTwo")
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoBinding)
}

func TestTaggedSimplePtrStructSingletonIndirectFailsWhenNotPtr(t *testing.T) {
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "tagOne").To(SimplePtrStruct{})
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeDoesNotImplement)
}

func TestTaggedSimplePtrStructIndirectFailsWhenNoFinalBinding(t *testing.T) {
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "tagOne").To(&SimplePtrStruct{})
	require.NoError(t, err)
	_, err = CreateInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoFinalBinding)
}

// ***** additional simple tagged tests *****

func TestTaggedSimpleStructSingletonDirectTwoBindings(t *testing.T) {
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "tagOne").ToSingleton(SimpleStruct{"hello"})
	require.NoError(t, err)
	err = module.BindTagged((*SimpleInterface)(nil), "tagTwo").ToSingleton(SimpleStruct{"good day"})
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.GetTagged((*SimpleInterface)(nil), "tagOne")
	require.NoError(t, err)
	simpleInterface := object.(SimpleInterface)
	require.Equal(t, "hello", simpleInterface.Foo())
	object, err = injector.GetTagged((*SimpleInterface)(nil), "tagTwo")
	require.NoError(t, err)
	simpleInterface = object.(SimpleInterface)
	require.Equal(t, "good day", simpleInterface.Foo())
}

// ***** simple constructor tests *****

type BarInterface interface {
	Bar() int
}

type BarPtrStruct struct {
	bar int
}

func (this *BarPtrStruct) Bar() int {
	return this.bar
}

type SecondInterface interface {
	Foo() SimpleInterface
	Bar() BarInterface
}

type SecondPtrStruct struct {
	foo SimpleInterface
	bar BarInterface
}

func (this *SecondPtrStruct) Foo() SimpleInterface {
	return this.foo
}

func (this *SecondPtrStruct) Bar() BarInterface {
	return this.bar
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

func TestConstructorDirectInterfaceInjection(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	require.NoError(t, err)
	err = module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	require.NoError(t, err)
	err = module.Bind((*SecondInterface)(nil)).ToConstructor(createSecondInterface)
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.Get((*SecondInterface)(nil))
	require.NoError(t, err)
	secondInterface := object.(SecondInterface)
	require.Equal(t, "hello", secondInterface.Foo().Foo())
	require.Equal(t, 1, secondInterface.Bar().Bar())
}

func TestConstructorErrReturned(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	require.NoError(t, err)
	err = module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	require.NoError(t, err)
	err = module.Bind((*SecondInterface)(nil)).ToConstructor(createSecondInterfaceErr)
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	_, err = injector.Get((*SecondInterface)(nil))
	require.Equal(t, "XYZ", err.Error())
}

func TestConstructorErrNoBindingReturned(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	require.NoError(t, err)
	err = module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	require.NoError(t, err)
	err = module.Bind((*SecondInterface)(nil)).ToConstructor(createSecondInterfaceErrNoBinding)
	require.NoError(t, err)
	_, err = CreateInjector(module)
	require.Error(t, err)
	require.Contains(t, err.Error(), InjectErrorTypeNoBinding)
	require.Contains(t, err.Error(), "inject.UnboundInterface")
}

func TestConstructorWithEvilCounter(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*BarInterface)(nil)).ToConstructor(createEvilBarInterface)
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)

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
}

func TestSingletonConstructorWithEvilCounter(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*BarInterface)(nil)).ToSingletonConstructor(createEvilBarInterface)
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)

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
}

func TestSingletonConstructorWithEvilCounterErr(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*BarInterface)(nil)).ToSingletonConstructor(createEvilBarInterfaceErr)
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)

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
		require.Equal(t, "XYZ 1", barInterfaceErr.err.Error())
	}

	close(evilChan)
}

func TestSingletonConstructorWithEvilCounterMultipleInjectors(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*BarInterface)(nil)).ToSingletonConstructor(createEvilBarInterface)
	require.NoError(t, err)
	injector1, err := CreateInjector(module)
	require.NoError(t, err)
	injector2, err := CreateInjector(module)
	require.NoError(t, err)
	injector3, err := CreateInjector(module)
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
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "tagOne").ToSingleton(&SimplePtrStruct{"hello"})
	require.NoError(t, err)
	err = module.BindTagged((*BarInterface)(nil), "tagTwo").ToSingleton(&BarPtrStruct{1})
	require.NoError(t, err)
	err = module.Bind((*SecondInterface)(nil)).ToTaggedConstructor(createSecondInterfaceTaggedOne)
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.Get((*SecondInterface)(nil))
	require.NoError(t, err)
	secondInterface := object.(SecondInterface)
	require.Equal(t, "hello", secondInterface.Foo().Foo())
	require.Equal(t, 1, secondInterface.Bar().Bar())
}

func TestTaggedConstructorOneHasNoTag(t *testing.T) {
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "tagOne").ToSingleton(&SimplePtrStruct{"hello"})
	require.NoError(t, err)
	err = module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	require.NoError(t, err)
	err = module.Bind((*SecondInterface)(nil)).ToTaggedConstructor(createSecondInterfaceTaggedOneHasNoTag)
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.Get((*SecondInterface)(nil))
	require.NoError(t, err)
	secondInterface := object.(SecondInterface)
	require.Equal(t, "hello", secondInterface.Foo().Foo())
	require.Equal(t, 1, secondInterface.Bar().Bar())
}

func TestTaggedConstructorNoTags(t *testing.T) {
	module := CreateModule()
	err := module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"hello"})
	require.NoError(t, err)
	err = module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	require.NoError(t, err)
	err = module.Bind((*SecondInterface)(nil)).ToTaggedConstructor(createSecondInterfaceTaggedNoTags)
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.Get((*SecondInterface)(nil))
	require.NoError(t, err)
	secondInterface := object.(SecondInterface)
	require.Equal(t, "hello", secondInterface.Foo().Foo())
	require.Equal(t, 1, secondInterface.Bar().Bar())
}

func TestTaggedConstructorOneHasNoTagMultipleBindings(t *testing.T) {
	module := CreateModule()
	err := module.BindTagged((*SimpleInterface)(nil), "tagOne").ToSingleton(&SimplePtrStruct{"hello"})
	require.NoError(t, err)
	err = module.BindTagged((*SimpleInterface)(nil), "tagTwo").ToSingleton(&SimplePtrStruct{"goodbye"})
	require.NoError(t, err)
	err = module.Bind((*SimpleInterface)(nil)).ToSingleton(&SimplePtrStruct{"another"})
	require.NoError(t, err)
	err = module.Bind((*BarInterface)(nil)).ToSingleton(&BarPtrStruct{1})
	require.NoError(t, err)
	err = module.Bind((*SecondInterface)(nil)).ToTaggedConstructor(createSecondInterfaceTaggedOneHasNoTag)
	require.NoError(t, err)
	injector, err := CreateInjector(module)
	require.NoError(t, err)
	object, err := injector.Get((*SecondInterface)(nil))
	require.NoError(t, err)
	secondInterface := object.(SecondInterface)
	require.Equal(t, "hello", secondInterface.Foo().Foo())
	require.Equal(t, 1, secondInterface.Bar().Bar())
	object, err = injector.GetTagged((*SimpleInterface)(nil), "tagTwo")
	require.NoError(t, err)
	require.Equal(t, "goodbye", object.(SimpleInterface).Foo())
	object, err = injector.Get((*SimpleInterface)(nil))
	require.NoError(t, err)
	require.Equal(t, "another", object.(SimpleInterface).Foo())
}

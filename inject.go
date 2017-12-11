/*
Package inject is guice-inspired dependency injection for Go.

https://github.com/google/guice/wiki/Motivation

This project is in no way affiliated with the Guice project, but I recommend reading their
docs to better understand the concepts.


Concepts


	Module

A Module is analogous to Guice's AbstractModule, used for setting up your dependencies.
This allows you to bind structs, struct pointers, interfaces, and primitives to
singletons, constructors, with or without tags.

An interface can have a binding to another type, or to a singleton or constructor.

	type SayHello interface {
		Hello() string
	}

	type SayHelloOne struct {
		value string
	}

	func (i *SayHelloOne) Hello() string {
		return i.value
	}

	module := inject.NewModule()
	module.BindInterface((*SayHello)(nil)).To(&SayHelloOne{}) // valid, but must provide a binding to *SayHelloOne.
	module.Bind(&SayHelloOne{}).ToSingleton(&SayHelloOne{"Salutations"}) // there we go

An interface can also be bound to a singleton or constructor.

	module.Bind((*SayHello)(nil)).ToSingleton(&SayHelloOne{"Salutations"})

A struct, struct pointer, or primitive must have a direct binding to a singleton or constructor.

All errors from binding will be returned as one error when calling inject.NewInjector(...).

	Injector

An Injector is analogous to Guice's Injector, providing your dependencies.

Given the binding:

	module.Bind((*SayHello)(nil)).ToSingleton(&SayHelloOne{"Salutations"})

We are able to get a value for SayHello.

	func printHello(aboveModule inject.Module) error {
		injector, err := inject.NewInjector(aboveModule)
		if err != nil {
			return err
		}
		sayHelloObj, err := injector.Get((*SayHello)(nil))
		if err != nil {
			return err
		}
		fmt.Println(sayHelloObj.(SayHello).Hello()) // will print "Salutations"
		return nil
	}

See the Injector interface for other methods.


	Constructor

A constructor is a function that takes injected values as parameters, and returns a value and an error.

	type SayHelloToSomeone interface {
		Greetings() string
	}

	type SayHelloToSomeoneOne struct {
		sayHello SayHello
		person string
	}

	func (i *SayHelloToSomeoneOne) Greetings() string {
		return fmt.Sprintf("%s, %s!", i.sayHello.Hello(), i.person)
	}

We can set up a constructor to take zero values, or values that require a binding in some module
passed to NewInjector().

	func doStuff() error {
		m1 := inject.NewModule()
		m1.Bind((*SayHello)(nil)).ToSingleton(&SayHelloOne{"Salutations"})
		m2 := inject.NewModule()
		m2.Bind((*SayHelloToSomeone)(nil)).ToConstructor(newSayHelloToSomeone)
		injector, err := inject.NewInjector(m1, m2)
		if err != nil {
			return err
		}
		sayHelloToSomeoneObj, err := injector.Get((*SayHelloToSomeone)(nil))
		if err != nil {
			return err
		}
		fmt.Println(sayHelloToSomeoneObj.(SayHelloToSomeone).Greetings()) // will print "Saluatations, Alice!"
		return nil
	}

	func newSayHelloToSomeone(sayHello SayHello) (SayHelloToSomeone, error) {
		return &SayHelloToSomeoneOne{sayHello, "Alice"}, nil
	}

A singleton constructor will be called exactly once for the entire application.

	var (
		unsafeCounter := 0
	)

	func doStuff() error {
		m1 := inject.NewModule()
		m1.Bind((*SayHello)(nil)).ToSingleton(&SayHelloOne{"Salutations"})
		m2 := inject.NewModule()
		m2.Bind((*SayHelloToSomeone)(nil)).ToSingletonConstructor(newSayHelloToSomeone)
		injector, err := inject.NewInjector(m1, m2)
		if err != nil {
			return err
		}
		sayHelloToSomeoneObj1, err := injector.Get((*SayHelloToSomeone)(nil))
		if err != nil {
			return err
		}
		fmt.Println(sayHelloToSomeoneObj1.(SayHelloToSomeone).Greetings()) // will print "Saluatations, Alice1!"
		sayHelloToSomeoneObj2, err := injector.Get((*SayHelloToSomeone)(nil))
		if err != nil {
			return err
		}
		fmt.Println(sayHelloToSomeoneObj2.(SayHelloToSomeone).Greetings()) // will print "Saluatations, Alice1!"
		return nil
	}

	func newSayHelloToSomeone(sayHello SayHello) (SayHelloToSomeone, error) {
		unsafeCounter++
		return &SayHelloToSomeoneOne{sayHello, fmt.Sprintf("Alice%d", unsafeCounter)}, nil
	}

Functions be called from an injector using the Call function. These functions have the same parameter
requirements as constructors, but can have any return types.

	func doStuffWithAboveM1(m1 inject.Module) error {
		injector, err := inject.NewInjector(m1)
		if err != nil {
			return err
		}
		values, err := injector.Call(getStuff)
		if err != nil {
			return err
		}
		fmt.Println(values[0) // "Salutations"
		fmt.Println(values[1]) // 4
		return nil
	}

	func getStuff(sayHello SayHello) (string, int) {
		return sayHello.Hello(), 4
	}

See the methods on Module and Constructor for more details.


	Tags

A tag allows named multiple bindings of one type. As an example, let's consider if we want to
have multiple ways to say hello.

	func doStuff() error {
		module := inject.NewModule()
		module.BindTagged("english", (*SayHello)(nil)).ToSingleton(&SayHelloOne{"Hello"})
		module.BindTagged("german", (*SayHello)(nil)).ToSingleton(&SayHelloOne{"Guten Tag"})
		module.BindTagged("austrian", (*SayHello)(nil)).ToSingleton(&SayHelloOne{"Grüß Gott"})
		injector, err := inject.NewInjector(module)
		if err != nil {
			return err
		}
		_ = printHello("english", injector) // not error checking for the sake of shorter docs
		_ = printHello("german", injector)
		_ = printHello("austrian", injector)
		return nil
	}

	func printHello(tag string, injector inject.Injector) error {
		sayHelloObj, err := injector.GetTagged(tag)
		if err != nil {
			return err
		}
		fmt.Println(sayHelloObj.(SayHello).Hello())
		return nil
	}

Structs can also be populated using the tag "inject".

	type PopulateOne struct {
		// must be public
		English SayHello `inject:"english"`
		German SayHello `inject:"german"`
		Austrian SayHello `inject:"austrian"`
	}

	func printAllHellos(aboveInjector inject.Injector) error {
		populateOne := &PopulateOne{}
		if err := injector.Populate(populateOne); err != nil {
			return err
		}
		fmt.Println(populateOne.English.Hello())
		fmt.Println(populateOne.German.Hello())
		fmt.Println(populateOne.Austrian.Hello())
		return nil
	}

Constructors can be tagged using structs, either named or anonymous.

	type SayHowdy struct { // not interface, for this example
		value string
	}

	func(s *SayHowdy) Howdy() {
		return s.value
	}

	type PopulateTwo struct {
		PopulateOne *PopulateOne
		SayHowdy *SayHowdy
	}

	func doStuff(aboveModule inject.Module) error {
		module := inject.NewModule()
		module.Bind(&PopulateTwo).ToTaggedConstructor(newPopulateTwo)
		injector, err := inject.NewInjector(aboveModule, module)
		if err != nil {
			return err
		}
		populateTwo, err := injector.Get(&PopulateTwo{})
		if err != nil {
			return err
		}
		fmt.Printf("%+v\n", populateTwo)
		return nil
	}

	func newPopulateTwo(populateOne *PopulateOne) (*PopulateTwo, error) {
		return &PopulateTwo{
			PopulateOne, populateOne,
			SayHowdy: &SayHowdy{"howdy"},
		}, nil
	}

	// an anonymous struct can also be used in a constructor.
	func newPopulateTwoAnonymous(str struct {
		// must be public
		English SayHello `inject:"english"`
		German SayHello `inject:"german"`
		Austrian SayHello `inject:"austrian"`
	}) (*PopulateTwo, error) {
		return &PopulateTwo{
			PopulateOne: &PopulateOne{
				English: str.English,
				German: str.German,
				Austrian: str.Austrian,
			},
			SayHowdy: &SayHowdy{"howdy"},
		}, nil
	}

A constructor can mix tagged values with untagged values in the input struct.

	func doStuff(aboveModuleAgain injector.Module) error {
		aboveModuleAgain.Bind(&SayHowdy{}).ToSingleton(&SayHowdy{"howdy"})
		aboveModuleAgain.Bind(&PopulateTwo).ToTaggedConstructor(newPopulateTwo)
		injector, err := inject.NewInjector(aboveModuleAgain)
		if err != nil {
			return err
		}
		populateTwo, err := injector.Get(&PopulateTwo{})
		if err != nil {
			return err
		}
		fmt.Printf("%+v\n", populateTwo)
		return nil
	}

	func newPopulateTwo(str struct {
		English SayHello `inject:"english"`
		German SayHello `inject:"german"`
		Austrian SayHello `inject:"austrian"`
		SayHowdy SayHowdy
	}) (*PopulateTwo, error) {
		return &PopulateTwo{
			PopulateOne: &PopulateOne{
				English: str.English,
				German: str.German,
				Austrian: str.Austrian,
			},
			SayHowdy: str.SayHowdy,
		}, nil
	}

The CallTagged function works similarly to Call, except can take parameters like a tagged constructor.


Both Module and Injector implement fmt.Stringer for inspection, however this may be added to in the future
to allow semantic inspection of bindings.

For testing, production modules may be overridden with test bindings as follows:

	module := createProductionModule()

	override := NewModule()
	override.Bind((*ExternalService)(nil)).ToSingleton(createMockExternalService())

	injector, err := NewInjector(Override(module).With(override))

*/
package inject // import "go.pedge.io/inject"

import (
	"fmt"
)

// Module sets up your dependencies.
//
// Note that none of the calls to Module are thread-safe, it is your responsibility
// to make sure multiple goroutines are not calling a single module.
type Module interface {
	fmt.Stringer
	Bind(from ...interface{}) Builder
	BindTagged(tag string, from ...interface{}) Builder
	BindInterface(fromInterface ...interface{}) InterfaceBuilder
	BindTaggedInterface(tag string, fromInterface ...interface{}) InterfaceBuilder
	BindTaggedBool(tag string) Builder
	BindTaggedInt(tag string) Builder
	BindTaggedInt8(tag string) Builder
	BindTaggedInt16(tag string) Builder
	BindTaggedInt32(tag string) Builder
	BindTaggedInt64(tag string) Builder
	BindTaggedUint(tag string) Builder
	BindTaggedUint8(tag string) Builder
	BindTaggedUint16(tag string) Builder
	BindTaggedUint32(tag string) Builder
	BindTaggedUint64(tag string) Builder
	BindTaggedFloat32(tag string) Builder
	BindTaggedFloat64(tag string) Builder
	BindTaggedComplex64(tag string) Builder
	BindTaggedComplex128(tag string) Builder
	BindTaggedString(tag string) Builder
	Install(others ...Module)
}

// NewModule creates a new Module.
func NewModule() Module { return newModule() }

// Builder is the return value from a Bind call from a Module.
type Builder interface {
	ToSingleton(singleton interface{})
	ToConstructor(constructor interface{})
	ToSingletonConstructor(constructor interface{})
	ToTaggedConstructor(constructor interface{})
	ToTaggedSingletonConstructor(constructor interface{})
}

// InterfaceBuilder is the return value when binding an interface from a Module.
type InterfaceBuilder interface {
	Builder
	To(to interface{})
}

// Injector provides your dependencies.
type Injector interface {
	fmt.Stringer
	Get(from interface{}) (interface{}, error)
	GetTagged(tag string, from interface{}) (interface{}, error)
	GetTaggedBool(tag string) (bool, error)
	GetTaggedInt(tag string) (int, error)
	GetTaggedInt8(tag string) (int8, error)
	GetTaggedInt16(tag string) (int16, error)
	GetTaggedInt32(tag string) (int32, error)
	GetTaggedInt64(tag string) (int64, error)
	GetTaggedUint(tag string) (uint, error)
	GetTaggedUint8(tag string) (uint8, error)
	GetTaggedUint16(tag string) (uint16, error)
	GetTaggedUint32(tag string) (uint32, error)
	GetTaggedUint64(tag string) (uint64, error)
	GetTaggedFloat32(tag string) (float32, error)
	GetTaggedFloat64(tag string) (float64, error)
	GetTaggedComplex64(tag string) (complex64, error)
	GetTaggedComplex128(tag string) (complex128, error)
	GetTaggedString(tag string) (string, error)
	Call(function interface{}) ([]interface{}, error)
	CallTagged(taggedFunction interface{}) ([]interface{}, error)
	Populate(populateStruct interface{}) error
}

// NewInjector creates a new Injector for the specified Modules.
//
// Note that Modules are not thread-safe, it is your responsibility to make sure
// all Modules have all bindings in place before passing them as parameters to NewInjector.
func NewInjector(modules ...Module) (Injector, error) { return newInjector(modules) }

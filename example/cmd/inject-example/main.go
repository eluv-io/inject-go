package main

import (
	"fmt"
	"os"

	"github.com/eluv-io/inject-go"
	"github.com/eluv-io/inject-go/example/api"
	"github.com/eluv-io/inject-go/example/cloud"
	"github.com/eluv-io/inject-go/example/more"
	"github.com/eluv-io/inject-go/example/stuff"
)

func main() {
	if err := do(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func do() error {
	injector, err := inject.NewInjector(
		api.NewModule(),
		cloud.NewModule(),
		more.NewModule(),
		stuff.NewModule(),
	)
	if err != nil {
		return err
	}
	obj, err := injector.Get((*api.Api)(nil))
	if err != nil {
		return err
	}
	apiObj := obj.(api.Api)
	provider := "aws"
	if len(os.Args) > 1 {
		provider = os.Args[1]
	}
	response, err := apiObj.Do(api.Request{Provider: provider, Foo: "this is fun"})
	if err != nil {
		return err
	}
	fmt.Printf("%v %v\n", response.Bar, response.Baz)
	return nil
}

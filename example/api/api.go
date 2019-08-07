package api // import "github.com/eluv-io/inject-go/example/api"

import (
	"fmt"

	"github.com/eluv-io/inject-go"
	"github.com/eluv-io/inject-go/example/cloud"
	"github.com/eluv-io/inject-go/example/more"
)

func NewModule() inject.Module {
	module := inject.NewModule()
	module.Bind((*Api)(nil)).ToTaggedSingletonConstructor(createApi)
	return module
}

type Request struct {
	Provider string
	Foo      string
}

type Response struct {
	Bar string
	Baz int
}

type Api interface {
	Do(Request) (*Response, error)
}

type api struct {
	awsProvider          cloud.Provider
	digitalOceanProvider cloud.Provider
	moreThings           more.MoreThings
}

func createApi(s struct {
	AwsProvider          cloud.Provider `inject:"aws"`
	DigitalOceanProvider cloud.Provider `inject:"digital_ocean"`
	MoreThings           more.MoreThings
}) (Api, error) {
	return &api{s.AwsProvider, s.DigitalOceanProvider, s.MoreThings}, nil
}

func (a *api) Do(request Request) (*Response, error) {
	provider, err := a.getProvider(request.Provider)
	if err != nil {
		return nil, err
	}
	instance, err := provider.NewInstance()
	if err != nil {
		return nil, err
	}
	result, err := instance.RunCommand(cloud.Command{Path: "ls"})
	if err != nil {
		return nil, err
	}
	s, err := a.moreThings.MoreStuffToDo(1)
	if err != nil {
		return nil, err
	}
	return &Response{fmt.Sprintf("%s %s %s", request.Foo, result.Message, s), result.ExitCode}, nil
}

func (a *api) getProvider(provider string) (cloud.Provider, error) {
	switch provider {
	case "aws":
		return a.awsProvider, nil
	case "digital_ocean":
		return a.digitalOceanProvider, nil
	default:
		return nil, fmt.Errorf("api: Unknown provider %v", provider)
	}
}

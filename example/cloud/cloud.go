package cloud // import "github.com/eluv-io/inject-go/example/cloud"

import (
	"github.com/eluv-io/inject-go"
	"github.com/eluv-io/inject-go/example/stuff"
)

func NewModule() inject.Module {
	module := inject.NewModule()
	module.BindTagged("aws", (*Provider)(nil)).ToSingletonConstructor(createAwsProvider)
	module.BindTagged("digital_ocean", (*Provider)(nil)).ToSingletonConstructor(createDigitalOceanProvider)
	return module
}

type Command struct {
	Path string
}

type Result struct {
	Message  string
	ExitCode int
}

type Instance interface {
	RunCommand(Command) (*Result, error)
}

type Provider interface {
	NewInstance() (Instance, error)
}

type instance struct {
	data         string
	stuffService stuff.StuffService
}

func (i *instance) RunCommand(command Command) (*Result, error) {
	ii, err := i.stuffService.DoStuff("pwd")
	if err != nil {
		return nil, err
	}
	if command.Path == "ls" {
		return &Result{i.data, ii}, nil
	} else {
		return &Result{i.data, 1}, nil
	}
}

type awsProvider struct {
	stuffService stuff.StuffService
}

func createAwsProvider(stuffService stuff.StuffService) (Provider, error) {
	return &awsProvider{stuffService}, nil
}

func (i *awsProvider) NewInstance() (Instance, error) {
	return &instance{"aws can do stuff", i.stuffService}, nil
}

type digitalOceanProvider struct {
	stuffService stuff.StuffService
}

func createDigitalOceanProvider(stuffService stuff.StuffService) (Provider, error) {
	return &digitalOceanProvider{stuffService}, nil
}

func (i *digitalOceanProvider) NewInstance() (Instance, error) {
	return &instance{"digitalOcean can also do stuff", i.stuffService}, nil
}

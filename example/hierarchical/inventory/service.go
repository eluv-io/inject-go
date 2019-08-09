package inventory

import (
	"fmt"

	"github.com/eluv-io/inject-go"
	"github.com/eluv-io/inject-go/example/hierarchical"
)

func CreateService(parent inject.Injector, config Config) (hierarchical.Service, error) {
	inj, err := parent.NewChildInjector(nil, NewModule(config))
	if err != nil {
		return nil, err
	}
	svc, err := inj.Get((*Service)(nil))
	return svc.(*Service), err
}

func newService(config Config) *Service {
	return &Service{
		config: config,
	}
}

type Config struct {
}

type Service struct {
	config Config
}

func (Service) Start() {
	fmt.Println("Inventory service starting")
}

func (Service) Stop() {
	fmt.Println("Inventory service stopping")
}

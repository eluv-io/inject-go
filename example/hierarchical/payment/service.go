package payment

import (
	"fmt"

	"github.com/eluv-io/inject-go"
	"github.com/eluv-io/inject-go/example/hierarchical"
)

type Overrides inject.Module

func CreateService(parent inject.Injector, config Config) (hierarchical.Service, error) {
	m := NewModule(config)
	inj, err := parent.NewChildInjector((*Overrides)(nil), m)
	if err != nil {
		return nil, err
	}
	svc, err := inj.Get((*Service)(nil))
	return svc.(*Service), err
}

func newService(config Config, processor Processor) *Service {
	return &Service{
		config:    config,
		processor: processor,
	}
}

type Config struct {
}

type Service struct {
	config    Config
	processor Processor
}

func (s *Service) Start() {
	fmt.Println("Payment service starting")
	s.processor.Process(19.75)
}

func (s *Service) Stop() {
	fmt.Println("Payment service stopping")
}

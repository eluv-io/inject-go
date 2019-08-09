package global

import (
	"errors"
	"fmt"

	"github.com/eluv-io/inject-go"
	"github.com/eluv-io/inject-go/example/hierarchical"
	"github.com/eluv-io/inject-go/example/hierarchical/inventory"
	"github.com/eluv-io/inject-go/example/hierarchical/payment"
)

func newServicesFactory(inj inject.Injector) *ServicesFactory {
	return &ServicesFactory{inj: inj}
}

type ServicesFactory struct {
	inj inject.Injector // global injector
}

func (f *ServicesFactory) CreateService(sc ServiceConfig) (hierarchical.Service, error) {
	switch sc.Type {
	case "Payment":
		return payment.CreateService(f.inj, sc.Config.(payment.Config))
	case "Inventory":
		return inventory.CreateService(f.inj, sc.Config.(inventory.Config))
	}
	return nil, errors.New(fmt.Sprintf("unknown service type [%s]", sc.Type))
}

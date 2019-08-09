package payment

import (
	"github.com/eluv-io/inject-go"
)

func NewModule(conf Config) inject.Module {
	m := inject.NewModule()
	m.Bind(Config{}).ToSingleton(conf)
	m.BindSingletonConstructor(newService)
	m.BindSingletonConstructor(newCreditCardProcessor)
	return m
}

package global

import (
	"github.com/eluv-io/inject-go"
)

func NewModule() inject.Module {
	m := inject.NewModule()
	m.BindSingletonConstructor(loadConfig)
	m.BindSingletonConstructor(newApp)
	m.BindSingletonConstructor(newServicesFactory)
	m.BindSingletonConstructor(newStore)
	return m
}

package inventory

import (
	"github.com/eluv-io/inject-go"
)

func NewModule(conf Config) inject.Module {
	m := inject.NewModule()
	m.Bind(Config{}).ToSingleton(conf)
	m.BindSingletonConstructor(newService)
	return m
}

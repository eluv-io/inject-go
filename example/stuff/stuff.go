package stuff // import "github.com/eluv-io/inject-go/example/stuff"

import (
	"github.com/eluv-io/inject-go"
)

func NewModule() inject.Module {
	module := inject.NewModule()
	module.Bind((*StuffService)(nil)).ToConstructor(createStuffService)
	return module
}

type StuffService interface {
	DoStuff(string) (int, error)
}

type stuffService struct{}

func createStuffService() (StuffService, error) {
	return &stuffService{}, nil
}

func (s *stuffService) DoStuff(ss string) (int, error) {
	if ss == "pwd" {
		return 0, nil
	} else {
		return -1, nil
	}
}

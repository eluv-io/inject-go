package more

import (
	"fmt"

	"github.com/peter-edge/go-inject"
)

func CreateModule() inject.Module {
	module := inject.CreateModule()
	module.Bind((*MoreThings)(nil)).ToSingleton(&moreThings{})
	return module
}

type MoreThings interface {
	MoreStuffToDo(int) (string, error)
}

type moreThings struct{}

func (m *moreThings) MoreStuffToDo(i int) (string, error) {
	return fmt.Sprintf("but there's not much to do here %v", i), nil
}

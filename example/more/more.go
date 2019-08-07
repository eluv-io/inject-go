package more // import "github.com/eluv-io/inject-go/example/more"

import (
	"fmt"

	"github.com/eluv-io/inject-go"
)

func NewModule() inject.Module {
	module := inject.NewModule()
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

package global

import (
	"fmt"
	"log"

	"github.com/eluv-io/inject-go"
	"github.com/eluv-io/inject-go/example/hierarchical"
)

func CreateApp() (*App, error) {
	return CreateAppFrom(NewModule())
}

func CreateAppFrom(m inject.Module) (*App, error) {
	inj, err := inject.NewInjector(m)
	if err != nil {
		return nil, err
	}
	app, err := inj.Get((*App)(nil))
	if err != nil {
		return nil, err
	}

	// also retrieve and print the dependency tree
	tree, _ := inj.DependencyTree()
	fmt.Println(tree)

	return app.(*App), err
}

func newApp(config Config, factory *ServicesFactory) *App {
	app := &App{config: config}
	for _, svc := range config.Services {
		service, err := factory.CreateService(svc)
		if err != nil {
			log.Fatal(err)
		}
		app.services = append(app.services, service)
	}
	return app
}

type App struct {
	config   Config
	services []hierarchical.Service
}

func (a *App) Start() {
	fmt.Println("Starting services")
	for _, svc := range a.services {
		svc.Start()
	}
}

func (a *App) Stop() {
	fmt.Println("Stopping services")
	for _, svc := range a.services {
		svc.Stop()
	}
}

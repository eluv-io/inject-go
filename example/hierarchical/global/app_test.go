package global_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/eluv-io/inject-go"
	"github.com/eluv-io/inject-go/example/hierarchical/global"
	"github.com/eluv-io/inject-go/example/hierarchical/payment"
)

func TestApp(t *testing.T) {
	paymentModule := inject.NewModule()
	paymentModule.BindSingletonConstructor(newMockProcessor)

	overrides := inject.NewModule()
	overrides.Bind((*payment.Overrides)(nil)).ToSingleton(paymentModule)
	m := inject.Override(global.NewModule()).With(overrides)

	// in order to add the payment.Overrides, it's not absolutely necessary to
	// override the parent bindings, since payment.Overrides is never bound in
	// the production code. So this would actually be enough:
	//   m := global.NewModule()
	//   m.Bind((*payment.Overrides)(nil)).ToSingleton(paymentModule)

	app, err := global.CreateAppFrom(m)
	if err != nil {
		log.Fatal(err)
	}
	app.Start()
	app.Stop()
}

type mockProcessor struct {
}

func (m *mockProcessor) Process(amount float64) {
	fmt.Printf("mock processing $%.2f\n", amount)
}

func newMockProcessor() payment.Processor {
	return &mockProcessor{}
}

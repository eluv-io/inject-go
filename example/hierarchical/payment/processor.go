package payment

import (
	"github.com/eluv-io/inject-go/example/hierarchical"
)

type Processor interface {
	Process(amount float64)
}

func newCreditCardProcessor(store hierarchical.Store) Processor {
	return &CreditCardProcessor{store}
}

type CreditCardProcessor struct {
	store hierarchical.Store
}

func (*CreditCardProcessor) Process(amount float64) {
	panic("CreditCardProcessor.Process no implemented")
}

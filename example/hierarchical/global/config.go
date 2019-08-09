package global

import (
	"github.com/eluv-io/inject-go/example/hierarchical/inventory"
	"github.com/eluv-io/inject-go/example/hierarchical/payment"
)

func loadConfig() Config {
	// this would normally load the config from e.g. a file or database
	return Config{
		Services: []ServiceConfig{
			{
				Type:   "Payment",
				Config: payment.Config{},
			},
			{
				Type:   "Inventory",
				Config: inventory.Config{},
			},
		},
	}
}

type ServiceConfig struct {
	Type   string
	Config interface{}
}

type Config struct {
	Services []ServiceConfig
}

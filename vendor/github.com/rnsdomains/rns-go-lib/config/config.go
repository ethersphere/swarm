package config

import (
	"github.com/caarlos0/env"
)

func init() {
	env.Parse(&cfg)
}

// Configuration is the struct that holds the values of the network configuration
// env and envDefault are required by env library
// it corresponds with the os environment variable name and also its default value if omitted
type Configuration struct {
	NetworkNodeAddress string `env:"RNS_NETWORK_NODE_ADDRESS" envDefault:"https://public-node.rsk.co"`
	ResolverAddress    string `env:"RNS_RESOLVER_ADDRESS" envDefault:"0x99a12be4C89CbF6CFD11d1F2c029904a7B644368"`
}

var cfg Configuration = Configuration{}

// GetConfiguration loads the environment variables into a Configuration struct and returns it
func GetConfiguration() Configuration {
	return cfg
}

// SetConfiguration sets the configuration for the blockchain endpoint and resolver contract
func SetConfiguration(endpoint string, contract string) {
	if endpoint != "" {
		cfg.NetworkNodeAddress = endpoint
	}
	if contract != "" {
		cfg.ResolverAddress = contract
	}
}

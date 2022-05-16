package main

import (
	"log"

	"github.com/brigadecore/brigade-foundations/os"
	"github.com/brigadecore/brigade/sdk/v3/restmachinery"
	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue/amqp"
)

func apiClientConfig() (string, string, restmachinery.APIClientOptions, error) {
	opts := restmachinery.APIClientOptions{}
	address, err := os.GetRequiredEnvVar("API_ADDRESS")
	if err != nil {
		return address, "", opts, err
	}
	log.Println("API_ADDRESS: ", address)
	token, err := os.GetRequiredEnvVar("API_TOKEN")
	if err != nil {
		return address, token, opts, err
	}
	opts.AllowInsecureConnections, err =
		os.GetBoolFromEnvVar("API_IGNORE_CERT_WARNINGS", false)
	return address, token, opts, err
}

// readerFactoryConfig returns an amqp.ReaderFactoryConfig based on
// configuration obtained from environment variables.
func readerFactoryConfig() (amqp.ReaderFactoryConfig, error) {
	config := amqp.ReaderFactoryConfig{}
	var err error
	config.Address, err = os.GetRequiredEnvVar("AMQP_ADDRESS")
	if err != nil {
		return config, err
	}
	log.Println("AMQP_ADDRESS: ", config.Address)
	config.Username, err = os.GetRequiredEnvVar("AMQP_USERNAME")
	if err != nil {
		return config, err
	}
	log.Println("AMQP_USERNAME: ", config.Username)
	config.Password, err = os.GetRequiredEnvVar("AMQP_PASSWORD")
	return config, err
}

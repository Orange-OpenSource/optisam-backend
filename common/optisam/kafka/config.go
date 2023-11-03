package kafka

import (
	"errors"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConfig struct {
	BootstrapServers       string
	SecurityProtocol       string
	SslKeyLocation         string
	SslCertificateLocation string
	SslCaLocation          string
}

func (k KafkaConfig) Validate() error {
	if k.BootstrapServers == "" {
		return errors.New("bootstrap servers is required")
	}
	return nil
}

func (k KafkaConfig) LoadConfig() kafka.ConfigMap {
	conf := kafka.ConfigMap{}
	conf.SetKey("bootstrap.servers", k.BootstrapServers)
	conf.SetKey("security.protocol", k.SecurityProtocol)
	conf.SetKey("ssl.key.location", k.SslKeyLocation)
	conf.SetKey("ssl.certificate.location", k.SslCertificateLocation)
	conf.SetKey("ssl.ca.location", k.SslCaLocation)
	return conf
}

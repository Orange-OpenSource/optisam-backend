package kafka

import (
	"errors"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

func BuildConsumer(c KafkaConfig, customConfig map[string]string) (consumer *kafka.Consumer, err error) {
	conf := c.LoadConfig()
	conf.SetKey("group.id", "kafka-optisam")
	conf.SetKey("auto.offset.reset", "earliest")
	for k, v := range customConfig {
		conf.SetKey(k, v)
	}
	consumer, err = kafka.NewConsumer(&conf)
	if err != nil {
		logger.Log.Sugar().Error("failed to create a new consumer: %v", zap.Error(err))
		err = errors.New("failed to create a new consumer:" + err.Error())
		return
	}
	return
}

func BuildProducer(c KafkaConfig, customConfig map[string]string) (producer *kafka.Producer, err error) {
	conf := c.LoadConfig()
	for k, v := range customConfig {
		conf.SetKey(k, v)
	}
	producer, err = kafka.NewProducer(&conf)
	if err != nil {
		logger.Log.Sugar().Error("failed to create a new producer: %v", zap.Error(err))
		err = errors.New("failed to create a new producer:" + err.Error())
		return
	}
	return
}

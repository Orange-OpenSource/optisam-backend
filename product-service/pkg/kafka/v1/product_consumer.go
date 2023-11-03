package v1

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkaConnect "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/kafka"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	v1Svc "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/service/v1"
)

const (
	TopicUpsertNominativeUsers                     = "upsert_nominative_users"
	TopicUpsertNominativeUsersRetry                = "upsert_nominative_users_retry"
	RetryDelay                                     = 30
	TopicProcessNomPostgresSuccess                 = "process_nom_postgres_success"
	TopicProcessNomPostgresSuccessRetry            = "process_nom_postgres_success_retry"
	TopicProcessNomDgraphSuccess                   = "process_nom_dgraph_success"
	TopicProcessNomDgraphSuccessRetry              = "process_nom_dgraph_success_retry"
	TopicUpsertNominativeUsersPostgres             = "upsert_nominative_users_postgres"
	TopicUpsertNominativeUsersPostgresRetry        = "upsert_nominative_users_postgres_retry"
	TopicUpsertNominativeUsersDgraph               = "upsert_nominative_users_dgraph"
	TopicUpsertNominativeUsersDgraphRetry          = "upsert_nominative_users_dgraph_retry"
	TopicUpdateNominativeUserRequest               = "update_nominative_user_request"
	TopicUpdateNominativeUserRequestRetry          = "update_nominative_user_request_retry"
	TopicDeadLetterQueue                           = "dead_letter_queue"
	TopicProcessNominativeUserRecordsPostgresRetry = "process_nominative_users_records_postgres_retry"
	TopicProcessNominativeUserRecordsDgraphRetry   = "process_nominative_users_records_dgraph_retry"
)

var (
	upserNomUsers                               = TopicUpsertNominativeUsers
	upserNomUsersRetry                          = TopicUpsertNominativeUsersRetry
	upsertNominativeUsersPostgresRetry          = TopicUpsertNominativeUsersPostgresRetry
	upsertNominativeUsersDgraphRetry            = TopicUpsertNominativeUsersDgraphRetry
	updateNomUserReqRetry                       = TopicUpdateNominativeUserRequestRetry
	updateNomUserReq                            = TopicUpdateNominativeUserRequest
	upsertNominativeUsersPostgres               = TopicUpsertNominativeUsersPostgres
	upsertNominativeUsersDgraph                 = TopicUpsertNominativeUsersDgraph
	processUpsertNominativeUsersPostgres        = TopicProcessNominativeUserRecordsPostgresRetry
	processUpsertNominativeUsersDgraph          = TopicProcessNominativeUserRecordsDgraphRetry
	topicProcessNomPostgresSuccess              = TopicProcessNomPostgresSuccess
	topicProcessNomPostgresSuccessRetry         = TopicProcessNomPostgresSuccessRetry
	topicProcessNomDgraphSuccess                = TopicProcessNomDgraphSuccess
	topicProcessNomDgraphSuccessRetry           = TopicProcessNomDgraphSuccessRetry
	listTopics                                  = []string{upserNomUsers, upserNomUsersRetry, upsertNominativeUsersPostgresRetry, upsertNominativeUsersDgraphRetry, updateNomUserReqRetry, upsertNominativeUsersPostgres, upsertNominativeUsersDgraph, processUpsertNominativeUsersPostgres, processUpsertNominativeUsersDgraph, updateNomUserReq, TopicProcessNomDgraphSuccessRetry}
	upserNomUsersChanSig                        = make(chan os.Signal, 1)
	upserNomUsersRetryChanSig                   = make(chan os.Signal, 1)
	upsertNominativeUsersPostgresRetryChanSig   = make(chan os.Signal, 1)
	upsertNominativeUsersDgraphRetryChanSig     = make(chan os.Signal, 1)
	updateNomUserReqRetryChanSig                = make(chan os.Signal, 1)
	upsertNominativeUsersPostgresChanSig        = make(chan os.Signal, 1)
	processUpsertNominativeUsersDgraphChanSig   = make(chan os.Signal, 1)
	upsertNominativeUsersDgraphChanSig          = make(chan os.Signal, 1)
	processUpsertNominativeUsersPostgresChanSig = make(chan os.Signal, 1)
)

func invokeConsumerUpserNomUsers(topic string, c *kafka.Consumer, ch chan os.Signal, productServer *v1Svc.ProductServiceServer) error {
	err := c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		logger.Log.Sugar().Error("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	logger.Log.Sugar().Info("consumer opened for topic %v", topic)
	// Set up a channel for handling Ctrl-C, etc
	//sigchan := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		run := true
		for run {
			select {
			case sig := <-ch:
				logger.Log.Sugar().Debug("Caught signal %v: terminating\n", sig)
				run = false
			default:
				message, err := c.ReadMessage(100 * time.Millisecond)
				if err != nil {
					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
					continue
				}
				switch *message.TopicPartition.Topic {
				case string(topic):
					handleUpsertNominativeUsersRequest(*message, productServer)
				default:
					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}

			}
		}
	}()
	return nil
}

func invokeConsumerprocessUpsertNominativeUsersDgraph(topic string, c *kafka.Consumer, ch chan os.Signal, productServer *v1Svc.ProductServiceServer) error {
	err := c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		logger.Log.Sugar().Error("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	logger.Log.Sugar().Info("consumer opened for topic %v", topic)

	// Set up a channel for handling Ctrl-C, etc
	//sigchan := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		run := true
		for run {
			select {
			case sig := <-ch:
				logger.Log.Sugar().Debug("Caught signal %v: terminating\n", sig)
				run = false
			default:
				message, err := c.ReadMessage(100 * time.Millisecond)
				if err != nil {
					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
					continue
				}
				switch *message.TopicPartition.Topic {
				case string(topic):
					processNominativeDgraphBatch(*message, productServer)
				default:
					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}

			}
		}
	}()
	return nil
}
func invokeConsumerUpsertNominativeUsersPostgresRetry(topic string, c *kafka.Consumer, ch chan os.Signal, productServer *v1Svc.ProductServiceServer) error {
	err := c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		logger.Log.Sugar().Error("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	logger.Log.Sugar().Info("consumer opened for topic %v", topic)

	// Set up a channel for handling Ctrl-C, etc
	//sigchan := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		run := true
		for run {
			select {
			case sig := <-ch:
				logger.Log.Sugar().Debug("Caught signal %v: terminating\n", sig)
				run = false
			default:
				message, err := c.ReadMessage(100 * time.Millisecond)
				if err != nil {
					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
					continue
				}
				switch *message.TopicPartition.Topic {
				case string(topic):
					handlePostgresNominativeUserRequest(*message, productServer, &sync.WaitGroup{})
				default:
					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}

			}
		}
	}()
	return nil
}
func invokeConsumerUpsertNominativeUsersDgraphRetry(topic string, c *kafka.Consumer, ch chan os.Signal, productServer *v1Svc.ProductServiceServer) error {
	err := c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		logger.Log.Sugar().Error("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	logger.Log.Sugar().Info("consumer opened for topic %v", topic)

	// Set up a channel for handling Ctrl-C, etc
	//sigchan := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		run := true
		for run {
			select {
			case sig := <-ch:
				logger.Log.Sugar().Debug("Caught signal %v: terminating\n", sig)
				run = false
			default:
				message, err := c.ReadMessage(100 * time.Millisecond)
				if err != nil {
					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
					continue
				}
				switch *message.TopicPartition.Topic {
				case string(topic):
					handleDgraphNominativeUserRequestRetry(*message, productServer)
				default:
					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}

			}
		}
	}()
	return nil
}
func invokeConsumerUpdateNomUserReqRetry(topic string, c *kafka.Consumer, ch chan os.Signal, productServer *v1Svc.ProductServiceServer) error {
	err := c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		logger.Log.Sugar().Error("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	logger.Log.Sugar().Info("consumer opened for topic %v", topic)

	// Set up a channel for handling Ctrl-C, etc
	//sigchan := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		run := true
		for run {
			select {
			case sig := <-ch:
				logger.Log.Sugar().Debug("Caught signal %v: terminating\n", sig)
				run = false
			default:
				message, err := c.ReadMessage(100 * time.Millisecond)
				if err != nil {
					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
					continue
				}
				switch *message.TopicPartition.Topic {
				case string(topic):
					handleUpdateNominativeUserRequestRetry(*message, productServer)
				default:
					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}

			}
		}
	}()
	return nil
}
func invokeConsumerUpsertNominativeUsersPostgres(topic string, c *kafka.Consumer, ch chan os.Signal, productServer *v1Svc.ProductServiceServer) error {
	err := c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		logger.Log.Sugar().Error("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	logger.Log.Sugar().Info("consumer opened for topic %v", topic)

	// Set up a channel for handling Ctrl-C, etc
	//sigchan := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		run := true
		for run {
			select {
			case sig := <-ch:
				logger.Log.Sugar().Debug("Caught signal %v: terminating\n", sig)
				run = false
			default:
				message, err := c.ReadMessage(100 * time.Millisecond)
				if err != nil {
					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
					continue
				}
				switch *message.TopicPartition.Topic {
				case string(topic):
					processPostgresNominativeUserUpsert(*message, productServer)
				default:
					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}

			}
		}
	}()
	return nil
}
func invokeConsumerUpsertNominativeUsersDgraph(topic string, c *kafka.Consumer, ch chan os.Signal, productServer *v1Svc.ProductServiceServer) error {
	err := c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		logger.Log.Sugar().Error("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	logger.Log.Sugar().Info("consumer opened for topic %v", topic)

	// Set up a channel for handling Ctrl-C, etc
	//sigchan := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		run := true
		for run {
			select {
			case sig := <-ch:
				logger.Log.Sugar().Debug("Caught signal %v: terminating\n", sig)
				run = false
			default:
				message, err := c.ReadMessage(100 * time.Millisecond)
				if err != nil {
					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
					continue
				}
				switch *message.TopicPartition.Topic {
				case string(topic):
					processNominativeDgraphBatch(*message, productServer)
				default:
					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}

			}
		}
	}()
	return nil
}
func invokeConsumerProcessUpsertNominativeUsersPostgres(topic string, c *kafka.Consumer, ch chan os.Signal, productServer *v1Svc.ProductServiceServer) error {
	err := c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		logger.Log.Sugar().Error("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	logger.Log.Sugar().Info("consumer opened for topic %v", topic)

	// Set up a channel for handling Ctrl-C, etc
	//sigchan := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		run := true
		for run {
			select {
			case sig := <-ch:
				logger.Log.Sugar().Debug("Caught signal %v: terminating\n", sig)
				run = false
			default:
				message, err := c.ReadMessage(100 * time.Millisecond)
				if err != nil {
					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
					continue
				}
				switch *message.TopicPartition.Topic {
				case string(topic):
					processPostgresNominativeUserUpsert(*message, productServer)
				default:
					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}

			}
		}
	}()
	return nil
}
func ProductConsumer(c *kafka.Consumer, productServer *v1Svc.ProductServiceServer) error {
	for _, v := range listTopics {
		c, err := kafkaConnect.BuildConsumer(productServer.Cfg.Kafka, map[string]string{
			"fetch.message.max.bytes": "120971520",
		})
		if err != nil {
			logger.Log.Sugar().Debug("failed to open consumer: %v", err)
			return fmt.Errorf("failed to open consumer: %v", err)
		}
		switch {
		case upserNomUsers == v:
			invokeConsumerUpserNomUsers(upserNomUsers, c, upserNomUsersChanSig, productServer)
		case upserNomUsersRetry == v:
			invokeConsumerUpserNomUsers(upserNomUsersRetry, c, upserNomUsersRetryChanSig, productServer)
		case upsertNominativeUsersPostgresRetry == v:
			invokeConsumerUpsertNominativeUsersPostgresRetry(upsertNominativeUsersPostgresRetry, c, upsertNominativeUsersPostgresRetryChanSig, productServer)
		case upsertNominativeUsersDgraphRetry == v:
			invokeConsumerUpsertNominativeUsersDgraphRetry(upsertNominativeUsersDgraphRetry, c, upsertNominativeUsersDgraphRetryChanSig, productServer)
		case updateNomUserReqRetry == v:
			invokeConsumerUpdateNomUserReqRetry(updateNomUserReqRetry, c, updateNomUserReqRetryChanSig, productServer)
			//postgres processing
		case upsertNominativeUsersPostgres == v:
			invokeConsumerUpsertNominativeUsersPostgres(upsertNominativeUsersPostgres, c, upsertNominativeUsersPostgresChanSig, productServer)
		//postgres retry
		case processUpsertNominativeUsersPostgres == v:
			invokeConsumerProcessUpsertNominativeUsersPostgres(processUpsertNominativeUsersPostgres, c, processUpsertNominativeUsersPostgresChanSig, productServer)
		//dgraph batch processing
		case upsertNominativeUsersDgraph == v:
			invokeConsumerUpsertNominativeUsersDgraph(upsertNominativeUsersDgraph, c, upsertNominativeUsersDgraphChanSig, productServer)
		//dgraph batch processing retry
		case processUpsertNominativeUsersDgraph == v:
			invokeConsumerprocessUpsertNominativeUsersDgraph(processUpsertNominativeUsersDgraph, c, processUpsertNominativeUsersDgraphChanSig, productServer)
		}
	}
	return nil
}
func getNoOfRetriesFromHeader(headers []kafka.Header) int {
	for _, v := range headers {
		if v.Key == NoOfRetries {
			noretries, _ := strconv.Atoi(string(v.Value))
			return noretries
		}
	}
	return 0
}

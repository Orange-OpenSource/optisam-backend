package v1

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkaConnect "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/kafka"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	v1Svc "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/service/v1"
)

const (
	TopicUpsertNominativeUsers            = "upsert_nominative_users"
	TopicUpsertNominativeUsersRetry       = "upsert_nominative_users_retry"
	TopicUpdateNominativeUserRequest      = "update_nominative_user_request"
	TopicUpdateNominativeUserRequestRetry = "update_nominative_user_request_retry"
	//TopicProcessNominativeUserRequestRetry = "process_nominative_user_request_retry"
	TopicProcessNomPostgresSuccess             = "process_nom_postgres_success"
	TopicProcessNomPostgresSuccessRetry        = "process_nom_postgres_success_retry"
	TopicProcessNomDgraphSuccess               = "process_nom_dgraph_success"
	TopicProcessNomDgraphSuccessRetry          = "process_nom_dgraph_success_retry"
	TopicCDCNominativeUserRequestsSuccess      = "public.import_connector.public.nominative_user_requests"
	TopicCDCNominativeUserRequestsSuccessRetry = "public.import_connector.public.nominative_user_requests_retry"
	RetryDelay                                 = 30
	TopicDeadLetterQueue                       = "dead_letter_queue"
)

var (
	updateNominativeUserReq = TopicUpdateNominativeUserRequest
	//updateNominativeUserReqRetry          = TopicUpdateNominativeUserRequestRetry
	topicProcessNomPostgresSuccess                    = TopicProcessNomPostgresSuccess
	topicProcessNomPostgresSuccessRetry               = TopicProcessNomPostgresSuccessRetry
	topicProcessNomDgraphSuccess                      = TopicProcessNomDgraphSuccess
	topicProcessNomDgraphSuccessRetry                 = TopicProcessNomDgraphSuccessRetry
	topicCDCNominativeUserRequestsSuccess             = TopicCDCNominativeUserRequestsSuccess
	topicCDCNominativeUserRequestsSuccessRetry        = TopicCDCNominativeUserRequestsSuccessRetry
	listTopics                                        = []string{updateNominativeUserReq, topicProcessNomPostgresSuccess, topicProcessNomPostgresSuccessRetry, topicProcessNomDgraphSuccess, topicProcessNomDgraphSuccessRetry, topicCDCNominativeUserRequestsSuccess, topicCDCNominativeUserRequestsSuccessRetry}
	updateNominativeUserReqChanSig                    = make(chan os.Signal, 1)
	topicProcessNomPostgresSuccessChanSig             = make(chan os.Signal, 1)
	topicProcessNomPostgresSuccessRetryChanSig        = make(chan os.Signal, 1)
	topicProcessNomDgraphSuccessChanSig               = make(chan os.Signal, 1)
	topicProcessNomDgraphSuccessRetryChanSig          = make(chan os.Signal, 1)
	topicCDCNominativeUserRequestsSuccessChanSig      = make(chan os.Signal, 1)
	topicCDCNominativeUserRequestsSuccessRetryChanSig = make(chan os.Signal, 1)

	//updateNominativeUserReqRetryChanSig = make(chan os.Signal, 1)

	NoOfRetries = "no_of_retries"
)

func ImportConsumer(i v1Svc.ImportServiceServer) error {
	for _, v := range listTopics {
		c, err := kafkaConnect.BuildConsumer(i.Config.Kafka, map[string]string{})
		if err != nil {
			logger.Log.Sugar().Debug("failed to open consumer: %v", err)
			return fmt.Errorf("failed to open consumer: %v", err)
		}
		switch {
		case v == updateNominativeUserReq:
			invokeConsumerUpdateNominativeUserReq(updateNominativeUserReq, c, updateNominativeUserReqChanSig, &i)
		case v == topicProcessNomPostgresSuccess:
			invokeConsumerProcessNomPostgresSuccess(topicProcessNomPostgresSuccess, c, topicProcessNomPostgresSuccessChanSig, &i)
		case v == topicProcessNomPostgresSuccessRetry:
			invokeConsumerProcessNomPostgresSuccess(topicProcessNomPostgresSuccessRetry, c, topicProcessNomPostgresSuccessRetryChanSig, &i)
		case v == topicProcessNomDgraphSuccess:
			invokeConsumerProcessNomDgraphSuccess(topicProcessNomDgraphSuccess, c, topicProcessNomDgraphSuccessChanSig, &i)
		case v == topicProcessNomDgraphSuccessRetry:
			invokeConsumerProcessNomDgraphSuccess(topicProcessNomDgraphSuccessRetry, c, topicProcessNomDgraphSuccessRetryChanSig, &i)
			// case v == topicCDCNominativeUserRequestsSuccess:
			// 	invokeConsumerCDCNominativeUserRequests(topicCDCNominativeUserRequestsSuccess, c, topicCDCNominativeUserRequestsSuccessChanSig, &i)
			// case v == topicCDCNominativeUserRequestsSuccessRetry:
			// 	invokeConsumerCDCNominativeUserRequestsRetry(topicCDCNominativeUserRequestsSuccessRetry, c, topicCDCNominativeUserRequestsSuccessRetryChanSig, &i)
		}
	}
	return nil
}
func invokeConsumerUpdateNominativeUserReq(topic string, c *kafka.Consumer, ch chan os.Signal, importServer *v1Svc.ImportServiceServer) error {
	err := c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		logger.Log.Sugar().Error("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	// Set up a channel for handling Ctrl-C, etc
	logger.Log.Sugar().Info("Consumer opened for topic : %v", topic)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		run := true
		for run {
			select {
			case sig := <-ch:
				logger.Log.Sugar().Info("Caught signal %v: terminating\n", sig)
				run = false
			default:
				message, err := c.ReadMessage(100 * time.Millisecond)
				//logger.Log.Sugar().Info("called for process nom users")
				if err != nil {
					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
					continue
				}

				switch *message.TopicPartition.Topic {
				case updateNominativeUserReq:
					processNominativeUserReq(*message, importServer)
				default:
					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}
			}
		}
	}()
	return nil
}

func invokeConsumerProcessNomPostgresSuccess(topic string, c *kafka.Consumer, ch chan os.Signal, importServer *v1Svc.ImportServiceServer) error {

	err := c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		logger.Log.Sugar().Error("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	// Set up a channel for handling Ctrl-C, etc
	logger.Log.Sugar().Info("Consumer opened for topic : %v", topic)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		run := true
		for run {
			select {
			case sig := <-ch:
				logger.Log.Sugar().Debug("Caught signal %v: terminating\n", sig)
				run = false
			default:
				message, err := c.ReadMessage(100 * time.Microsecond)
				if err != nil {
					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
					continue
				}

				switch *message.TopicPartition.Topic {
				case string(topic):
					processNomPostgresSuccess(*message, importServer)
				default:
					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}
			}
		}
	}()
	return nil
}

func invokeConsumerProcessNomDgraphSuccess(topic string, c *kafka.Consumer, ch chan os.Signal, importServer *v1Svc.ImportServiceServer) error {

	err := c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		logger.Log.Sugar().Error("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	// Set up a channel for handling Ctrl-C, etc
	logger.Log.Sugar().Info("Consumer opened for topic : %v", topic)
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
					processNomDgraphSuccess(*message, importServer)
				default:
					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}
			}
		}
	}()
	return nil
}

// func invokeConsumerCDCNominativeUserRequests(topic string, c *kafka.Consumer, ch chan os.Signal, importServer *v1Svc.ImportServiceServer) error {

// 	err := c.SubscribeTopics([]string{topic}, nil)
// 	if err != nil {
// 		logger.Log.Sugar().Error("failed to open consumer: %v", err)
// 		return fmt.Errorf("failed to open consumer: %v", err)
// 	}
// 	// Set up a channel for handling Ctrl-C, etc
// 	logger.Log.Sugar().Info("Consumer opened for topic : %v", topic)
// 	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
// 	go func() {
// 		run := true
// 		for run {
// 			select {
// 			case sig := <-ch:
// 				logger.Log.Sugar().Debug("Caught signal %v: terminating\n", sig)
// 				run = false
// 			default:
// 				message, err := c.ReadMessage(100 * time.Millisecond)
// 				if err != nil {
// 					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
// 					continue
// 				}

// 				switch *message.TopicPartition.Topic {
// 				case string(topic):
// 					processCDCNominativeUserRequests(*message, importServer)
// 				default:
// 					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
// 						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
// 				}
// 			}
// 		}
// 	}()
// 	return nil
// }

// func invokeConsumerCDCNominativeUserRequestsRetry(topic string, c *kafka.Consumer, ch chan os.Signal, importServer *v1Svc.ImportServiceServer) error {

// 	err := c.SubscribeTopics([]string{topic}, nil)
// 	if err != nil {
// 		logger.Log.Sugar().Error("failed to open consumer: %v", err)
// 		return fmt.Errorf("failed to open consumer: %v", err)
// 	}
// 	// Set up a channel for handling Ctrl-C, etc
// 	logger.Log.Sugar().Info("Consumer opened for topic : %v", topic)
// 	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
// 	go func() {
// 		run := true
// 		for run {
// 			select {
// 			case sig := <-ch:
// 				logger.Log.Sugar().Debug("Caught signal %v: terminating\n", sig)
// 				run = false
// 			default:
// 				message, err := c.ReadMessage(100 * time.Millisecond)
// 				if err != nil {
// 					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
// 					continue
// 				}

// 				switch *message.TopicPartition.Topic {
// 				case string(topic):
// 					processCDCNominativeUserRequests(*message, importServer)
// 				default:
// 					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
// 						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
// 				}
// 			}
// 		}
// 	}()
// 	return nil
// }

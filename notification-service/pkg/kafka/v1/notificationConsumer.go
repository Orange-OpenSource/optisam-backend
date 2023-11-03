package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/repository/v1/postgres/db"

	//v1Kafka "gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/kafka/v1"
	kafkaConnect "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/kafka"
	v1Svc "gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/service/v1"
)

const (
	TopicEmailNotification      = "email_notification"
	TopicEmailNotificationRetry = "email_notification_retry"
	TopicDeadLetterQueue        = "dead_letter_queue"
)

var (
	NoOfRetries            = "no_of_retries"
	emailNotification      = TopicEmailNotification
	emailNotificationRetry = TopicEmailNotificationRetry
	deadLetterQueue        = TopicDeadLetterQueue
	listTopics             = []string{TopicDeadLetterQueue, TopicEmailNotification}
	incidentSubject        = "OPTISAM: This is the incident message"
	incidentMessage        = `
	<!DOCTYPE html>
	<html>
	<body>
		<h1>Incident Report</h1>
		<p><strong>Topic:</strong> {{topic}}</p>
		<p><strong>Number of Retries:</strong> {{no_of_retries}}</p>
		<p><strong>Message:</strong> {{message}}</p>
		<h1>Rapport d'Incident</h1>
        <p><strong>Sujet :</strong> {{topic}}</p>
        <p><strong>Nombre de Tentatives :</strong> {{no_of_retries}}</p>
        <p><strong>Message :</strong> {{message}}</p>
	</body>
	</html>
`
)

func NotificationConsumer(notificationServer *v1Svc.NotificationServiceServer) error {
	for _, v := range listTopics {
		c, err := kafkaConnect.BuildConsumer(notificationServer.Config.Kafka, map[string]string{})
		if err != nil {
			logger.Log.Sugar().Debug("failed to open consumer: %v", err)
			return fmt.Errorf("failed to open consumer: %v", err)
		}
		switch {
		case v == deadLetterQueue:
			DeadLetterConsumer(c, notificationServer)
		case v == emailNotification:
			EmailNotificationConsumer(c, notificationServer)
		}
	}
	return nil
}

func DeadLetterConsumer(c *kafka.Consumer, notificationServer *v1Svc.NotificationServiceServer) error {

	err := c.SubscribeTopics([]string{deadLetterQueue}, nil)
	if err != nil {
		logger.Log.Sugar().Error("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	// Set up a channel for handling Ctrl-C, etc
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		run := true
		for run {
			select {
			case sig := <-sigchan:
				logger.Log.Sugar().Debug("Caught signal %v: terminating\n", sig)
				run = false
			default:
				message, err := c.ReadMessage(1000 * time.Millisecond)
				if err != nil {
					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
					continue
				}
				switch *message.TopicPartition.Topic {
				case string(deadLetterQueue):
					processDeadLetterQueue(*message, notificationServer)
				default:
					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}

			}
		}
	}()
	return nil
}

func EmailNotificationConsumer(c *kafka.Consumer, notificationServer *v1Svc.NotificationServiceServer) error {

	err := c.SubscribeTopics([]string{emailNotification, emailNotificationRetry}, nil)
	if err != nil {
		logger.Log.Sugar().Error("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	// Set up a channel for handling Ctrl-C, etc
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		run := true
		for run {
			select {
			case sig := <-sigchan:
				logger.Log.Sugar().Debug("Caught signal %v: terminating\n", sig)
				run = false
			default:
				message, err := c.ReadMessage(100 * time.Millisecond)
				if err != nil {
					//logger.Log.Sugar().Error("error reading messages from consumer %v\n", err)
					continue
				}
				switch *message.TopicPartition.Topic {
				case string(emailNotification):
					processEmailNotification(*message, notificationServer)
				case string(emailNotificationRetry):
					processEmailNotification(*message, notificationServer)
				default:
					fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
						*message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}

			}
		}
	}()
	return nil
}

func processDeadLetterQueue(message kafka.Message, notificationServer *v1Svc.NotificationServiceServer) {
	noOfRetries := getNoOfRetriesFromHeader(message.Headers)
	// if noOfRetries > 20 {
	// 	produceToDLQ(notificationServer, message, emailNotificationRetry)
	// 	return
	// }
	// if noOfRetries > 0 {
	// 	time.Sleep(time.Second * 30)
	// }
	//noOfRetries = noOfRetries + 1

	err := notificationServer.NotificationRepo.PublishToDLQ(context.Background(), db.PublishToDLQParams{Topic: *message.TopicPartition.Topic, NoOfRetries: int32(noOfRetries), Message: message.Value})

	if err != nil {
		logger.Log.Sugar().Error("error  inserting data to dead letter queue%v\n", err)
		handleError(notificationServer, &deadLetterQueue, message.Value, noOfRetries, err)
	}

	incidentMessage = strings.Replace(incidentMessage, "{{topic}}", *message.TopicPartition.Topic, -1)
	incidentMessage = strings.Replace(incidentMessage, "{{no_of_retries}}", strconv.Itoa(noOfRetries), -1)
	incidentMessage = strings.Replace(incidentMessage, "{{message}}", string(message.Value), -1)

	req := v1.SendMailRequest{
		MailSubject: incidentSubject,
		MailMessage: incidentMessage,
		MailTo:      []string{"optisam_india_team@easymail.orange.com"},
	}
	_, err = notificationServer.SendMail(context.Background(), &req)
	if err != nil {
		logger.Log.Sugar().Error("error recived from SendMail %v\n", err)
		handleError(notificationServer, &deadLetterQueue, message.Value, noOfRetries, err)
	} else {
		logger.Log.Sugar().Info("mail processed\n")
	}
}

func processEmailNotification(message kafka.Message, notificationServer *v1Svc.NotificationServiceServer) {
	noOfRetries := getNoOfRetriesFromHeader(message.Headers)
	if noOfRetries > 20 {
		produceToDLQ(notificationServer, message, emailNotificationRetry)
		return
	}
	if noOfRetries > 0 {
		time.Sleep(time.Second * 30)
	}
	noOfRetries = noOfRetries + 1
	req := v1.SendMailRequest{}
	err := json.Unmarshal(message.Value, &req)
	if err != nil {
		logger.Log.Sugar().Error("error unmarshaling SendMailRequest %v\n", err)
		handleError(notificationServer, &emailNotificationRetry, message.Value, noOfRetries, err)
	}
	_, err = notificationServer.SendMail(context.Background(), &req)
	if err != nil {
		logger.Log.Sugar().Error("error recived from SendMail %v\n", err)
		handleError(notificationServer, &emailNotificationRetry, message.Value, noOfRetries, err)
	} else {
		logger.Log.Sugar().Info("mail processed\n")
	}
}

// func handleEmailNotificationError(message kafka.Message, p *kafka.Producer) {
// 	emailNotificationRetry := TopicEmailNotificationRetry
// 	if req.NoOfRetiries < 10 {
// 		req.NoOfRetiries = 1 + req.NoOfRetiries
// 		notificationReq, _ := json.Marshal(req)
// 		err := p.Produce(&kafka.Message{
// 			TopicPartition: kafka.TopicPartition{Topic: &emailNotificationRetry, Partition: kafka.PartitionAny},
// 			Value:          []byte(notificationReq),
// 		}, nil)
// 		if err != nil {
// 			logger.Log.Sugar().Errorw("notification service - forgot password - handleEmailNotificationError - error producing retry event" + err.Error())
// 		} else {
// 			logger.Log.Sugar().Debug("successfully produced event to email_notification_retry")
// 		}
// 	}
// }

func handleError(notificationServer *v1Svc.NotificationServiceServer, topic *string, value []byte, noOfRetries int, err error) {
	notificationServer.KafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: topic, Partition: kafka.PartitionAny},
		Value:          value,
		Headers:        []kafka.Header{{Key: NoOfRetries, Value: []byte(strconv.Itoa(noOfRetries))}, {Key: "error", Value: []byte(err.Error())}},
	}, nil)
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

func produceToDLQ(notificationServer *v1Svc.NotificationServiceServer, msg kafka.Message, topic string) {
	err := ""
	for _, v := range msg.Headers {
		if v.Key == "error" {
			err = string(v.Value)
		}
	}
	dql := v1.DeadLetterQueue{
		TopicName: topic,
		Error:     err,
		Message:   fmt.Sprint(msg),
	}
	d, _ := json.Marshal(&dql)
	topicDQL := TopicDeadLetterQueue
	notificationServer.KafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topicDQL,
			//Partition: rand.Int31n(importServer.Config.NoOfPartitions)},
			Partition: kafka.PartitionAny},
		Value: d,
	}, nil)
}

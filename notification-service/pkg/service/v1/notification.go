package v1

import (
	"context"
	"crypto/tls"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/config"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/repository/v1"
	gomail "gopkg.in/mail.v2"
)

// productServiceServer is implementation of v1.authServiceServer proto interface
type NotificationServiceServer struct {
	Config           *config.Config
	KafkaProducer    *kafka.Producer
	NotificationRepo repo.Notification
	GomailDialer     *gomail.Dialer
}

// NewAccountServiceServer creates Auth service
func NewNotificationServiceServer(cfg *config.Config, kafkaProducer *kafka.Producer, NotificationRepo repo.Notification, gomailDialer *gomail.Dialer) *NotificationServiceServer {
	return &NotificationServiceServer{
		Config:           cfg,
		KafkaProducer:    kafkaProducer,
		NotificationRepo: NotificationRepo,
		GomailDialer:     gomailDialer,
	}
}

// Create Product
func (n *NotificationServiceServer) SendMail(ctx context.Context, req *v1.SendMailRequest) (res *v1.SendMailResponse, err error) {
	// logger.("Start ")
	logger.Log.Sugar().Infof("Notification Service/sendMail Starting")

	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetHeader("From", n.Config.SMTP.Email)
	// Set E-Mail receivers
	m.SetHeader("To", req.MailTo...)
	// Set E-Mail subject
	m.SetHeader("Subject", req.MailSubject)
	// Set E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/html", req.GetMailMessage())

	// Settings for SMTP server
	d := gomail.NewDialer(n.Config.SMTP.Host, int(n.Config.SMTP.Port), n.Config.SMTP.From, n.Config.SMTP.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	logger.Log.Sugar().Infof("Notification Service/sendMail-From,Host,Port,Password", n.Config.SMTP.From,
		n.Config.SMTP.Host, int(n.Config.SMTP.Port), n.Config.SMTP.Password)

	// Now send E-Mail
	if err := n.GomailDialer.DialAndSend(m); err != nil {
		logger.Log.Sugar().Errorf("Notification Service/sendMail failed while sending mail", err.Error())
		return &v1.SendMailResponse{Success: "false"}, err
	}
	logger.Log.Sugar().Infof("Notification Service/sendMail Ending")

	return &v1.SendMailResponse{Success: "true"}, nil
}

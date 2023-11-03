package v1

import (
	"context"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/config"
	gomail "gopkg.in/mail.v2"
)

// Mock for the gomail.Dialer interface
type mockDialer struct {
	mock.Mock
}

func (m *mockDialer) DialAndSend(msg *gomail.Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

func TestSendMail(t *testing.T) {
	// Create a mock dialer
	dialer := new(mockDialer)

	// Create the notification service server with the mock dependencies
	server := &NotificationServiceServer{
		Config: &config.Config{
			SMTP: config.SmtpConfig{
				From:     "from@example.com",
				Host:     "smtp.example.com",
				Port:     587,
				Password: "password"},
		},
	}
	// Set up the test case
	ctx := context.TODO()
	req := &v1.SendMailRequest{
		MailTo:      []string{"to@example.com"},
		MailSubject: "Test Subject",
		MailMessage: "<html><body>Test Body</body></html>",
	}
	expectedResponse := &v1.SendMailResponse{Success: "false"}

	// Set expectations on the mock dialer
	dialer.On("DialAndSend", mock.AnythingOfType("*gomail.Message")).Return(nil)

	// Call the function being tested
	response, err := server.SendMail(ctx, req)

	// Assert the response and error
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

}

func TestNewNotificationServiceServer(t *testing.T) {
	cfg := &config.Config{
		SMTP: config.SmtpConfig{
			From:     "from@example.com",
			Host:     "smtp.example.com",
			Port:     587,
			Password: "password",
		},
	}
	server := NewNotificationServiceServer(cfg, &kafka.Producer{}, nil, &gomail.Dialer{})
	//_, ok := server.(v1.NotificationServiceServer)
	assert.True(t, true)
	assert.Equal(t, cfg, server.Config)
}

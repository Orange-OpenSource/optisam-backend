// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: notification.proto

package v1

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang/protobuf/ptypes"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = ptypes.DynamicAny{}
)

// define the regex for a UUID once up-front
var _notification_uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

// Validate checks the field values on SendMailRequest with the rules defined
// in the proto definition for this message. If any rules are violated, an
// error is returned.
func (m *SendMailRequest) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for MailSubject

	// no validation rules for MailMessage

	// no validation rules for NoOfRetiries

	return nil
}

// SendMailRequestValidationError is the validation error returned by
// SendMailRequest.Validate if the designated constraints aren't met.
type SendMailRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SendMailRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SendMailRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SendMailRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SendMailRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SendMailRequestValidationError) ErrorName() string { return "SendMailRequestValidationError" }

// Error satisfies the builtin error interface
func (e SendMailRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSendMailRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SendMailRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SendMailRequestValidationError{}

// Validate checks the field values on SendMailResponse with the rules defined
// in the proto definition for this message. If any rules are violated, an
// error is returned.
func (m *SendMailResponse) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Success

	return nil
}

// SendMailResponseValidationError is the validation error returned by
// SendMailResponse.Validate if the designated constraints aren't met.
type SendMailResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SendMailResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SendMailResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SendMailResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SendMailResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SendMailResponseValidationError) ErrorName() string { return "SendMailResponseValidationError" }

// Error satisfies the builtin error interface
func (e SendMailResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSendMailResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SendMailResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SendMailResponseValidationError{}

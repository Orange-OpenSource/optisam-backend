// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: dps.proto

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
var _dps_uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

// Validate checks the field values on NotifyUploadRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *NotifyUploadRequest) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Scope

	if _, ok := _NotifyUploadRequest_Type_InLookup[m.GetType()]; !ok {
		return NotifyUploadRequestValidationError{
			field:  "Type",
			reason: "value must be in list [data metadata]",
		}
	}

	// no validation rules for UploadId

	// no validation rules for UploadedBy

	if len(m.GetFiles()) < 1 {
		return NotifyUploadRequestValidationError{
			field:  "Files",
			reason: "value must contain at least 1 item(s)",
		}
	}

	return nil
}

// NotifyUploadRequestValidationError is the validation error returned by
// NotifyUploadRequest.Validate if the designated constraints aren't met.
type NotifyUploadRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e NotifyUploadRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e NotifyUploadRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e NotifyUploadRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e NotifyUploadRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e NotifyUploadRequestValidationError) ErrorName() string {
	return "NotifyUploadRequestValidationError"
}

// Error satisfies the builtin error interface
func (e NotifyUploadRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sNotifyUploadRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = NotifyUploadRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = NotifyUploadRequestValidationError{}

var _NotifyUploadRequest_Type_InLookup = map[string]struct{}{
	"data":     {},
	"metadata": {},
}

// Validate checks the field values on NotifyUploadResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *NotifyUploadResponse) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Success

	return nil
}

// NotifyUploadResponseValidationError is the validation error returned by
// NotifyUploadResponse.Validate if the designated constraints aren't met.
type NotifyUploadResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e NotifyUploadResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e NotifyUploadResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e NotifyUploadResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e NotifyUploadResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e NotifyUploadResponseValidationError) ErrorName() string {
	return "NotifyUploadResponseValidationError"
}

// Error satisfies the builtin error interface
func (e NotifyUploadResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sNotifyUploadResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = NotifyUploadResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = NotifyUploadResponseValidationError{}

// Validate checks the field values on ListUploadRequest with the rules defined
// in the proto definition for this message. If any rules are violated, an
// error is returned.
func (m *ListUploadRequest) Validate() error {
	if m == nil {
		return nil
	}

	if val := m.GetPageNum(); val < 1 || val >= 1000 {
		return ListUploadRequestValidationError{
			field:  "PageNum",
			reason: "value must be inside range [1, 1000)",
		}
	}

	if m.GetPageSize() < 10 {
		return ListUploadRequestValidationError{
			field:  "PageSize",
			reason: "value must be greater than or equal to 10",
		}
	}

	if _, ok := ListUploadRequest_SortBy_name[int32(m.GetSortBy())]; !ok {
		return ListUploadRequestValidationError{
			field:  "SortBy",
			reason: "value must be one of the defined enum values",
		}
	}

	if _, ok := ListUploadRequest_SortOrder_name[int32(m.GetSortOrder())]; !ok {
		return ListUploadRequestValidationError{
			field:  "SortOrder",
			reason: "value must be one of the defined enum values",
		}
	}

	return nil
}

// ListUploadRequestValidationError is the validation error returned by
// ListUploadRequest.Validate if the designated constraints aren't met.
type ListUploadRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListUploadRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListUploadRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListUploadRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListUploadRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListUploadRequestValidationError) ErrorName() string {
	return "ListUploadRequestValidationError"
}

// Error satisfies the builtin error interface
func (e ListUploadRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListUploadRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListUploadRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListUploadRequestValidationError{}

// Validate checks the field values on ListUploadResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *ListUploadResponse) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for TotalRecords

	for idx, item := range m.GetUploads() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ListUploadResponseValidationError{
					field:  fmt.Sprintf("Uploads[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// ListUploadResponseValidationError is the validation error returned by
// ListUploadResponse.Validate if the designated constraints aren't met.
type ListUploadResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListUploadResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListUploadResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListUploadResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListUploadResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListUploadResponseValidationError) ErrorName() string {
	return "ListUploadResponseValidationError"
}

// Error satisfies the builtin error interface
func (e ListUploadResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListUploadResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListUploadResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListUploadResponseValidationError{}

// Validate checks the field values on Upload with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Upload) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for UploadId

	// no validation rules for Scope

	// no validation rules for FileName

	// no validation rules for Status

	// no validation rules for UploadedBy

	if v, ok := interface{}(m.GetUploadedOn()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return UploadValidationError{
				field:  "UploadedOn",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	// no validation rules for TotalRecords

	// no validation rules for SuccessRecords

	// no validation rules for FailedRecords

	// no validation rules for InvalidRecords

	return nil
}

// UploadValidationError is the validation error returned by Upload.Validate if
// the designated constraints aren't met.
type UploadValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e UploadValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e UploadValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e UploadValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e UploadValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e UploadValidationError) ErrorName() string { return "UploadValidationError" }

// Error satisfies the builtin error interface
func (e UploadValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sUpload.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = UploadValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = UploadValidationError{}

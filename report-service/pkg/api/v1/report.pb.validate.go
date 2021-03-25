// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: report.proto

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
var _report_uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

// Validate checks the field values on ListReportTypeRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *ListReportTypeRequest) Validate() error {
	if m == nil {
		return nil
	}

	return nil
}

// ListReportTypeRequestValidationError is the validation error returned by
// ListReportTypeRequest.Validate if the designated constraints aren't met.
type ListReportTypeRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListReportTypeRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListReportTypeRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListReportTypeRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListReportTypeRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListReportTypeRequestValidationError) ErrorName() string {
	return "ListReportTypeRequestValidationError"
}

// Error satisfies the builtin error interface
func (e ListReportTypeRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListReportTypeRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListReportTypeRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListReportTypeRequestValidationError{}

// Validate checks the field values on ListReportTypeResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *ListReportTypeResponse) Validate() error {
	if m == nil {
		return nil
	}

	for idx, item := range m.GetReportType() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface {
			Validate() error
		}); ok {
			if err := v.Validate(); err != nil {
				return ListReportTypeResponseValidationError{
					field:  fmt.Sprintf("ReportType[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// ListReportTypeResponseValidationError is the validation error returned by
// ListReportTypeResponse.Validate if the designated constraints aren't met.
type ListReportTypeResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListReportTypeResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListReportTypeResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListReportTypeResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListReportTypeResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListReportTypeResponseValidationError) ErrorName() string {
	return "ListReportTypeResponseValidationError"
}

// Error satisfies the builtin error interface
func (e ListReportTypeResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListReportTypeResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListReportTypeResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListReportTypeResponseValidationError{}

// Validate checks the field values on ReportType with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *ReportType) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for ReportTypeId

	// no validation rules for ReportTypeName

	return nil
}

// ReportTypeValidationError is the validation error returned by
// ReportType.Validate if the designated constraints aren't met.
type ReportTypeValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ReportTypeValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ReportTypeValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ReportTypeValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ReportTypeValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ReportTypeValidationError) ErrorName() string { return "ReportTypeValidationError" }

// Error satisfies the builtin error interface
func (e ReportTypeValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sReportType.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ReportTypeValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ReportTypeValidationError{}

// Validate checks the field values on SubmitReportRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *SubmitReportRequest) Validate() error {
	if m == nil {
		return nil
	}

	if !_SubmitReportRequest_Scope_Pattern.MatchString(m.GetScope()) {
		return SubmitReportRequestValidationError{
			field:  "Scope",
			reason: "value does not match regex pattern \"\\\\b[A-Z]{3}\\\\b\"",
		}
	}

	// no validation rules for ReportTypeId

	switch m.ReportMetadata.(type) {

	case *SubmitReportRequest_AcqrightsReport:

		if v, ok := interface{}(m.GetAcqrightsReport()).(interface {
			Validate() error
		}); ok {
			if err := v.Validate(); err != nil {
				return SubmitReportRequestValidationError{
					field:  "AcqrightsReport",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *SubmitReportRequest_ProductEquipmentsReport:

		if v, ok := interface{}(m.GetProductEquipmentsReport()).(interface {
			Validate() error
		}); ok {
			if err := v.Validate(); err != nil {
				return SubmitReportRequestValidationError{
					field:  "ProductEquipmentsReport",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// SubmitReportRequestValidationError is the validation error returned by
// SubmitReportRequest.Validate if the designated constraints aren't met.
type SubmitReportRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SubmitReportRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SubmitReportRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SubmitReportRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SubmitReportRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SubmitReportRequestValidationError) ErrorName() string {
	return "SubmitReportRequestValidationError"
}

// Error satisfies the builtin error interface
func (e SubmitReportRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSubmitReportRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SubmitReportRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SubmitReportRequestValidationError{}

var _SubmitReportRequest_Scope_Pattern = regexp.MustCompile("\\b[A-Z]{3}\\b")

// Validate checks the field values on AcqRightsReport with the rules defined
// in the proto definition for this message. If any rules are violated, an
// error is returned.
func (m *AcqRightsReport) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Editor

	return nil
}

// AcqRightsReportValidationError is the validation error returned by
// AcqRightsReport.Validate if the designated constraints aren't met.
type AcqRightsReportValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e AcqRightsReportValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e AcqRightsReportValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e AcqRightsReportValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e AcqRightsReportValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e AcqRightsReportValidationError) ErrorName() string { return "AcqRightsReportValidationError" }

// Error satisfies the builtin error interface
func (e AcqRightsReportValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sAcqRightsReport.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = AcqRightsReportValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = AcqRightsReportValidationError{}

// Validate checks the field values on ProductEquipmentsReport with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *ProductEquipmentsReport) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Editor

	// no validation rules for EquipType

	return nil
}

// ProductEquipmentsReportValidationError is the validation error returned by
// ProductEquipmentsReport.Validate if the designated constraints aren't met.
type ProductEquipmentsReportValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ProductEquipmentsReportValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ProductEquipmentsReportValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ProductEquipmentsReportValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ProductEquipmentsReportValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ProductEquipmentsReportValidationError) ErrorName() string {
	return "ProductEquipmentsReportValidationError"
}

// Error satisfies the builtin error interface
func (e ProductEquipmentsReportValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sProductEquipmentsReport.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ProductEquipmentsReportValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ProductEquipmentsReportValidationError{}

// Validate checks the field values on SubmitReportResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *SubmitReportResponse) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Success

	return nil
}

// SubmitReportResponseValidationError is the validation error returned by
// SubmitReportResponse.Validate if the designated constraints aren't met.
type SubmitReportResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SubmitReportResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SubmitReportResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SubmitReportResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SubmitReportResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SubmitReportResponseValidationError) ErrorName() string {
	return "SubmitReportResponseValidationError"
}

// Error satisfies the builtin error interface
func (e SubmitReportResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSubmitReportResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SubmitReportResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SubmitReportResponseValidationError{}

// Validate checks the field values on ListReportRequest with the rules defined
// in the proto definition for this message. If any rules are violated, an
// error is returned.
func (m *ListReportRequest) Validate() error {
	if m == nil {
		return nil
	}

	if val := m.GetPageNum(); val < 1 || val >= 1000 {
		return ListReportRequestValidationError{
			field:  "PageNum",
			reason: "value must be inside range [1, 1000)",
		}
	}

	if m.GetPageSize() < 10 {
		return ListReportRequestValidationError{
			field:  "PageSize",
			reason: "value must be greater than or equal to 10",
		}
	}

	if _, ok := _ListReportRequest_SortBy_InLookup[m.GetSortBy()]; !ok {
		return ListReportRequestValidationError{
			field:  "SortBy",
			reason: "value must be in list [report_id report_type report_status created_by created_on]",
		}
	}

	if _, ok := SortOrder_name[int32(m.GetSortOrder())]; !ok {
		return ListReportRequestValidationError{
			field:  "SortOrder",
			reason: "value must be one of the defined enum values",
		}
	}

	if !_ListReportRequest_Scope_Pattern.MatchString(m.GetScope()) {
		return ListReportRequestValidationError{
			field:  "Scope",
			reason: "value does not match regex pattern \"\\\\b[A-Z]{3}\\\\b\"",
		}
	}

	return nil
}

// ListReportRequestValidationError is the validation error returned by
// ListReportRequest.Validate if the designated constraints aren't met.
type ListReportRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListReportRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListReportRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListReportRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListReportRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListReportRequestValidationError) ErrorName() string {
	return "ListReportRequestValidationError"
}

// Error satisfies the builtin error interface
func (e ListReportRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListReportRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListReportRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListReportRequestValidationError{}

var _ListReportRequest_SortBy_InLookup = map[string]struct{}{
	"report_id":     {},
	"report_type":   {},
	"report_status": {},
	"created_by":    {},
	"created_on":    {},
}

var _ListReportRequest_Scope_Pattern = regexp.MustCompile("\\b[A-Z]{3}\\b")

// Validate checks the field values on ListReportResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *ListReportResponse) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for TotalRecords

	for idx, item := range m.GetReports() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface {
			Validate() error
		}); ok {
			if err := v.Validate(); err != nil {
				return ListReportResponseValidationError{
					field:  fmt.Sprintf("Reports[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// ListReportResponseValidationError is the validation error returned by
// ListReportResponse.Validate if the designated constraints aren't met.
type ListReportResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListReportResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListReportResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListReportResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListReportResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListReportResponseValidationError) ErrorName() string {
	return "ListReportResponseValidationError"
}

// Error satisfies the builtin error interface
func (e ListReportResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListReportResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListReportResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListReportResponseValidationError{}

// Validate checks the field values on Report with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Report) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for ReportId

	// no validation rules for ReportType

	// no validation rules for ReportStatus

	// no validation rules for CreatedBy

	if v, ok := interface{}(m.GetCreatedOn()).(interface {
		Validate() error
	}); ok {
		if err := v.Validate(); err != nil {
			return ReportValidationError{
				field:  "CreatedOn",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// ReportValidationError is the validation error returned by Report.Validate if
// the designated constraints aren't met.
type ReportValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ReportValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ReportValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ReportValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ReportValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ReportValidationError) ErrorName() string { return "ReportValidationError" }

// Error satisfies the builtin error interface
func (e ReportValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sReport.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ReportValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ReportValidationError{}

// Validate checks the field values on DownloadReportRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *DownloadReportRequest) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for ReportID

	if !_DownloadReportRequest_Scope_Pattern.MatchString(m.GetScope()) {
		return DownloadReportRequestValidationError{
			field:  "Scope",
			reason: "value does not match regex pattern \"\\\\b[A-Z]{3}\\\\b\"",
		}
	}

	return nil
}

// DownloadReportRequestValidationError is the validation error returned by
// DownloadReportRequest.Validate if the designated constraints aren't met.
type DownloadReportRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DownloadReportRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DownloadReportRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DownloadReportRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DownloadReportRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DownloadReportRequestValidationError) ErrorName() string {
	return "DownloadReportRequestValidationError"
}

// Error satisfies the builtin error interface
func (e DownloadReportRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDownloadReportRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DownloadReportRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DownloadReportRequestValidationError{}

var _DownloadReportRequest_Scope_Pattern = regexp.MustCompile("\\b[A-Z]{3}\\b")

// Validate checks the field values on DownloadReportResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *DownloadReportResponse) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for ReportData

	return nil
}

// DownloadReportResponseValidationError is the validation error returned by
// DownloadReportResponse.Validate if the designated constraints aren't met.
type DownloadReportResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DownloadReportResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DownloadReportResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DownloadReportResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DownloadReportResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DownloadReportResponseValidationError) ErrorName() string {
	return "DownloadReportResponseValidationError"
}

// Error satisfies the builtin error interface
func (e DownloadReportResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDownloadReportResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DownloadReportResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DownloadReportResponseValidationError{}

package v1

import (
	"time"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

type ListNominativeUsersFileUpload struct {
	Id                     int32                    `json:"id,omitempty"`
	Scope                  string                   `json:"scope,omitempty"`
	Swidtag                string                   `json:"swidtag,omitempty"`
	AggregationsId         int32                    `json:"aggregations_id,omitempty"`
	ProductEditor          string                   `json:"product_editor,omitempty"`
	UploadedBy             string                   `json:"uploaded_by,omitempty"`
	NominativeUsersDetails []*NominativeUserDetails `json:"nominative_users_details,omitempty"`
	RecordSucceed          int32                    `json:"record_succeed,omitempty"`
	RecordFailed           int32                    `json:"record_failed,omitempty"`
	FileName               string                   `json:"file_name,omitempty"`
	SheetName              string                   `json:"sheet_name,omitempty"`
	FileStatus             string                   `json:"file_status,omitempty"`
	UploadedAt             time.Time                `json:"uploaded_at,omitempty"`
	UploadId               string                   `json:"upload_id,omitempty"`
	ProductName            string                   `json:"product_name,omitempty"`
	ProductVersion         string                   `json:"product_version,omitempty"`
	AggregationName        string                   `json:"aggregation_name,omitempty"`
	Name                   string                   `json:"name,omitempty"`
	Type                   string                   `json:"type,omitempty"`
}

type NominativeUserDetails struct {
	FirstName      string `json:"first_name,omitempty"`
	UserName       string `json:"user_name,omitempty"`
	Email          string `json:"email,omitempty"`
	Profile        string `json:"profile,omitempty"`
	ActivationDate string `json:"activation_date,omitempty"`
	Comments       string `json:"comments,omitempty"`
}

type ListNominativeUsersFileUploadResponse struct {
	Total       int32                            `json:"total,omitempty"`
	FileDetails []*ListNominativeUsersFileUpload `json:"file_details,omitempty"`
}

type NominativeUser struct {
	Editor               string               `json:"editor,omitempty"`
	ProductName          string               `json:"product_name,omitempty"`
	AggregationName      string               `json:"aggregation_name,omitempty"`
	ProductVersion       string               `json:"product_version,omitempty"`
	UserName             string               `json:"user_name,omitempty"`
	FirstName            string               `json:"first_name,omitempty"`
	UserEmail            string               `json:"user_email,omitempty"`
	Profile              string               `json:"profile,omitempty"`
	ActivationDate       *timestamp.Timestamp `json:"activation_date,omitempty"`
	AggregationId        int32                `json:"aggregation_id,omitempty"`
	Id                   int32                `json:"id,omitempty"`
	Comment              string               `json:"comment,omitempty"`
	ActivationDateString string               `json:"activation_date_string,omitempty"`
	ActivationDateValid  bool                 `json:"activation_date_valid,omitempty"`
}

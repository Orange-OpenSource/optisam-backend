package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/repository/v1/postgres/db"
	v1Svc "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/service/v1"
	v1Notification "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/notification-service/pkg/api/v1"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/product-service/pkg/api/v1"
	v1Product "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/product-service/pkg/api/v1"
)

var (
	NominativeUserRequestCompletedSubject = "OPTISAM: Nominative user upload request completed"
	//NominativeUserRequestCompletedBody    = "Nominative user upload request completed"
	NominativeUserRequestCompletedBody = `<body id="i4r2" style="box-sizing: border-box; margin: 0;">
	  <meta charset="utf-8">
	  <title>Request completion of Nominative user
	  </title>
	<div id="iwli" class="container" style="box-sizing: border-box; max-width: 600px; margin-top: 0px; margin-right: auto; margin-bottom: 0px; margin-left: auto; padding-top: 20px; padding-right: 20px; padding-bottom: 20px; padding-left: 20px; border-top-left-radius: 10px; border-top-right-radius: 10px; border-bottom-right-radius: 10px; border-bottom-left-radius: 10px; background-color: rgb(255, 255, 255); box-shadow: rgba(0, 0, 0, 0.1) 0px 0px 10px;">
	    <p id="iflm" style="box-sizing: border-box; font-size: 16px; color: rgb(0, 0, 0); line-height: 1.5;">Hello,
	    </p>
	    <p id="ie6g" style="box-sizing: border-box; font-size: 16px; color: rgb(0, 0, 0); line-height: 1.5;">Nominative user upload request completed.
	    </p>
	    <p id="ikjyd" style="box-sizing: border-box; font-size: 16px; color: rgb(0, 0, 0); line-height: 1.5;">Thanks </p>

	    <p id="iflm" style="box-sizing: border-box; font-size: 16px; color: rgb(0, 0, 0); line-height: 1.5;">Bonjour,
	    </p>
	    <p id="ie6g" style="box-sizing: border-box; font-size: 16px; color: rgb(0, 0, 0); line-height: 1.5;">Demande de téléchargement d'utilisateur nominatif terminée.
	    </p>
	    <p id="ikjyd" style="box-sizing: border-box; font-size: 16px; color: rgb(0, 0, 0); line-height: 1.5;">Merci </p>
	  </div>
	</body>`
	TopicEmailNotification = "email_notification"
)

func processNominativeUserReq(message kafka.Message, importServer *v1Svc.ImportServiceServer) {
	noOfRetries := getNoOfRetriesFromHeader(message.Headers)
	processnomUsersReqRetry := TopicUpsertNominativeUsersRetry
	if noOfRetries > 20 {
		produceToDLQ(importServer, message, processnomUsersReqRetry)
		return
	}
	if noOfRetries > 0 {
		time.Sleep(time.Second * 30)
	}

	req := v1Product.UpdateNominativeUserRequest{}
	noOfRetries = noOfRetries + 1
	json.Unmarshal(message.Value, &req)
	failedRecords, _ := json.Marshal(req.GetRecordFailed())
	successRecords, _ := json.Marshal(req.GetRecordSucceed())
	err := importServer.ImportRepo.UpdateNominativeUserRequestAnalysisTx(context.Background(), db.UpdateNominativeUserRequestAnalysisParams{
		UploadID:           req.GetUploadId(),
		TotalDgraphBatches: sql.NullInt32{Int32: req.GetTotalDgraphBatches(), Valid: true},
	}, db.UpdateNominativeUserDetailsRequestAnalysisParams{
		RecordFailed:  failedRecords,
		RecordSucceed: successRecords,
	})
	if err != nil {
		logger.Log.Sugar().Errorw("import service - upsert nominativeUsers - processNominativeUserReq - error UpsertNominativeUserRequest - " + err.Error())
		handleError(importServer, &processnomUsersReqRetry, message.Value, noOfRetries, errors.New("import service - upsert nominativeUsers - processNominativeUserReq - UpsertNominativeUserRequest"+err.Error()))
	} else {
		logger.Log.Sugar().Info("import service - upsert nominativeUsers - processNominativeUserReq - successfully inserted data to UpsertNominativeUserRequest")
	}
}

func processNomPostgresSuccess(message kafka.Message, importServer *v1Svc.ImportServiceServer) {
	noOfRetries := getNoOfRetriesFromHeader(message.Headers)
	processNomPostgresSuccessRetry := TopicProcessNomPostgresSuccessRetry
	if noOfRetries > 20 {
		produceToDLQ(importServer, message, processNomPostgresSuccessRetry)
		return
	}
	if noOfRetries > 0 {
		time.Sleep(time.Second * 3)
	}

	req := v1.UpdatePostgresBatchSucessNomUSers{}
	noOfRetries = noOfRetries + 1
	json.Unmarshal(message.Value, &req)
	r, err := importServer.ImportRepo.UpdateNominativeUserRequestPostgresSuccess(context.Background(), db.UpdateNominativeUserRequestPostgresSuccessParams{
		UploadID:        req.GetUploadId(),
		PostgresSuccess: sql.NullBool{Bool: req.GetSuccess(), Valid: true},
	})
	if err != nil {
		logger.Log.Sugar().Errorw("import service - upsert nominativeUsers - processNomPostgresSuccess - error UpdateNominativeUserRequestPostgresSuccess - " + err.Error())
		handleError(importServer, &processNomPostgresSuccessRetry, message.Value, noOfRetries, errors.New("import service - upsert nominativeUsers - processNomPostgresSuccess - UpdateNominativeUserRequestPostgresSuccess"+err.Error()))
	}
	if r.DgraphCompletedBatches.Int32 == r.TotalDgraphBatches.Int32 {
		err := handleFileUploadSuccess(importServer, r.RequestID, req.Scope)
		if err != nil {
			logger.Log.Sugar().Errorw("import service - upsert nominativeUsers - processNomPostgresSuccess - error handleFileUploadSuccess - " + err.Error())
			handleError(importServer, &processNomPostgresSuccessRetry, message.Value, noOfRetries, errors.New("import service - upsert nominativeUsers - processNomPostgresSuccess - UpdateNominativeUserRequestPostgresSuccess"+err.Error()))
			return
		}
	}
	logger.Log.Sugar().Info("import service - upsert nominativeUsers - processNomPostgresSuccess - successfully inserted data to UpdateNominativeUserRequestPostgresSuccess")
}

func processNomDgraphSuccess(message kafka.Message, importServer *v1Svc.ImportServiceServer) {
	noOfRetries := getNoOfRetriesFromHeader(message.Headers)
	processNomDgraphSuccessRetry := TopicProcessNomDgraphSuccessRetry
	if noOfRetries > 20 {
		produceToDLQ(importServer, message, processNomDgraphSuccessRetry)
		return
	}
	if noOfRetries > 0 {
		time.Sleep(time.Second * 30)
	}

	req := v1.UpdateDgraphBatchSuccessCount{}
	noOfRetries = noOfRetries + 1
	json.Unmarshal(message.Value, &req)
	r, err := importServer.ImportRepo.UpdateNominativeUserRequestDgraphBatchSuccess(context.Background(), req.GetUploadId())
	if err != nil {
		logger.Log.Sugar().Errorw("import service - upsert nominativeUsers - processNomPostgresSuccess - error UpdateNominativeUserRequestDgraphBatchSuccess - " + err.Error())
		handleError(importServer, &processNomDgraphSuccessRetry, message.Value, noOfRetries, errors.New("import service - upsert nominativeUsers - processNomPostgresSuccess - UpdateNominativeUserRequestDgraphBatchSuccess"+err.Error()))
		return
	}
	if (r.DgraphCompletedBatches.Int32 == r.TotalDgraphBatches.Int32) && (r.PostgresSuccess.Bool == true) {
		err := handleFileUploadSuccess(importServer, r.RequestID, req.Scope)
		if err != nil {
			logger.Log.Sugar().Errorw("import service - upsert nominativeUsers - processNomPostgresSuccess - error handleFileUploadSuccess - " + err.Error())
			handleError(importServer, &processNomDgraphSuccessRetry, message.Value, noOfRetries, errors.New("import service - upsert nominativeUsers - processNomPostgresSuccess - UpdateNominativeUserRequestPostgresSuccess"+err.Error()))
			return
		}
	}
	logger.Log.Sugar().Info("import service - upsert nominativeUsers - processNomPostgresSuccess - successfully inserted data to UpdateNominativeUserRequestDgraphBatchSuccess")

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

func produceToDLQ(importServer *v1Svc.ImportServiceServer, msg kafka.Message, topic string) {
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
	importServer.KafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topicDQL,
			//Partition: rand.Int31n(importServer.Config.NoOfPartitions)},
			Partition: kafka.PartitionAny},
		Value: d,
	}, nil)
}

func handleError(importServer *v1Svc.ImportServiceServer, topic *string, value []byte, noOfRetries int, err error) {
	importServer.KafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: topic, Partition: rand.Int31n(importServer.Config.NoOfPartitions)},
		Value:          value,
		Headers:        []kafka.Header{{Key: NoOfRetries, Value: []byte(strconv.Itoa(noOfRetries))}, {Key: "error", Value: []byte(err.Error())}},
	}, nil)
}

// func processCDCNominativeUserRequests(message kafka.Message, importServer *v1Svc.ImportServiceServer) {
// 	noOfRetries := getNoOfRetriesFromHeader(message.Headers)
// 	if noOfRetries > 20 {
// 		produceToDLQ(importServer, message, topicCDCNominativeUserRequestsSuccessRetry)
// 		return
// 	}
// 	if noOfRetries > 0 {
// 		time.Sleep(time.Second * 3)
// 	}

// 	req := TopicCDCNominativeUserRequests{}
// 	noOfRetries = noOfRetries + 1
// 	err := json.Unmarshal(message.Value, &req)
// 	if err != nil {
// 		logger.Log.Sugar().Errorw("import service - upsert nominativeUsers - processCDCNominativeUserRequests - error Unmarshal - " + err.Error())
// 		handleError(importServer, &topicCDCNominativeUserRequestsSuccessRetry, message.Value, noOfRetries, errors.New("import service - upsert nominativeUsers - processCDCNominativeUserRequests - error Unmarshal - "+err.Error()))
// 	}
// 	r := req.Payload.After
// 	switch {
// 	case r.PostgresSuccess && (r.DgraphCompletedBatches == r.TotalDgraphBatches) && r.Status == "PENDING":
// 		var status string
// 		resp, err := importServer.ImportRepo.ListNominativeUsersUploadedFiles(context.Background(), db.ListNominativeUsersUploadedFilesParams{
// 			Scope:        []string{r.Scope},
// 			ID:           int32(r.RequestID),
// 			FileUploadID: true,
// 		})
// 		if err != nil {
// 			logger.Log.Sugar().Errorw("import service - upsert nominativeUsers - processCDCNominativeUserRequests - ListNominativeUsersUploadedFiles - error fetching file details - " + err.Error())
// 			handleError(importServer, &topicCDCNominativeUserRequestsSuccessRetry, message.Value, noOfRetries, errors.New("import service - upsert nominativeUsers - processCDCNominativeUserRequests ListNominativeUsersUploadedFiles - error fetching file details -  "+err.Error()))
// 		}
// 		if len(resp) > 0 {
// 			if len(resp[0].RecordFailed) > 0 {
// 				status = "PARTIAL"
// 			} else {
// 				status = "SUCCESS"
// 			}
// 		}

// 		err = importServer.ImportRepo.UpdateNominativeUserRequestSuccess(context.Background(), db.UpdateNominativeUserRequestSuccessParams{
// 			UploadID:      r.UploadID,
// 			DgraphSuccess: sql.NullBool{Bool: true, Valid: true},
// 			Status:        status,
// 		})
// 		if err != nil {
// 			logger.Log.Sugar().Errorw("import service - upsert nominativeUsers - processCDCNominativeUserRequests - error UpdateNominativeUserRequestSuccess - " + err.Error())
// 			handleError(importServer, &topicCDCNominativeUserRequestsSuccessRetry, message.Value, noOfRetries, errors.New("import service - upsert nominativeUsers - processCDCNominativeUserRequests - UpdateNominativeUserRequestSuccess"+err.Error()))
// 		} else {
// 			logger.Log.Sugar().Info("import service - upsert nominativeUsers - processCDCNominativeUserRequests - successfully inserted data to UpdateNominativeUserRequestSuccess,request completed")
// 			in := v1Notification.SendMailRequest{
// 				MailSubject: NominativeUserRequestCompletedSubject,
// 				MailMessage: NominativeUserRequestCompletedBody,
// 				MailTo:      []string{r.CreatedBy},
// 			}
// 			notificationReq, _ := json.Marshal(in)
// 			t := TopicEmailNotification
// 			importServer.KafkaProducer.Produce(&kafka.Message{
// 				TopicPartition: kafka.TopicPartition{Topic: &t, Partition: rand.Int31n(importServer.Config.NoOfPartitions)},
// 				Value:          []byte(notificationReq),
// 			}, nil)
// 		}
// 		// importServer.KafkaProducer.Produce(&kafka.Message{
// 		// 	TopicPartition: kafka.TopicPartition{Topic: topic, Partition: kafka.PartitionAny},
// 		// 	Value:          value,
// 		// 	Headers:        []kafka.Header{{Key: NoOfRetries, Value: []byte(strconv.Itoa(noOfRetries))}, {Key: "error", Value: []byte(err.Error())}},
// 		// }, nil)
// 		// case r.DgraphCompletedBatches == r.TotalDgraphBatches && r.Status == "PENDING":
// 		// 	err := importServer.ImportRepo.UpdateNominativeUserRequestDgraphSuccess(context.Background(), db.UpdateNominativeUserRequestDgraphSuccessParams{
// 		// 		UploadID:      r.UploadID,
// 		// 		DgraphSuccess: sql.NullBool{Bool: true, Valid: true},
// 		// 	})
// 		// 	if err != nil {
// 		// 		logger.Log.Sugar().Errorw("import service - upsert nominativeUsers - processCDCNominativeUserRequests - error UpdateNominativeUserRequestDgraphSuccess - " + err.Error())
// 		// 		handleError(importServer, &topicCDCNominativeUserRequestsSuccessRetry, message.Value, noOfRetries, errors.New("import service - upsert nominativeUsers - processCDCNominativeUserRequests - UpdateNominativeUserRequestDgraphSuccess"+err.Error()))
// 		// 	} else {
// 		// 		logger.Log.Sugar().Info("import service - upsert nominativeUsers - processCDCNominativeUserRequests - successfully inserted data to UpdateNominativeUserRequestDgraphSuccess")
// 		// 	}
// 	}
// }

func handleFileUploadSuccess(importServer *v1Svc.ImportServiceServer, reqId int32, scope string) (err error) {
	var status, createdBy string
	resp, err := importServer.ImportRepo.ListNominativeUsersUploadedFiles(context.Background(), db.ListNominativeUsersUploadedFilesParams{
		Scope:        []string{scope},
		ID:           reqId,
		FileUploadID: true,
		PageNum:      0,
		PageSize:     20,
	})
	if err != nil {
		logger.Log.Sugar().Errorw("import service - upsert nominativeUsers - processCDCNominativeUserRequests - ListNominativeUsersUploadedFiles - error fetching file details - " + err.Error())
		return
		//handleError(importServer, &topicCDCNominativeUserRequestsSuccessRetry, message.Value, noOfRetries, errors.New("import service - upsert nominativeUsers - processCDCNominativeUserRequests ListNominativeUsersUploadedFiles - error fetching file details -  "+err.Error()))
	}
	if len(resp) > 0 {
		if resp[0].RecordFailed_2.(int64) > 0 {
			status = "PARTIAL"
		} else {
			status = "SUCCESS"
		}
		createdBy = resp[0].CreatedBy.String
		//scope = resp[0].Scope
	}

	err = importServer.ImportRepo.UpdateNominativeUserRequestSuccess(context.Background(), db.UpdateNominativeUserRequestSuccessParams{
		RequestID:     reqId,
		DgraphSuccess: sql.NullBool{Bool: true, Valid: true},
		Status:        status,
	})
	if err != nil {
		logger.Log.Sugar().Errorw("import service - upsert nominativeUsers - processCDCNominativeUserRequests - error UpdateNominativeUserRequestSuccess - " + err.Error())
		return
		//handleError(importServer, &topicCDCNominativeUserRequestsSuccessRetry, message.Value, noOfRetries, errors.New("import service - upsert nominativeUsers - processCDCNominativeUserRequests - UpdateNominativeUserRequestSuccess"+err.Error()))
	} else {
		logger.Log.Sugar().Info("import service - upsert nominativeUsers - processCDCNominativeUserRequests - successfully inserted data to UpdateNominativeUserRequestSuccess,request completed")
		in := v1Notification.SendMailRequest{
			MailSubject: NominativeUserRequestCompletedSubject,
			MailMessage: NominativeUserRequestCompletedBody,
			MailTo:      []string{createdBy},
		}
		notificationReq, _ := json.Marshal(in)
		t := TopicEmailNotification
		importServer.KafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &t, Partition: rand.Int31n(importServer.Config.NoOfPartitions)},
			Value:          []byte(notificationReq),
		}, nil)
	}
	return
}

// type TopicCDCNominativeUserRequests struct {
// 	Payload struct {
// 		Before any `json:"before"`
// 		After  struct {
// 			RequestID              int       `json:"request_id"`
// 			UploadID               string    `json:"upload_id"`
// 			Scope                  string    `json:"scope"`
// 			Swidtag                string    `json:"swidtag"`
// 			Status                 string    `json:"status"`
// 			ProductName            string    `json:"product_name"`
// 			ProductVersion         string    `json:"product_version"`
// 			AggregationID          string    `json:"aggregation_id"`
// 			Editor                 string    `json:"editor"`
// 			FileName               string    `json:"file_name"`
// 			FileLocation           string    `json:"file_location"`
// 			SheetName              string    `json:"sheet_name"`
// 			PostgresSuccess        bool      `json:"postgres_success"`
// 			DgraphSuccess          bool      `json:"dgraph_success"`
// 			TotalDgraphBatches     int       `json:"total_dgraph_batches"`
// 			DgraphCompletedBatches int       `json:"dgraph_completed_batches"`
// 			CreatedAt              time.Time `json:"created_at"`
// 			CreatedBy              string    `json:"created_by"`
// 		} `json:"after"`
// 		Source struct {
// 			Version   string `json:"version"`
// 			Connector string `json:"connector"`
// 			Name      string `json:"name"`
// 			TsMs      int64  `json:"ts_ms"`
// 			Snapshot  string `json:"snapshot"`
// 			Db        string `json:"db"`
// 			Sequence  string `json:"sequence"`
// 			Schema    string `json:"schema"`
// 			Table     string `json:"table"`
// 			TxID      int    `json:"txId"`
// 			Lsn       int64  `json:"lsn"`
// 			Xmin      any    `json:"xmin"`
// 		} `json:"source"`
// 		Op          string `json:"op"`
// 		TsMs        int64  `json:"ts_ms"`
// 		Transaction any    `json:"transaction"`
// 	} `json:"payload"`
// }

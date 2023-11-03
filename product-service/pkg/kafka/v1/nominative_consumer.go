package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/mail"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	dgworker "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/worker/dgraph"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/xuri/excelize/v2"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	v1Svc "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/service/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	//"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	YYYYMMDD string = "2006-01-02"
	DDMMYYYY string = "02-01-2006"
)

var (
	NoOfRetries = "no_of_retries"
)
var dateFormats = []string{YYYYMMDD, DDMMYYYY}

func filterNominativeUsers(req *v1.UpserNominativeUserRequest) (nomUsersValid, nomUsersInValid []*v1.NominativeUser, err error) {
	users := make(map[string]bool)
	//fromFile := req.FileName != ""
	for _, v := range req.UserDetails {
		var startTime time.Time
		var err error
		nomUser := v1.NominativeUser{}
		if v.ActivationDate != "" {
			ts, err := strconv.Atoi(v.ActivationDate)
			if err == nil {
				startTime, _ = excelize.ExcelDateToTime(float64(ts), false)
			}
			if startTime.IsZero() {
				if len(v.ActivationDate) <= 10 {
					if strings.Contains(v.ActivationDate, "/") {
						v.ActivationDate = strings.ReplaceAll(v.ActivationDate, "/", "-")
					}
					for _, format := range dateFormats {
						startTime, err = time.Parse(format, v.ActivationDate)
						if err == nil {
							break
						}
					}
					if startTime.IsZero() {
						logger.Log.Sugar().Errorw("error parsing time")
					}
				} else if len(v.ActivationDate) > 10 && len(v.ActivationDate) <= 24 {
					if strings.Contains(v.ActivationDate, "/") && len(v.ActivationDate) <= 8 {
						startTime, err = time.Parse("06/2/1T15:04:05.000Z", v.ActivationDate)
					} else if strings.Contains(v.ActivationDate, "/") {
						startTime, err = time.Parse("2006/01/02T15:04:05.000Z", v.ActivationDate)
					} else {
						startTime, err = time.Parse("2006-01-02T15:04:05.000Z", v.ActivationDate)
					}
				}
			}
			nomUser.ActivationDateString = v.ActivationDate
			if err == nil {
				nomUser.ActivationDate = timestamppb.New(startTime)
				nomUser.ActivationDateValid = true
			}
			err = nil
		}
		_, err = mail.ParseAddress(v.Email)
		if err != nil {
			nomUser.Comment = "Invalid email format"
		}
		if _, ok := users[v.Email+v.Profile]; ok {
			nomUser.Comment = "duplicate entry"
			err = errors.New("duplicate entry")
		} else {
			users[v.Email+v.Profile] = true
		}
		if err != nil {
			nomUser.ActivationDate = timestamppb.New(startTime)
			nomUser.UserEmail = v.GetEmail()
			nomUser.FirstName = v.GetFirstName()
			nomUser.Profile = v.GetProfile()
			nomUser.UserName = v.GetUserName()
			nomUsersInValid = append(nomUsersInValid, &nomUser)
			// if !fromFile {
			// 	return nomUsersValid, nomUsersInValid, err
			// }
			continue
		} else {
			nomUsersValid = append(nomUsersValid, &v1.NominativeUser{
				UserName:       v.GetUserName(),
				UserEmail:      v.GetEmail(),
				FirstName:      v.GetFirstName(),
				Profile:        v.GetProfile(),
				ActivationDate: timestamppb.New(startTime),
			})
		}
	}
	return nomUsersValid, nomUsersInValid, err
}

func handlePostgresNominativeUserRequest(message kafka.Message, productServer *v1Svc.ProductServiceServer, wg *sync.WaitGroup) {
	defer wg.Done()
	noOfRetries := getNoOfRetriesFromHeader(message.Headers)
	upsertNominativeUsersPostgres := TopicUpsertNominativeUsersPostgres
	upsertNominativeUsersPostgresRetry := TopicUpsertNominativeUsersPostgresRetry
	if noOfRetries > 20 {
		produceToDLQ(productServer, message, upsertNominativeUsersPostgresRetry)
		return
	}
	if noOfRetries > 0 {
		time.Sleep(time.Second * 30)
	}
	noOfRetries = noOfRetries + 1
	err := productServer.KafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &upsertNominativeUsersPostgres,
			Partition: rand.Int31n(productServer.Cfg.NoOfPartitions)},
		Value: message.Value,
	}, nil)
	if err != nil {
		logger.Log.Sugar().Errorw("product service - upsert nominativeUsers - handlePostgresNominativeUserRequest - error producing upsert_nominative_users_postgres event" + err.Error())
		handleError(productServer, &upsertNominativeUsersPostgresRetry, message.Value, noOfRetries, errors.New("product service - upsert nominativeUsers - handlePostgresNominativeUserRequest - error producing upsert_nominative_users_postgres event"+err.Error()))
	} else {
		logger.Log.Sugar().Debug("product service - upsert nominativeUsers - handlePostgresNominativeUserRequest - successfully produced event to upsert_nominative_users_postgres")
	}
}

func handleUpsertNominativeUsersRequest(message kafka.Message, productServer *v1Svc.ProductServiceServer) {
	logger.Log.Sugar().Debug("Called handleUpsertNominativeUsersRequest")
	noOfRetries := getNoOfRetriesFromHeader(message.Headers)
	upserNomUsersRetry := TopicUpsertNominativeUsersRetry
	if noOfRetries > 20 {
		produceToDLQ(productServer, message, upserNomUsersRetry)
		return
	}
	if noOfRetries > 0 {
		time.Sleep(time.Second * RetryDelay)
	}
	req := v1.UpserNominativeUserRequest{}
	noOfRetries = noOfRetries + 1
	err := json.Unmarshal(message.Value, &req)
	if err != nil {
		logger.Log.Sugar().Error("TopicUpsertNominativeUsers : error unmarshaling UpserNominativeUserRequest %v\n", err)
		handleError(productServer, &upserNomUsersRetry, message.Value, noOfRetries, errors.New("TopicUpsertNominativeUsers : error unmarshaling UpserNominativeUserRequest :"+err.Error()))
	}
	validNomUsers, inValidNomUsers, err := filterNominativeUsers(&req)
	var wg sync.WaitGroup
	if err != nil {
		logger.Log.Sugar().Errorw("product service - upsert nominativeUsers - handleUpsertNominativeUsersRequest - error producing upsert_nominative_users_retry event" + err.Error())
		handleError(productServer, &upserNomUsersRetry, message.Value, noOfRetries, err)
	}
	noOfDgraphBatches := make(chan int32)
	req.ValidNominativeUsers = validNomUsers
	r, _ := json.Marshal(&req)
	message.Value = r
	wg.Add(3)
	go handlePostgresNominativeUserRequest(kafka.Message{Value: r}, productServer, &wg)
	go handleDgraphNominativeUserRequest(kafka.Message{Value: r}, productServer, &wg, noOfDgraphBatches)
	go handleUpdateNominativeUserRequest(kafka.Message{Value: r}, noOfDgraphBatches, validNomUsers, inValidNomUsers, &wg, productServer, req.GetUploadId())
	wg.Wait()
	close(noOfDgraphBatches)

}

func handleUpdateNominativeUserRequest(message kafka.Message, noOfDgraphBatches chan int32, validNomUsers, inValidNomUsers []*v1.NominativeUser, wg *sync.WaitGroup, productServer *v1Svc.ProductServiceServer, uploadId string) {
	defer wg.Done()
	updateNomUserReqRetry := TopicUpdateNominativeUserRequestRetry

	noOfRetries := 1
	noBatches := <-noOfDgraphBatches
	updateNomUserReq := TopicUpdateNominativeUserRequest
	updateRequest := v1.UpdateNominativeUserRequest{
		UploadId:           uploadId,
		RecordSucceed:      validNomUsers,
		RecordFailed:       inValidNomUsers,
		TotalDgraphBatches: noBatches,
	}
	updateNominativeUserReq, _ := json.Marshal(&updateRequest)
	err := productServer.KafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &updateNomUserReq,
			Partition: rand.Int31n(productServer.Cfg.NoOfPartitions)},
		Value: updateNominativeUserReq,
	}, nil)
	if err != nil {
		logger.Log.Sugar().Errorw("product service - upsert nominativeUsers - handleUpdateNominativeUserRequest - error producing update_nominative_user_request event" + err.Error())
		handleError(productServer, &updateNomUserReqRetry, updateNominativeUserReq, noOfRetries, errors.New("product service - upsert nominativeUsers - handleUpdateNominativeUserRequest - error producing update_nominative_user_request event"+err.Error()))
	} else {
		logger.Log.Sugar().Debug("product service - upsert nominativeUsers - handleUpdateNominativeUserRequest - successfully produced event to update_nominative_user_request")
	}
}
func handleUpdateNominativeUserRequestRetry(message kafka.Message, productServer *v1Svc.ProductServiceServer) {
	noOfRetries := getNoOfRetriesFromHeader(message.Headers)
	updateNomUserReqRetry := TopicUpdateNominativeUserRequestRetry
	if noOfRetries > 20 {
		produceToDLQ(productServer, message, updateNomUserReqRetry)
		return
	}
	if noOfRetries > 0 {
		time.Sleep(time.Second * RetryDelay)
	}
	noOfRetries = noOfRetries + 1
	updateNomUserReq := TopicUpdateNominativeUserRequest

	err := productServer.KafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &updateNomUserReq,
			Partition: rand.Int31n(productServer.Cfg.NoOfPartitions)},
		Value: message.Value,
	}, nil)
	if err != nil {
		logger.Log.Sugar().Errorw("product service - upsert nominativeUsers - handleUpdateNominativeUserRequest - error producing update_nominative_user_request event" + err.Error())
		handleError(productServer, &updateNomUserReqRetry, message.Value, noOfRetries, errors.New("product service - upsert nominativeUsers - handleUpdateNominativeUserRequest - error producing update_nominative_user_request event"+err.Error()))
	} else {
		logger.Log.Sugar().Debug("product service - upsert nominativeUsers - handleUpdateNominativeUserRequest - successfully produced event to update_nominative_user_request")
	}
}
func handleDgraphNominativeUserRequest(message kafka.Message, productServer *v1Svc.ProductServiceServer, wg *sync.WaitGroup, noOfDgraphBatches chan int32) {
	defer wg.Done()
	noOfRetries := getNoOfRetriesFromHeader(message.Headers)
	upsertNominativeUsersDgraphRetry := TopicUpsertNominativeUsersDgraphRetry
	if noOfRetries > 20 {
		produceToDLQ(productServer, message, upsertNominativeUsersDgraphRetry)
		return
	}
	if noOfRetries > 0 {
		time.Sleep(time.Second * RetryDelay)
	}
	req := v1.UpserNominativeUserRequest{}
	json.Unmarshal(message.Value, &req)
	noOfRetries = noOfRetries + 1
	upsertNominativeReqDgraph := prepairUpsertNominativeUserDgraphRequest(&req, productServer.Cfg.DgraphBatchSize)
	upsertNominativeUsersDgraph := TopicUpsertNominativeUsersDgraph

	noOfDgraphBatches <- int32(len(upsertNominativeReqDgraph))
	for _, v := range upsertNominativeReqDgraph {
		dgraphNomUsersBatch, _ := json.Marshal(v)
		err := productServer.KafkaProducer.Produce(&kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &upsertNominativeUsersDgraph,
			Partition: rand.Int31n(productServer.Cfg.NoOfPartitions)},
			Value: dgraphNomUsersBatch,
		}, nil)
		if err != nil {
			logger.Log.Sugar().Errorw("product service - upsert nominativeUsers - handleDgraphNominativeUserRequest - error producing upsert_nominative_users_dgraph event" + err.Error())
			handleError(productServer, &upsertNominativeUsersDgraphRetry, dgraphNomUsersBatch, noOfRetries, errors.New("product service - upsert nominativeUsers - handleDgraphNominativeUserRequest - error producing upsert_nominative_users_dgraph event"+err.Error()))
		} else {
			logger.Log.Sugar().Debug("product service - upsert handleDgraphNominativeUserRequest - successfully produced event to upsert_nominative_users_dgraph")
		}
	}
}

func handleError(productServer *v1Svc.ProductServiceServer, topic *string, value []byte, noOfRetries int, err error) {
	productServer.KafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: topic, Partition: rand.Int31n(productServer.Cfg.NoOfPartitions)},
		Value:          value,
		Headers:        []kafka.Header{kafka.Header{Key: NoOfRetries, Value: []byte(strconv.Itoa(noOfRetries))}, kafka.Header{Key: "error", Value: []byte(err.Error())}},
	}, nil)
}

func produceToDLQ(productServer *v1Svc.ProductServiceServer, msg kafka.Message, topic string) {
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
	productServer.KafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topicDQL,
			Partition: rand.Int31n(productServer.Cfg.NoOfPartitions)},
		Value: d,
	}, nil)
}

func handleDgraphNominativeUserRequestRetry(message kafka.Message, productServer *v1Svc.ProductServiceServer) {
	noOfRetries := getNoOfRetriesFromHeader(message.Headers)
	upsertNominativeUsersDgraphRetry := TopicUpsertNominativeUsersDgraphRetry
	if noOfRetries > 20 {
		produceToDLQ(productServer, message, upsertNominativeUsersDgraphRetry)
		return
	}
	if noOfRetries > 0 {
		time.Sleep(time.Second * RetryDelay)
	}
	noOfRetries = noOfRetries + 1
	upsertNominativeUsersDgraph := TopicUpsertNominativeUsersDgraph
	err := productServer.KafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &upsertNominativeUsersDgraph,
			Partition: rand.Int31n(productServer.Cfg.NoOfPartitions)},
		Value: message.Value,
	}, nil)
	if err != nil {
		logger.Log.Sugar().Errorw("product service - upsert nominativeUsers - handleDgraphNominativeUserRequest - error producing upsert_nominative_users_dgraph event" + err.Error())
		handleError(productServer, &upsertNominativeUsersDgraphRetry, message.Value, noOfRetries, errors.New("product service - upsert nominativeUsers - handleDgraphNominativeUserRequest - error producing upsert_nominative_users_dgraph event"+err.Error()))
	} else {
		logger.Log.Sugar().Debug("product service - upsert handleDgraphNominativeUserRequest - successfully produced event to upsert_nominative_users_dgraph")
	}

}

func processNominativeDgraphBatch(msg kafka.Message, productServer *v1Svc.ProductServiceServer) {
	noOfRetries := getNoOfRetriesFromHeader(msg.Headers)
	processNomUserRecordsDgraphRetry := TopicProcessNominativeUserRecordsDgraphRetry
	if noOfRetries > 20 {
		produceToDLQ(productServer, msg, processNomUserRecordsDgraphRetry)
		return
	}
	if noOfRetries > 0 {
		time.Sleep(time.Second * 2)
	}
	noOfRetries = noOfRetries + 1
	var unur dgworker.UpserNominativeUserRequest
	_ = json.Unmarshal(msg.Value, &unur)
	var mutations []*api.Mutation
	var queries []string
	queries = append(queries, "query", "{")
	for i, v := range unur.UserDetails {
		query := `var(func: eq(nominative.user.email,"` + v.Email + `")) @filter(eq(type_name,"nominative_user") 
			AND eq(scopes,"` + unur.Scope + `") AND eq(nominative.user.profile,"` + v.Profile + `")
			AND eq(nominative.user.swidtag,"` + unur.SwidTag + `") AND eq(nominative.user.aggregation.id,"` + strconv.Itoa(int(unur.AggregationId)) + `")){
				user_` + strconv.Itoa(i) + ` as uid
			}`
		queries = append(queries, query)
		mutations = append(mutations, &api.Mutation{
			Cond: `@if(eq(len(user_` + strconv.Itoa(i) + `),0))`,
			SetNquads: []byte(`
				uid(user_` + strconv.Itoa(i) + `) <type_name> "nominative_user" .
				uid(user_` + strconv.Itoa(i) + `) <dgraph.type> "NominativeUser" .
				uid(user_` + strconv.Itoa(i) + `) <scopes> "` + unur.Scope + `" .
				uid(user_` + strconv.Itoa(i) + `) <created> "` + unur.CreatedBy + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.swidtag> "` + unur.SwidTag + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.editor> "` + unur.Editor + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.aggregation.id> "` + strconv.Itoa(int(unur.AggregationId)) + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.email> "` + v.Email + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.first_name> "` + v.FirstName + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.name> "` + v.UserName + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.profile> "` + v.Profile + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.activation.date> "` + v.ActivationDate.String() + `" .
				`),
			CommitNow: true,
		})
		mutations = append(mutations, &api.Mutation{
			Cond: `@if(NOT eq(len(user_` + strconv.Itoa(i) + `),0))`,
			SetNquads: []byte(`
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.first_name> "` + v.FirstName + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.name> "` + v.UserName + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.profile> "` + v.Profile + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.activation.date> "` + v.ActivationDate.String() + `" .
				`),
			CommitNow: true,
		})
		if unur.AggregationId > 0 {
			query = `var(func: eq(aggregation.id,"` + strconv.Itoa(int(unur.AggregationId)) + `")) @filter(eq(type_name,"aggregation") AND eq(scopes,"` + unur.Scope + `")){
					aggregation_` + strconv.Itoa(i) + ` as uid
				}`
			queries = append(queries, query)
			mutations = append(mutations, &api.Mutation{
				Cond: `@if(eq(len(user_` + strconv.Itoa(i) + `),0))`,
				SetNquads: []byte(`
					uid(aggregation_` + strconv.Itoa(i) + `)  <aggregation.nominative.users> uid(user_` + strconv.Itoa(i) + `) .
				`),
				CommitNow: true,
			})
		} else if unur.SwidTag != "" {
			query = `var(func: eq(product.swidtag,"` + unur.SwidTag + `")) @filter(eq(type_name,"product") AND eq(scopes,"` + unur.Scope + `")){
					product_` + strconv.Itoa(i) + ` as uid
				}`
			queries = append(queries, query)
			mutations = append(mutations, &api.Mutation{
				Cond: `@if(eq(len(user_` + strconv.Itoa(i) + `),0))`,
				SetNquads: []byte(`
					uid(product_` + strconv.Itoa(i) + `)  <product.nominative.users> uid(user_` + strconv.Itoa(i) + `) .
				`),
				CommitNow: true,
			})
		}

	}

	queries = append(queries, "}")
	q := strings.Join(queries, "\n")
	req := &api.Request{
		Query:     q,
		Mutations: mutations,
		CommitNow: true,
	}
	if _, err := productServer.Dg.NewTxn().Do(context.Background(), req); err != nil {
		logger.Log.Sugar().Errorw("product service - upsert nominativeUsers - processNominativeDgraphBatch - error in upsert nominativeUsers - " + err.Error())
		handleError(productServer, &processNomUserRecordsDgraphRetry, msg.Value, noOfRetries, errors.New("product service - upsert nominativeUsers - processNominativeDgraphBatch - error in upsert nominativeUsers - "+err.Error()))
	} else {
		logger.Log.Sugar().Debug("product service - upsert nominativeUsers - processNominativeDgraphBatch - successfully processed nominative user batch to Dgraph")
		updateDgraphSuccessRequest := v1.UpdateDgraphBatchSuccessCount{
			Success:  true,
			UploadId: unur.UploadId,
			Scope:    unur.Scope,
		}
		r, _ := json.Marshal(&updateDgraphSuccessRequest)
		err := productServer.KafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topicProcessNomDgraphSuccess,
				Partition: rand.Int31n(productServer.Cfg.NoOfPartitions)},
			Value: r,
		}, nil)
		if err != nil {
			logger.Log.Sugar().Errorw("product service - upsert nominativeUsers - processNominativeDgraphBatch - error producing process_nom_postgres_success event" + err.Error())
			handleError(productServer, &topicProcessNomDgraphSuccessRetry, msg.Value, noOfRetries, errors.New("product service - upsert nominativeUsers - processNominativeDgraphBatch - error producing process_nom_postgres_success event"+err.Error()))
		} else {
			logger.Log.Sugar().Debug("product service - upsert processNominativeDgraphBatch - successfully produced event to process_nom_postgres_success")
		}

	}
}

func processPostgresNominativeUserUpsert(message kafka.Message, productServer *v1Svc.ProductServiceServer) {
	noOfRetries := getNoOfRetriesFromHeader(message.Headers)
	processNomUserRecordsPostgresRetry := TopicProcessNominativeUserRecordsPostgresRetry
	if noOfRetries > 20 {
		produceToDLQ(productServer, message, processNomUserRecordsPostgresRetry)
		return
	}
	if noOfRetries > 0 {
		time.Sleep(time.Second * 30)
	}

	req := v1.UpserNominativeUserRequest{}
	noOfRetries = noOfRetries + 1
	json.Unmarshal(message.Value, &req)
	err := productServer.ProductRepo.UpsertNominativeUsersTx(context.Background(), &req)
	if err != nil {
		logger.Log.Sugar().Errorw("product service - upsert nominativeUsers - processPostgresNominativeUserUpsert - error UpsertNominativeUsersTx - " + err.Error())
		handleError(productServer, &processNomUserRecordsPostgresRetry, message.Value, noOfRetries, errors.New("product service - upsert nominativeUsers - processPostgresNominativeUserUpsert - UpsertNominativeUsersTx"+err.Error()))
	} else {
		logger.Log.Sugar().Debug("product service - upsert nominativeUsers - processPostgresNominativeUserUpsert - successfully inserted data to UpsertNominativeUsersTx")
		updatepostSuccessRequest := v1.UpdatePostgresBatchSucessNomUSers{
			Success:  true,
			UploadId: req.GetUploadId(),
			Scope:    req.Scope,
		}
		r, _ := json.Marshal(&updatepostSuccessRequest)
		err := productServer.KafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topicProcessNomPostgresSuccess,
				Partition: rand.Int31n(productServer.Cfg.NoOfPartitions)},
			Value: r,
		}, nil)
		if err != nil {
			logger.Log.Sugar().Errorw("product service - upsert nominativeUsers - processPostgresNominativeUserUpsert - error producing process_nom_dgraph_success event" + err.Error())
			handleError(productServer, &topicProcessNomPostgresSuccessRetry, message.Value, noOfRetries, errors.New("product service - upsert nominativeUsers - processPostgresNominativeUserUpsert - error producing process_nom_dgraph_success event"+err.Error()))
		} else {
			logger.Log.Sugar().Debug("product service - upsert processPostgresNominativeUserUpsert - successfully produced event to process_nom_dgraph_success")
		}
	}
}

func prepairUpsertNominativeUserDgraphRequest(req *v1.UpserNominativeUserRequest, batch int) (response []dgworker.UpserNominativeUserRequest) {
	var resp dgworker.UpserNominativeUserRequest
	resp.AggregationId = req.GetAggregationId()
	resp.Editor = req.GetEditor()
	resp.ProductName = req.GetProductVersion()
	resp.ProductVersion = req.GetProductVersion()
	resp.Scope = req.GetScope()
	resp.SwidTag = req.GetSwidTag()
	resp.CreatedBy = req.GetCreatedBy()
	resp.UploadId = req.UploadId
	usersBatch := createBatchNominativeUsers(req.GetValidNominativeUsers(), batch)
	for _, b := range usersBatch {
		var usrs []*dgworker.NominativeUserDetails
		for _, v := range b {
			var userDetails dgworker.NominativeUserDetails
			userDetails.Email = v.GetUserEmail()
			userDetails.FirstName = v.GetFirstName()
			userDetails.Profile = v.GetProfile()
			userDetails.UserName = v.GetUserName()
			if v.ActivationDateValid {
				userDetails.ActivationDate = v.ActivationDate.AsTime()
			}
			usrs = append(usrs, &userDetails)
		}
		resp.UserDetails = usrs
		response = append(response, resp)
	}
	return
}

func createBatchNominativeUsers(allUsers []*v1.NominativeUser, batch int) (batchUsers [][]*v1.NominativeUser) {
	//batch := 100
	for i := 0; i < len(allUsers); i += batch {
		j := i + batch
		if j > len(allUsers) {
			j = len(allUsers)
		}
		batchUsers = append(batchUsers, allUsers[i:j]) // Process the batch.
	}
	return batchUsers
}

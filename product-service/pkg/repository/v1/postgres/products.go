package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/product-service/pkg/api/v1"
	gendb "optisam-backend/product-service/pkg/repository/v1/postgres/db"
	"strconv"
	"strings"
	"time"

	"github.com/tabbed/pqtype"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProductRepository ...
type ProductRepository struct {
	*gendb.Queries
	db *sql.DB
}

// NewProductRepository creates new Repository
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		Queries: gendb.New(db),
		db:      db,
	}
}

// ProductRepositoryTx ...
type ProductRepositoryTx struct {
	*gendb.Queries
	db *sql.Tx
}

// NewProductRepositoryTx ...
func NewProductRepositoryTx(db *sql.Tx) *ProductRepositoryTx {
	return &ProductRepositoryTx{
		Queries: gendb.New(db),
		db:      db,
	}
}

// UpsertProductTx upserts products/ linking data
func (p *ProductRepository) UpsertProductTx(ctx context.Context, req *v1.UpsertProductRequest, user string) error {
	var addApplications, deleteApplications, deleteEquipment []string
	var addEquipments []*v1.UpsertProductRequestEquipmentEquipmentuser
	var upsertPartialFlag bool

	// Create Transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	pt := NewProductRepositoryTx(tx)

	if req.Applications.GetOperation() == "add" {
		upsertPartialFlag = true
		addApplications = req.GetApplications().GetApplicationId()
	} else {
		deleteApplications = req.GetApplications().GetApplicationId()
	}

	if req.Equipments.GetOperation() == "add" {
		upsertPartialFlag = true
		addEquipments = req.Equipments.GetEquipmentusers()
	} else {
		deleteEquipmentUsers := req.GetEquipments().GetEquipmentusers()
		deleteEquipment = make([]string, len(deleteEquipmentUsers))
		for _, d := range deleteEquipmentUsers {
			deleteEquipment = append(deleteEquipment, d.GetEquipmentId())
		}
	}

	// Upsert Product Master
	if upsertPartialFlag {
		err = pt.UpsertProductPartial(ctx, gendb.UpsertProductPartialParams{Scope: req.GetScope(), Swidtag: req.GetSwidTag(), CreatedBy: user})
	} else {
		upsertReq := gendb.UpsertProductParams{
			Swidtag:         req.GetSwidTag(),
			ProductName:     req.GetName(),
			ProductCategory: req.GetCategory(),
			ProductEdition:  req.GetEdition(),
			ProductEditor:   req.GetEditor(),
			ProductVersion:  req.GetVersion(),
			Scope:           req.GetScope(),
			OptionOf:        req.GetOptionOf(),
			CreatedBy:       user,
			CreatedOn:       time.Now(),
			UpdatedBy:       sql.NullString{String: user},
			UpdatedOn:       time.Now(),
		}
		if req.ProductType == v1.Producttype_on_premise {
			upsertReq.ProductType = gendb.ProductTypeONPREMISE
		} else {
			upsertReq.ProductType = gendb.ProductTypeSAAS
		}
		err = pt.UpsertProduct(ctx, upsertReq)
	}
	if err != nil {
		tx.Rollback() // nolint: errcheck
		logger.Log.Error("failed to upsert product", zap.Error(err))
		return err
	}

	for _, app := range addApplications {
		error := pt.UpsertProductApplications(ctx, gendb.UpsertProductApplicationsParams{
			Swidtag:       req.GetSwidTag(),
			ApplicationID: app,
			Scope:         req.GetScope()})
		if error != nil {
			tx.Rollback() // nolint: errcheck
			logger.Log.Error("Failed to execute UpsertProductApplications", zap.Error(error))
			return error
		}
	}

	for _, equip := range addEquipments {
		error := pt.UpsertProductEquipments(ctx, gendb.UpsertProductEquipmentsParams{
			Swidtag:         req.GetSwidTag(),
			EquipmentID:     equip.EquipmentId,
			AllocatedMetric: equip.AllocatedMetrics,
			NumOfUsers:      sql.NullInt32{Int32: equip.AllocatedUsers, Valid: true},
			Scope:           req.GetScope()})
		if error != nil {
			tx.Rollback() // nolint: errcheck
			logger.Log.Error("Failed to execute UpsertProductEquipments", zap.Error(error))
			return error
		}
	}

	if len(deleteApplications) > 0 {
		err = pt.DeleteProductApplications(ctx, gendb.DeleteProductApplicationsParams{
			ProductID:     req.GetSwidTag(),
			ApplicationID: deleteApplications,
			Scope:         req.GetScope(),
		})
		if err != nil {
			tx.Rollback() // nolint: errcheck
			logger.Log.Error("failed to delete product-application", zap.Error(err))
			return err
		}
	}

	if len(deleteEquipment) > 0 {
		err = pt.DeleteProductEquipments(ctx, gendb.DeleteProductEquipmentsParams{
			ProductID:   req.GetSwidTag(),
			EquipmentID: deleteEquipment,
			// SCOPE BASED CHANGE
			Scope: req.GetScope(),
		})
		if err != nil {
			tx.Rollback() // nolint: errcheck
			logger.Log.Error("failed to delete product-equipments", zap.Error(err))
			return err
		}
	}

	tx.Commit() // nolint: errcheck
	return nil
}

// DropProductDataTx drops all the products data/ and linking in a particular scope
func (p *ProductRepository) DropProductDataTx(ctx context.Context, scope string, deletionType v1.DropProductDataRequestDeletionTypes) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	pt := NewProductRepositoryTx(tx)

	if deletionType == v1.DropProductDataRequest_ACQRIGHTS || deletionType == v1.DropProductDataRequest_FULL {
		if err := pt.DeleteAcqrightsByScope(ctx, scope); err != nil {
			tx.Rollback() // nolint: errcheck
			logger.Log.Error("failed to delete acqrights data", zap.Error(err))
			return err
		}
		if err := pt.DeleteAggregatedRightsByScope(ctx, scope); err != nil {
			tx.Rollback() // nolint: errcheck
			logger.Log.Error("failed to delete aggrights data", zap.Error(err))
			return err
		}
		if err := pt.DeleteSharedDataByScope(ctx, scope); err != nil {
			tx.Rollback() // nolint: errcheck
			logger.Log.Error("failed to delete shared data", zap.Error(err))
			return err
		}
	}
	if deletionType == v1.DropProductDataRequest_PARK || deletionType == v1.DropProductDataRequest_FULL {
		if err := pt.DeleteProductsByScope(ctx, scope); err != nil {
			tx.Rollback() // nolint: errcheck
			logger.Log.Error("failed to delete products data", zap.Error(err))
			return err
		}
		if err := pt.DeleteOverallComputedLicensesByScope(ctx, scope); err != nil {
			tx.Rollback() // nolint: errcheck
			logger.Log.Error("failed to delete compliance data", zap.Error(err))
			return err
		}
	}
	tx.Commit() // nolint: errcheck
	return nil
}

func filterNominativeUsers(req *v1.UpserNominativeUserRequest) (nomUsersValid, nomUsersInValid []*v1.NominativeUser, err error) {
	users := make(map[string]bool)
	fromFile := req.FileName != ""
	for _, v := range req.UserDetails {
		var startTime time.Time
		var err error
		nomUser := v1.NominativeUser{}
		if v.ActivationDate != "" {
			if len(v.ActivationDate) <= 10 {
				if strings.Contains(v.ActivationDate, "/") && len(v.ActivationDate) <= 8 {
					startTime, err = time.Parse("06/2/1", v.ActivationDate)
					if err != nil {
						startTime, err = time.Parse("2006/02/01", v.ActivationDate)
					}
				} else if strings.Contains(v.ActivationDate, "/") {
					startTime, err = time.Parse("2006/01/02", v.ActivationDate)
				} else {
					startTime, err = time.Parse("2006-01-02", v.ActivationDate)
					if err != nil {
						startTime, err = time.Parse("06-02-01", v.ActivationDate)
					}
					if err != nil {
						startTime, err = time.Parse("01-02-06", v.ActivationDate)
					}
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
			nomUser.ActivationDateString = v.ActivationDate
			if err == nil {
				nomUser.ActivationDate = timestamppb.New(startTime)
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
			if !fromFile {
				return nomUsersValid, nomUsersInValid, err
			}
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

func replaceSQL(stmt string, len int) string {
	beforVals := stmt[:strings.IndexByte(stmt, '?')-1]
	afterVals := stmt[strings.LastIndexByte(stmt, '?')+2:]
	vals := stmt[strings.IndexByte(stmt, '?')-1 : strings.LastIndexByte(stmt, '?')+2]
	vals += strings.Repeat(","+vals, len)
	stmt = beforVals + vals + afterVals
	n := 0
	for strings.IndexByte(stmt, '?') != -1 {
		n++
		param := "$" + strconv.Itoa(n)
		stmt = strings.Replace(stmt, "?", param, 1)
	}
	return stmt
}

func bulkUpsertNomAgg(nomUsers []*v1.NominativeUser, updatedBy, createdBy, swidTag, scope, editor string, agg_id int) (string, []interface{}) {
	valueArgs := make([]interface{}, 0, len(nomUsers)*11)
	t := time.Now()
	for _, user := range nomUsers {
		valueArgs = append(valueArgs, scope)
		valueArgs = append(valueArgs, agg_id)
		valueArgs = append(valueArgs, user.ActivationDate.AsTime())
		valueArgs = append(valueArgs, user.GetUserEmail())
		valueArgs = append(valueArgs, user.GetUserName())
		valueArgs = append(valueArgs, user.GetFirstName())
		valueArgs = append(valueArgs, user.GetProfile())
		valueArgs = append(valueArgs, createdBy)
		valueArgs = append(valueArgs, updatedBy)
		valueArgs = append(valueArgs, t)
		valueArgs = append(valueArgs, t)
	}
	smt := `INSERT INTO nominative_user (scope,aggregations_id,activation_date,user_email,user_name,first_name,profile,
		created_by,updated_by,created_at,updated_at) VALUES (?, ?, ?,?, ?, ?,?, ?, ?,?, ?) ON CONFLICT (aggregations_id,scope,user_email,profile)DO UPDATE SET 
		activation_date = EXCLUDED.activation_date,user_name = EXCLUDED.user_name,first_name =EXCLUDED.first_name,
		updated_by =EXCLUDED.updated_by ,updated_at =EXCLUDED.updated_at;`
	smt = replaceSQL(smt, len(nomUsers)-1)
	return smt, valueArgs
}
func bulkUpsertNomProduct(nomUsers []*v1.NominativeUser, updatedBy, createdBy, swidTag, scope, editor string, agg_id int) (string, []interface{}) {
	valueArgs := make([]interface{}, 0, len(nomUsers)*12)
	t := time.Now()
	for _, user := range nomUsers {
		valueArgs = append(valueArgs, scope)
		valueArgs = append(valueArgs, swidTag)
		valueArgs = append(valueArgs, user.ActivationDate.AsTime())
		valueArgs = append(valueArgs, user.GetUserEmail())
		valueArgs = append(valueArgs, user.GetUserName())
		valueArgs = append(valueArgs, user.GetFirstName())
		valueArgs = append(valueArgs, user.GetProfile())
		valueArgs = append(valueArgs, editor)
		valueArgs = append(valueArgs, createdBy)
		valueArgs = append(valueArgs, updatedBy)
		valueArgs = append(valueArgs, t)
		valueArgs = append(valueArgs, t)
	}
	smt := `INSERT INTO nominative_user (scope,swidtag,activation_date,user_email,user_name,first_name,profile,product_editor,
		created_by,updated_by,created_at,updated_at) VALUES (?, ?, ?,?, ?, ?,?, ?, ?,?, ?, ?) ON CONFLICT (swidtag,scope,user_email,profile)DO UPDATE SET
		activation_date = EXCLUDED.activation_date,user_name = EXCLUDED.user_name,first_name =EXCLUDED.first_name,
		updated_by =EXCLUDED.updated_by ,updated_at =EXCLUDED.updated_at`
	smt = replaceSQL(smt, len(nomUsers)-1)
	return smt, valueArgs
}

// UpsertNominativeUserTx upserts nominative user data
func (p *ProductRepository) UpsertNominativeUsersTx(ctx context.Context, req *v1.UpserNominativeUserRequest, updatedBy, createdBy, swidTag string) error {
	// Create Transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	//t := time.Now()
	pt := NewProductRepositoryTx(tx)

	if req.GetAggregationId() > 0 {
		validNomUsers, inValidNomUsers, err := filterNominativeUsers(req)
		if err != nil {
			tx.Rollback() // nolint: errcheck
			logger.Log.Error("failed to upsert product", zap.Error(err))
			return err
		}
		usersBatch := createBatchNominativeUsers(validNomUsers)
		for _, users := range usersBatch {
			query, args := bulkUpsertNomAgg(users, updatedBy, createdBy, swidTag, req.GetScope(), req.GetEditor(), int(req.AggregationId))
			_, err = tx.Exec(query, args...)
			if err != nil {
				tx.Rollback() // nolint: errcheck
				logger.Log.Error("failed to upsert nominative user", zap.Error(err))
				return err
			}
		}
		if req.FileName != "" {
			users, _ := json.Marshal(inValidNomUsers)
			r := gendb.InsertNominativeUserFileUploadDetailsParams{
				Scope:          req.GetScope(),
				AggregationsID: sql.NullInt32{Int32: req.AggregationId, Valid: true},
				UploadedBy:     updatedBy,
				RecordSucceed:  sql.NullInt32{Int32: int32(len(validNomUsers)), Valid: true},
				RecordFailed:   sql.NullInt32{Int32: int32(len(inValidNomUsers)), Valid: true},
				FileName:       sql.NullString{String: req.GetFileName(), Valid: true},
				SheetName:      sql.NullString{String: req.GetSheetName(), Valid: true},
				UploadID:       req.GetUploadId(),
			}
			if len(inValidNomUsers) > 0 && len(validNomUsers) > 0 {
				r.NominativeUsersDetails = pqtype.NullRawMessage{RawMessage: users, Valid: true}
				r.FileStatus = gendb.FileStatusPARTIAL
			} else if len(inValidNomUsers) > 0 && len(validNomUsers) == 0 {
				r.NominativeUsersDetails = pqtype.NullRawMessage{RawMessage: users, Valid: true}
				r.FileStatus = gendb.FileStatusFAILED
			} else if len(validNomUsers) > 0 && len(inValidNomUsers) == 0 {
				r.FileStatus = gendb.FileStatusFAILED
			}
			err = pt.InsertNominativeUserFileUploadDetails(ctx, r)
			if err != nil {
				tx.Rollback() // nolint: errcheck
				logger.Log.Error("failed to upsert product", zap.Error(err))
				return err
			}
		}
		return nil
	} else if swidTag != "" {
		validNomUsers, inValidNomUsers, err := filterNominativeUsers(req)
		if err != nil {
			tx.Rollback() // nolint: errcheck
			logger.Log.Error("failed to upsert product", zap.Error(err))
			return err
		}
		usersBatch := createBatchNominativeUsers(validNomUsers)
		for _, users := range usersBatch {
			query, args := bulkUpsertNomProduct(users, updatedBy, createdBy, swidTag, req.GetScope(), req.GetEditor(), int(req.AggregationId))
			logger.Log.Sugar().Debugw("Nominative users", "args", args)
			_, err = tx.Exec(query, args...)
			if err != nil {
				tx.Rollback() // nolint: errcheck
				logger.Log.Error("failed to upsert nominative user", zap.Error(err))
				return err
			}
		}
		if req.FileName != "" {
			users, _ := json.Marshal(inValidNomUsers)
			r := gendb.InsertNominativeUserFileUploadDetailsParams{
				Scope:         req.GetScope(),
				Swidtag:       sql.NullString{String: swidTag, Valid: true},
				ProductEditor: sql.NullString{String: req.Editor, Valid: true},
				UploadedBy:    updatedBy,
				RecordSucceed: sql.NullInt32{Int32: int32(len(validNomUsers)), Valid: true},
				RecordFailed:  sql.NullInt32{Int32: int32(len(inValidNomUsers)), Valid: true},
				FileName:      sql.NullString{String: req.GetFileName(), Valid: true},
				SheetName:     sql.NullString{String: req.GetSheetName(), Valid: true},
				UploadID:      req.GetUploadId(),
			}
			if len(inValidNomUsers) > 0 && len(validNomUsers) > 0 {
				r.NominativeUsersDetails = pqtype.NullRawMessage{RawMessage: users, Valid: true}
				r.FileStatus = gendb.FileStatusPARTIAL
			} else if len(inValidNomUsers) > 0 && len(validNomUsers) == 0 {
				r.NominativeUsersDetails = pqtype.NullRawMessage{RawMessage: users, Valid: true}
				r.FileStatus = gendb.FileStatusFAILED
			} else if len(validNomUsers) > 0 && len(inValidNomUsers) == 0 {
				r.FileStatus = gendb.FileStatusFAILED
			}
			err = pt.InsertNominativeUserFileUploadDetails(ctx, r)
			if err != nil {
				tx.Rollback() // nolint: errcheck
				logger.Log.Error("failed to upsert product", zap.Error(err))
				return err
			}
		}
	}
	//tx.Commit() // nolint: errcheck
	return nil
}
func createBatchNominativeUsers(allUsers []*v1.NominativeUser) (batchUsers [][]*v1.NominativeUser) {
	batch := 5000
	for i := 0; i < len(allUsers); i += batch {
		j := i + batch
		if j > len(allUsers) {
			j = len(allUsers)
		}
		batchUsers = append(batchUsers, allUsers[i:j]) // Process the batch.
	}
	return batchUsers
}

// UpsertConcurrentUserTx upserts concurrent user data
func (p *ProductRepository) UpsertConcurrentUserTx(ctx context.Context, req *v1.ProductConcurrentUserRequest, createdBy string) error {
	// Create Transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	fmt.Println("started")
	currentDateTime := time.Now()
	theDate := time.Date(currentDateTime.Year(), currentDateTime.Month(), 1, 00, 00, 00, 000, time.Local)

	pt := NewProductRepositoryTx(tx)
	if req.GetId() > 0 {
		pConUser, err := pt.GetConcurrentUserByID(ctx, gendb.GetConcurrentUserByIDParams{Scope: req.GetScope(), ID: req.GetId()})
		if err != nil {
			logger.Log.Error("failed to update product concurrent user, unable to get data", zap.Error(err))
			return err
		}
		purchaseDate := pConUser.PurchaseDate
		if purchaseDate.Month() == currentDateTime.Month() {
			theDate = purchaseDate
		}
	}
	if req.GetIsAggregations() && req.GetAggregationId() > 0 {
		upsertReq := gendb.UpsertAggregationConcurrentUserParams{}
		upsertReq.AggregationID = sql.NullInt32{Int32: req.GetAggregationId(), Valid: true}
		upsertReq.Swidtag = sql.NullString{String: "", Valid: false}
		upsertReq.Scope = req.GetScope()
		upsertReq.PurchaseDate = theDate
		upsertReq.NumberOfUsers = sql.NullInt32{Int32: req.GetNumberOfUsers(), Valid: true}
		upsertReq.Team = sql.NullString{String: req.GetTeam(), Valid: true}
		upsertReq.ProfileUser = sql.NullString{String: req.GetProfileUser(), Valid: true}
		upsertReq.IsAggregations = sql.NullBool{Bool: req.GetIsAggregations(), Valid: true}

		upsertReq.CreatedBy = createdBy
		upsertReq.UpdatedBy = sql.NullString{String: createdBy, Valid: true}
		upsertReq.CreatedOn = currentDateTime
		upsertReq.UpdatedOn = currentDateTime
		err = pt.UpsertAggregationConcurrentUser(ctx, upsertReq)
		if err != nil {
			tx.Rollback() // nolint: errcheck
			logger.Log.Error("failed to upsert product concurrent user", zap.Error(err))
			return err
		}
	} else {
		upsertReq := gendb.UpsertConcurrentUserParams{}
		upsertReq.Swidtag = sql.NullString{String: req.Swidtag, Valid: true}
		upsertReq.Scope = req.GetScope()
		upsertReq.PurchaseDate = theDate
		upsertReq.NumberOfUsers = sql.NullInt32{Int32: req.GetNumberOfUsers(), Valid: true}
		upsertReq.Team = sql.NullString{String: req.GetTeam(), Valid: true}
		upsertReq.ProfileUser = sql.NullString{String: req.GetProfileUser(), Valid: true}
		upsertReq.IsAggregations = sql.NullBool{Bool: req.GetIsAggregations(), Valid: true}

		upsertReq.CreatedBy = createdBy
		upsertReq.UpdatedBy = sql.NullString{String: createdBy, Valid: true}
		upsertReq.CreatedOn = currentDateTime
		upsertReq.UpdatedOn = currentDateTime
		err = pt.UpsertConcurrentUser(ctx, upsertReq)
		if err != nil {
			tx.Rollback() // nolint: errcheck
			logger.Log.Error("failed to upsert product concurrent user", zap.Error(err))
			return err
		}
	}

	tx.Commit() // nolint: errcheck
	return nil
}

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	v1 "optisam-backend/catalog-service/pkg/api/v1"
	"optisam-backend/catalog-service/pkg/repository/v1/postgres/db"
	repo "optisam-backend/catalog-service/pkg/repository/v1/postgres/db"
	"optisam-backend/common/optisam/logger"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProductRepository for Dgraph
type ProductCatalogRepository struct {
	*repo.Queries
	db *sql.DB
	r  *redis.Client
}

var repoObj *ProductCatalogRepository

// NewProductRepository creates new Repository
func SetProductCatalogRepository(db *sql.DB, rc *redis.Client) *ProductCatalogRepository {
	if repoObj == nil {
		repoObj = &ProductCatalogRepository{
			db:      db,
			Queries: repo.New(db),
			r:       rc}
	}
	return repoObj
}

// GetProductRepository give repo object
func GetProductCatalogRepository() (obj *ProductCatalogRepository) {
	return repoObj
}

// ProductRepositoryTx ...
type ProductCatalogRepositoryTx struct {
	*repo.Queries
	db *sql.Tx
}

// NewProductRepositoryTx ...
func NewProductCatalogRepositoryTx(db *sql.Tx) *ProductCatalogRepositoryTx {
	return &ProductCatalogRepositoryTx{
		Queries: repo.New(db),
		db:      db,
	}
}

func (p *ProductCatalogRepository) InsertProductTx(ctx context.Context, req *v1.Product) (res *v1.Product, err error) {

	productname := strings.Trim(req.Name, " ")
	if productname == "" {
		return nil, status.Error(codes.Internal, "product name should not be empty")
	}

	metrics, err := json.Marshal(req.Metrics)
	if err != nil {
		logger.Log.Error("v1/service - InsertProduct - Marshal Error Metrics")
		return res, status.Error(codes.Internal, err.Error())
	}
	vendors, err := json.Marshal(req.SupportVendors)
	if err != nil {
		logger.Log.Error("v1/service - InsertProduct - Marshal Error Vendors")
		return res, status.Error(codes.Internal, err.Error())
	}
	usefullinks, err := json.Marshal(req.UsefulLinks)
	if err != nil {
		logger.Log.Error("v1/service - InsertProduct - Marshal Error Vendors")
		return res, status.Error(codes.Internal, err.Error())
	}
	// if req.OpenSource.IsOpenSource && req.OpenSource.OpenLicences == "" {
	// 	logger.Log.Error("v1/service - InsertProduct - OpenLicences should not Empty")
	// 	return res, status.Error(codes.Internal, "OpenLicencesError")

	// }
	// if req.CloseSource.IsCloseSource && len(req.CloseSource.CloseLicences) == 0 {
	// 	logger.Log.Error("v1/service - InsertProduct - CloseLicences should not Empty")
	// 	return res, status.Error(codes.Internal, "CloseLicencesError")
	// }

	// closesource, err := json.Marshal(req.CloseSource.CloseLicences)
	// if err != nil {
	// 	logger.Log.Error("v1/service - InsertProduct - Marshal Error CloseLicences")
	// 	return res, status.Error(codes.Internal, err.Error())
	// }

	tx, err := p.db.BeginTx(ctx, nil)

	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return res, err
	}

	pt := NewProductCatalogRepositoryTx(tx)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	editor, err := pt.Queries.GetEditorCatalogName(ctx, req.EditorID)
	if err != nil {
		logger.Log.Error("service/v1 | InsertProduct ", zap.Any("Error while fecting record", err))
		return res, status.Error(codes.Internal, "Error while fetching editor record")
	}

	req.Id = uuid.New().String()
	err = pt.Queries.InsertProductCatalog(ctx, db.InsertProductCatalogParams{
		ID:                 req.Id,
		Editorid:           req.EditorID,
		Name:               productname,
		GenearlInformation: sql.NullString{String: req.GenearlInformation, Valid: true},
		ContractTips:       sql.NullString{String: req.ContracttTips, Valid: true},
		Metrics:            metrics,
		LicencesOpensource: sql.NullString{String: req.OpenSource.OpenLicences, Valid: true},
		SupportVendors:     vendors,
		Location:           db.LocationType(req.GetLocationType()),
		CreatedOn:          time.Now(),
		UpdatedOn:          time.Now(),
		Recommendation:     db.ProductCatalogRecommendation(req.GetRecommendation()),
		UsefulLinks:        usefullinks,
		SwidTagProduct:     sql.NullString{String: req.SwidtagProduct, Valid: true},
		EditorName:         editor.Name,
		OpensourceType:     db.OpensourceType(req.OpenSource.GetOpenSourceType()),
		Licensing:          db.ProductCatalogLicensing(req.GetLicensing()),
	})
	if err != nil {
		logger.Log.Error("service/v1 | Insert | Insert product", zap.Any("Error while saving record", err))
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return res, status.Error(codes.Internal, "Error while saving record, Duplicate Product Name for an editor")
		}
		return res, status.Error(codes.Internal, "Error while saving record")
	}
	var swidtag string

	//insert in Empty Version
	if len(req.Version) == 0 {
		swidtag = strings.ReplaceAll(strings.Join([]string{strings.Trim(req.Name, " "), editor.Name}, "_"), " ", "_")
		err = pt.Queries.InsertVersionCatalog(ctx, db.InsertVersionCatalogParams{
			ID:            uuid.New().String(),
			PID:           req.Id,
			SwidTagSystem: swidtag,
		})
		if err != nil {
			logger.Log.Error("service/v1 | InsertProduct | InsertVersion", zap.Any("Error while saving records", err))
			return res, status.Error(codes.InvalidArgument, "Error while saving record")
		}
		//insert in Version
	} else if req.Version != nil {
		for _, ver := range req.GetVersion() {
			verName := strings.Trim(ver.Name, " ")
			if verName == "" {
				return nil, status.Error(codes.Internal, "version name should not be empty")
			}
			ver.Id = uuid.New().String()
			swidtag = strings.ReplaceAll(strings.Join([]string{strings.Trim(req.Name, " "), editor.Name, strings.Trim(ver.Name, " ")}, "_"), " ", "_")
			err = pt.Queries.InsertVersionCatalog(ctx, db.InsertVersionCatalogParams{
				ID:             ver.Id,
				PID:            req.Id,
				SwidTagSystem:  swidtag,
				Name:           strings.Trim(ver.Name, " "),
				Recommendation: sql.NullString{String: ver.Recommendation, Valid: true},
				EndOfLife:      sql.NullTime{Time: ver.EndOfLife.AsTime(), Valid: true},
				EndOfSupport:   sql.NullTime{Time: ver.EndOfSupport.AsTime(), Valid: true},
				SwidTagVersion: sql.NullString{String: ver.SwidtagVersion, Valid: true},
			})
			if err != nil {
				logger.Log.Error("service/v1 | InsertProduct | InsertVersion", zap.Any("Error while saving records", err))
				return res, status.Error(codes.InvalidArgument, "Error while saving record")
			}
		}
	}
	return req, err
}

func (p *ProductCatalogRepository) UpdateProductTx(ctx context.Context, req *v1.Product) (err error) {
	if req.Id == "" {
		logger.Log.Error("v1/service - UpdateProduct - Id should not be empty")
		return status.Error(codes.Internal, "IdDoesNotExist")
	}

	productname := strings.Trim(req.Name, " ")
	if productname == "" {
		return status.Error(codes.Internal, "product name should not be empty")
	}

	metrics, err := json.Marshal(req.Metrics)
	if err != nil {
		logger.Log.Error("v1/service - UpdateProduct - Marshal Error Metrics")
		return status.Error(codes.Internal, err.Error())
	}
	vendors, err := json.Marshal(req.SupportVendors)
	if err != nil {
		logger.Log.Error("v1/service - UpdateProduct - Marshal Error Vendors")
		return status.Error(codes.Internal, err.Error())
	}
	usefullinks, err := json.Marshal(req.UsefulLinks)
	if err != nil {
		logger.Log.Error("v1/service - UpdateProduct - Marshal Error Vendors")
		return status.Error(codes.Internal, err.Error())
	}
	// if req.OpenSource.IsOpenSource && req.OpenSource.OpenLicences == "" {
	// 	logger.Log.Error("v1/service - UpdateProduct - OpenLicences should not Empty")
	// 	return status.Error(codes.Internal, "OpenLicencesError")

	// }
	// if req.CloseSource.IsCloseSource && len(req.CloseSource.CloseLicences) == 0 {
	// 	logger.Log.Error("v1/service - UpdateProduct - CloseLicences should not Empty")
	// 	return status.Error(codes.Internal, "CloseLicencesError")
	// }

	// closesource, err := json.Marshal(req.CloseSource.CloseLicences)
	// if err != nil {
	// 	logger.Log.Error("v1/service - UpdateProduct - Marshal Error CloseLicences")
	// 	return status.Error(codes.Internal, err.Error())
	// }

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	pt := NewProductCatalogRepositoryTx(tx)
	defer func() {
		if err != nil {
			trxerr := tx.Rollback()
			logger.Log.Error("service/v1 | InsertProduct |", zap.Any("Error while Rollbackerror record", trxerr))
		} else {
			trxerr := tx.Commit()
			logger.Log.Error("service/v1 | InsertProduct |", zap.Any("Error while Commit record", trxerr))
		}
	}()

	editor, err := pt.Queries.GetEditorCatalogName(ctx, req.EditorID)
	if err != nil {
		logger.Log.Error("service/v1 | UpdateProduct ", zap.Any("Error while fecting record", err))
		return status.Error(codes.Internal, "Error while fetching editor record")
	}

	err = pt.Queries.UpdateProductCatalog(ctx, db.UpdateProductCatalogParams{
		Name:               productname,
		Editorid:           req.EditorID,
		GenearlInformation: sql.NullString{String: req.GenearlInformation, Valid: true},
		ContractTips:       sql.NullString{String: req.ContracttTips, Valid: true},
		SupportVendors:     vendors,
		Metrics:            metrics,
		LicencesOpensource: sql.NullString{String: req.OpenSource.OpenLicences, Valid: true},
		Location:           db.LocationType(req.GetLocationType()),
		UpdatedOn:          time.Now(),
		Recommendation:     db.ProductCatalogRecommendation(req.GetRecommendation()),
		UsefulLinks:        usefullinks,
		SwidTagProduct:     sql.NullString{String: req.SwidtagProduct, Valid: true},
		EditorName:         editor.Name,
		OpensourceType:     db.OpensourceType(req.OpenSource.GetOpenSourceType()),
		ID:                 req.Id,
		Licensing:          db.ProductCatalogLicensing(req.GetLicensing()),
	})

	if err != nil {
		logger.Log.Error("service/v1 | UpdateProduct ", zap.Any("Error while saving records", err))
		return status.Error(codes.Internal, "Error while saving record")
	}
	dbversion, err := pt.Queries.GetVersionCatalogByPrductID(ctx, req.Id)
	if err != nil {
		logger.Log.Error("service/v1 | UpdateProduct ", zap.Any("Error while Get Product", err))
		return status.Error(codes.Internal, "Error while getting product")
	}
	versionIDs := make(map[string]bool, len(dbversion))
	for _, ver := range dbversion {
		versionIDs[ver.ID] = true
	}
	var swidtag string
	for _, ver := range req.GetVersion() {
		if ver.Id == "" {
			//insert
			verName := strings.Trim(ver.Name, " ")
			if verName == "" {
				return status.Error(codes.Internal, "version name should not be empty")
			}
			swidtag = strings.ReplaceAll(strings.Join([]string{strings.Trim(req.Name, " "), editor.Name, strings.Trim(ver.Name, " ")}, "_"), " ", "_")
			ver.Id = uuid.New().String()
			err = pt.Queries.InsertVersionCatalog(ctx, db.InsertVersionCatalogParams{
				ID:             ver.Id,
				PID:            req.Id,
				Name:           strings.Trim(ver.Name, " "),
				SwidTagSystem:  swidtag,
				Recommendation: sql.NullString{String: ver.Recommendation, Valid: true},
				EndOfLife:      sql.NullTime{Time: ver.EndOfLife.AsTime(), Valid: true},
				EndOfSupport:   sql.NullTime{Time: ver.EndOfSupport.AsTime(), Valid: true},
				SwidTagVersion: sql.NullString{String: ver.SwidtagVersion, Valid: true},
			})
			if err != nil {
				logger.Log.Error("service/v1 | UpdateProduct | UpdateProduct", zap.Any("Error while updating/inserting/deleting version", err))
				return status.Error(codes.Internal, "Error while saving record")
			}
		} else if versionIDs[ver.Id] {
			//update
			verName := strings.Trim(ver.Name, " ")
			if verName == "" {
				return status.Error(codes.Internal, "version name should not be empty")
			}
			swidtag = strings.ReplaceAll(strings.Join([]string{strings.Trim(req.Name, " "), editor.Name, strings.Trim(ver.Name, " ")}, "_"), " ", "_")
			err = pt.Queries.UpdateVersionCatalog(ctx, db.UpdateVersionCatalogParams{
				Name:           strings.Trim(ver.Name, " "),
				SwidTagSystem:  swidtag,
				Recommendation: sql.NullString{String: ver.Recommendation, Valid: true},
				EndOfLife:      sql.NullTime{Time: ver.EndOfLife.AsTime(), Valid: true},
				EndOfSupport:   sql.NullTime{Time: ver.EndOfSupport.AsTime(), Valid: true},
				SwidTagVersion: sql.NullString{String: ver.SwidtagVersion, Valid: true},
				ID:             ver.Id,
			})
			if err != nil {
				logger.Log.Error("service/v1 | UpdateProduct | UpdateProduct", zap.Any("Error while updating/inserting/deleting version", err))
				return status.Error(codes.Internal, "Error while saving record")
			}
			delete(versionIDs, ver.Id)
		}
	}
	if versionIDs != nil {
		for key := range versionIDs {
			//delete
			err = pt.Queries.DeleteVersionCatalog(ctx, key)
			if err != nil {
				logger.Log.Error("service/v1 | UpdateProduct | UpdateProduct", zap.Any("Error while updating/inserting/deleting version", err))
				return status.Error(codes.Internal, "Error while deleting version")
			}
		}
	}
	if len(req.GetVersion()) == 0 {
		swidtag = strings.ReplaceAll(strings.Join([]string{strings.Trim(req.Name, " "), editor.Name}, "_"), " ", "_")
		err = pt.Queries.InsertVersionCatalog(ctx, db.InsertVersionCatalogParams{
			ID:            uuid.New().String(),
			PID:           req.Id,
			SwidTagSystem: swidtag,
		})
		if err != nil {
			logger.Log.Error("service/v1 | UpdateProduct | UpdateVersion", zap.Any("Error while saving records", err))
			return status.Error(codes.InvalidArgument, "Error while inserting empty version record")
		}
	}
	dbversion, err = pt.Queries.GetVersionCatalogByPrductID(ctx, req.Id)
	if len(dbversion) == 0 {
		swidtag = strings.ReplaceAll(strings.Join([]string{strings.Trim(req.Name, " "), editor.Name}, "_"), " ", "_")
		err = pt.Queries.InsertVersionCatalog(ctx, db.InsertVersionCatalogParams{
			ID:            uuid.New().String(),
			PID:           req.Id,
			SwidTagSystem: swidtag,
		})
		if err != nil {
			logger.Log.Error("service/v1 | InsertProduct | InsertVersion", zap.Any("Error while saving records", err))
			return status.Error(codes.InvalidArgument, "Error while saving record")
		}

	}
	return nil
}

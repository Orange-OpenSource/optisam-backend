package rest

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/julienschmidt/httprouter"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/iam"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	rest_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/rest"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/config"
	v1kaf "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/kafka/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/repository/v1/postgres"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/service/v1"
	"go.opencensus.io/plugin/ochttp"
	"go.uber.org/zap"
)

// RunServer runs HTTP/REST gateway
func RunServer(ctx context.Context, config *config.Config, db *repo.ImportRepository, kafkaProducer *kafka.Producer, kafkaConsumer *kafka.Consumer) error {
	// get the verify key to validate jwt
	verifyKey, err := iam.GetVerifyKey(config.IAM)
	if err != nil {
		logger.Log.Fatal("Failed to get verify key")
	}

	// get Authorization Policy
	authZPolicies, err := iam.NewOPA(ctx, config.IAM.RegoPath)
	if err != nil {
		logger.Log.Fatal("Failed to Load RBAC policies", zap.Error(err))
	}
	router := httprouter.New()
	grpcClientMap, err := grpc.GetGRPCConnections(ctx, config.GRPCServers)
	if err != nil {
		logger.Log.Fatal("Failed to initialize GRPC client")
	}
	h := v1.NewImportServiceServer(grpcClientMap, config, db, kafkaProducer, kafkaConsumer)
	err = v1kaf.ImportConsumer(*h)
	if err != nil {
		logger.Log.Fatal("Failed to initialize import consumer")
	}
	// TODO add a import handler here
	router.POST("/api/v1/import/data", h.UploadDataHandler)
	router.POST("/api/v1/import/metadata", h.UploadMetaDataHandler)
	router.POST("/api/v1/import/config", h.CreateConfigHandler)
	router.PUT("/api/v1/import/config/:config_id", h.UpdateConfigHandler)
	router.POST("/api/v1/import/globaldata", h.UploadGlobalDataHandler)
	router.GET("/api/v1/import/download", h.DownloadFile)
	router.POST("/api/v1/import/download/nominative", h.DownloadFileNominativeUser)
	router.POST("/api/v1/import/upload", h.UploadFiles)
	router.POST("/api/v1/import/nominative/user", h.ImportNominativeUser)
	router.POST("/api/v1/import/uploadcatalogdata", h.UploadCatalogData)
	router.GET("/api/v1/import/nominative/users/fileupload", h.ListNominativeUserFileUploads)

	//api/v1/product/nominative/users/fileupload?scope=AAA&page_num=1&page_size=50&sort_by=name&sort_order=asc

	srv := &http.Server{
		Addr: ":" + config.HTTPPort,
		Handler: rest_middleware.AddCORS([]string{"*"},
			rest_middleware.AddLogger(logger.Log,
				rest_middleware.ValidateAuth(verifyKey,
					rest_middleware.ValidateAuthZ(authZPolicies, &ochttp.Handler{Handler: router})),
			)),
	}
	//   Handler:router,

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		_, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		_ = srv.Shutdown(ctx)
	}()

	logger.Log.Info("starting import-service - ", zap.String("port", config.HTTPPort))
	return srv.ListenAndServe()
}

module optisam-backend

require (
	contrib.go.opencensus.io/exporter/jaeger v0.2.0
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	contrib.go.opencensus.io/integrations/ocsql v0.1.3
	github.com/InVisionApp/go-health v2.1.0+incompatible
	github.com/InVisionApp/go-logger v1.0.1 // indirect
	github.com/dgraph-io/dgo/v2 v2.2.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v20.10.17+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/envoyproxy/protoc-gen-validate v0.1.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/go-playground/validator/v10 v10.2.0
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.2.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/lib/pq v1.9.0
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/open-policy-agent/opa v0.43.1
	github.com/opencensus-integrations/ocsql v0.1.3
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.0
	github.com/rogpeppe/go-internal v1.6.2 // indirect
	github.com/rs/cors v1.6.0
	github.com/rubenv/sql-migrate v0.0.0-20200212082348-64f95ea68aa3
	github.com/shopspring/decimal v1.2.0
	github.com/smartystreets/assertions v0.0.0-20190116191733-b6c0e53d7304 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.5.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.8.0
	github.com/tabbed/pqtype v0.1.1
	github.com/uber/jaeger-client-go v2.22.1+incompatible // indirect
	github.com/valyala/fasthttp v1.16.0 // indirect
	github.com/xuri/excelize/v2 v2.4.1
	go.opencensus.io v0.23.0
	go.uber.org/zap v1.17.0
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f
	google.golang.org/genproto v0.0.0-20220107163113-42d7afdf6368
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/oauth2.v3 v3.9.5
)

go 1.13

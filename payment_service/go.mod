module paymentservice

go 1.24.2

require (
	common_library v0.0.0-00010101000000-000000000000
	fileservice v0.0.0-00010101000000-000000000000
	github.com/georgysavva/scany/v2 v2.1.4
	github.com/golang-migrate/migrate/v4 v4.18.2
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/jackc/pgx/v5 v5.7.4
	github.com/pashagolub/pgxmock/v4 v4.7.0
	github.com/stretchr/testify v1.9.0
	go.uber.org/mock v0.5.1
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
	schedule_service v0.0.0-00010101000000-000000000000
	userservice v0.0.0-00010101000000-000000000000
)

replace common_library => ../common_library

replace userservice => ../user_service

replace fileservice => ../file_service

replace schedule_service => ../schedule_service

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250422160041-2d3770c4ea7f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

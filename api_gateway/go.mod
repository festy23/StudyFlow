module apigateway

go 1.24.2

require (
	common_library v0.0.0
	fileservice v0.0.0-00010101000000-000000000000
	github.com/go-chi/chi/v5 v5.2.1
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/redis/go-redis/v9 v9.7.3
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
	homework_service v0.0.0-00010101000000-000000000000
	paymentservice v0.0.0-00010101000000-000000000000
	schedule_service v0.0.0-00010101000000-000000000000
	userservice v0.0.0
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250422160041-2d3770c4ea7f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

replace (
	common_library => ../common_library
	fileservice => ../file_service
	homework_service => ../homework_service
	paymentservice => ../payment_service
	schedule_service => ../schedule_service
	userservice => ../user_service
)

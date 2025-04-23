module paymentservice

go 1.24.2

require (
	fileservice v0.0.0-00010101000000-000000000000
	go.uber.org/mock v0.5.1
	google.golang.org/grpc v1.72.0
	schedule_service v0.0.0-00010101000000-000000000000
	userservice v0.0.0-00010101000000-000000000000
)

replace commonlibrary => ../common_library

replace userservice => ../user_service

replace fileservice => ../file_service

replace schedule_service => ../schedule_service

require (
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250422160041-2d3770c4ea7f // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

module template

go 1.26.1

replace auth-service => ../services/auth-service

replace cms-service => ../services/cms-service

replace github.com/Mognus/go-grpc-crud => ../services/lib

require (
	auth-service v0.0.0-00010101000000-000000000000
	cms-service v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0
	github.com/joho/godotenv v1.5.1
	golang.org/x/net v0.52.0
	google.golang.org/grpc v1.80.0
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.11-20260415201107-50325440f8f2.1 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260427160629-7cedc36a6bc4 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

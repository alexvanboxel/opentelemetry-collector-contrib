module github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudpubsubreceiver

go 1.16

require (
	cloud.google.com/go/pubsub v1.11.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.28.0
	go.uber.org/zap v1.17.0
	google.golang.org/api v0.47.0
	google.golang.org/genproto v0.0.0-20210524142926-3e3a6030be83
	google.golang.org/grpc v1.38.0
)

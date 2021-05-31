module github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxcorrelationexporter

go 1.14

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.0.0-00010101000000-000000000000
	github.com/signalfx/signalfx-agent/pkg/apm v0.0.0-20201015185032-52a4f97df2a4
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.16.1-0.20201207152538-326931de8c32
	go.uber.org/zap v1.17.0
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk => ../../internal/splunk

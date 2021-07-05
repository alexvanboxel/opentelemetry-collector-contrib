module github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor

go 1.14

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	go.opencensus.io v0.22.5
	go.opentelemetry.io/collector v0.16.1-0.20201207152538-326931de8c32
	go.uber.org/zap v1.18.1
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig => ./../../internal/k8sconfig

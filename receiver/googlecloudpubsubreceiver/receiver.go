// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package googlecloudpubsubreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudpubsubreceiver"

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"strings"
	"sync"
	"time"

	pubsub "cloud.google.com/go/pubsub/apiv1"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudpubsubreceiver/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.uber.org/zap"
)

// https://cloud.google.com/pubsub/docs/reference/rpc/google.pubsub.v1#streamingpullrequest
type pubsubReceiver struct {
	logger             *zap.Logger
	obsrecv            *receiverhelper.ObsReport
	tracesConsumer     consumer.Traces
	metricsConsumer    consumer.Metrics
	logsConsumer       consumer.Logs
	userAgent          string
	config             *Config
	client             *pubsub.SubscriberClient
	tracesUnmarshaler  ptrace.Unmarshaler
	metricsUnmarshaler pmetric.Unmarshaler
	logsUnmarshaler    plog.Unmarshaler
	handler            *internal.StreamHandler
	startOnce          sync.Once
	telemetrySettings  component.TelemetrySettings
}

func (receiver *pubsubReceiver) setTracesConsumer(next consumer.Traces) {
	receiver.tracesConsumer = next
}
func (receiver *pubsubReceiver) setMetricsConsumer(next consumer.Metrics) {
	receiver.metricsConsumer = next
}
func (receiver *pubsubReceiver) setLogsConsumer(next consumer.Logs) {
	receiver.logsConsumer = next
}

type encoding int

const (
	unknown         encoding = iota
	otlpProtoTrace           = iota
	otlpProtoMetric          = iota
	otlpProtoLog             = iota
	rawTextLog               = iota
)

type compression int

const (
	uncompressed compression = iota
	gZip                     = iota
)

func (receiver *pubsubReceiver) handleLogStrings(ctx context.Context, rawData []byte, timestamp time.Time) error {
	if receiver.logsConsumer == nil {
		return nil
	}
	data := string(rawData)

	out := plog.NewLogs()
	logs := out.ResourceLogs()
	rls := logs.AppendEmpty()

	ills := rls.ScopeLogs().AppendEmpty()
	lr := ills.LogRecords().AppendEmpty()

	lr.Body().SetStr(data)
	lr.SetTimestamp(pcommon.NewTimestampFromTime(timestamp))
	return receiver.logsConsumer.ConsumeLogs(ctx, out)
}

func decompress(payload []byte, compression compression) ([]byte, error) {
	if compression == gZip {
		reader, err := gzip.NewReader(bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		return io.ReadAll(reader)
	}
	return payload, nil
}

func (receiver *pubsubReceiver) handleTrace(ctx context.Context, payload []byte, compression compression) error {
	payload, err := decompress(payload, compression)
	if err != nil {
		return err
	}
	otlpData, err := receiver.tracesUnmarshaler.UnmarshalTraces(payload)
	count := otlpData.SpanCount()
	if err != nil {
		return err
	}
	ctx = receiver.obsrecv.StartTracesOp(ctx)
	err = receiver.tracesConsumer.ConsumeTraces(ctx, otlpData)
	receiver.obsrecv.EndTracesOp(ctx, reportFormatProtobuf, count, err)
	return nil
}

func (receiver *pubsubReceiver) handleMetric(ctx context.Context, payload []byte, compression compression) error {
	payload, err := decompress(payload, compression)
	if err != nil {
		return err
	}
	otlpData, err := receiver.metricsUnmarshaler.UnmarshalMetrics(payload)
	count := otlpData.MetricCount()
	if err != nil {
		return err
	}
	ctx = receiver.obsrecv.StartMetricsOp(ctx)
	err = receiver.metricsConsumer.ConsumeMetrics(ctx, otlpData)
	receiver.obsrecv.EndMetricsOp(ctx, reportFormatProtobuf, count, err)
	return nil
}

func (receiver *pubsubReceiver) handleLog(ctx context.Context, payload []byte, compression compression) error {
	payload, err := decompress(payload, compression)
	if err != nil {
		return err
	}
	otlpData, err := receiver.logsUnmarshaler.UnmarshalLogs(payload)
	count := otlpData.LogRecordCount()
	if err != nil {
		return err
	}
	ctx = receiver.obsrecv.StartLogsOp(ctx)
	err = receiver.logsConsumer.ConsumeLogs(ctx, otlpData)
	receiver.obsrecv.EndLogsOp(ctx, reportFormatProtobuf, count, err)
	return nil
}

func detectEncoding(config *Config, attributes map[string]string) (encoding, compression) {
	otlpEncoding := unknown
	otlpCompression := uncompressed

	ceType := attributes["ce-type"]
	ceContentType := attributes["content-type"]
	if strings.HasSuffix(ceContentType, "application/protobuf") {
		switch ceType {
		case "org.opentelemetry.otlp.traces.v1":
			otlpEncoding = otlpProtoTrace
		case "org.opentelemetry.otlp.metrics.v1":
			otlpEncoding = otlpProtoMetric
		case "org.opentelemetry.otlp.logs.v1":
			otlpEncoding = otlpProtoLog
		}
	} else if strings.HasSuffix(ceContentType, "text/plain") {
		otlpEncoding = rawTextLog
	}

	if otlpEncoding == unknown && config.Encoding != "" {
		switch config.Encoding {
		case "otlp_proto_trace":
			otlpEncoding = otlpProtoTrace
		case "otlp_proto_metric":
			otlpEncoding = otlpProtoMetric
		case "otlp_proto_log":
			otlpEncoding = otlpProtoLog
		case "raw_text":
			otlpEncoding = rawTextLog
		}
	}

	ceContentEncoding := attributes["content-encoding"]
	if ceContentEncoding == "gzip" {
		otlpCompression = gZip
	}

	if otlpCompression == uncompressed && config.Compression != "" {
		if config.Compression == "gzip" {
			otlpCompression = gZip
		}
	}
	return otlpEncoding, otlpCompression
}

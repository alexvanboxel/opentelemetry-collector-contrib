// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package googlecloudpubsubreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudpubsubreceiver"

import (
	"context"
	"strings"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudpubsubreceiver/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
)

const (
	reportTransport      = "pubsub"
	reportFormatProtobuf = "protobuf"
)

type psReceiver interface {
	receiver.Traces
	receiver.Metrics
	receiver.Logs

	setTracesConsumer(consumer.Traces)
	setMetricsConsumer(consumer.Metrics)
	setLogsConsumer(consumer.Logs)
}

func NewFactory() receiver.Factory {
	f := &pubsubReceiverFactory{
		receivers: make(map[*Config]psReceiver),
	}
	return receiver.NewFactory(
		metadata.Type,
		f.CreateDefaultConfig,
		receiver.WithTraces(f.CreateTracesReceiver, metadata.TracesStability),
		receiver.WithMetrics(f.CreateMetricsReceiver, metadata.MetricsStability),
		receiver.WithLogs(f.CreateLogsReceiver, metadata.LogsStability),
	)
}

type pubsubReceiverFactory struct {
	receivers map[*Config]psReceiver
}

func (factory *pubsubReceiverFactory) CreateDefaultConfig() component.Config {
	return &Config{}
}

func (factory *pubsubReceiverFactory) ensureReceiver(params receiver.CreateSettings, config *Config) (psReceiver, error) {
	var receiver psReceiver
	receiver = factory.receivers[config]
	if receiver != nil {
		return receiver, nil
	}
	obsrecv, err := receiverhelper.NewObsReport(receiverhelper.ObsReportSettings{
		ReceiverID:             params.ID,
		Transport:              reportTransport,
		ReceiverCreateSettings: params,
	})
	if err != nil {
		return nil, err
	}
	psr := pubsubReceiver{
		logger:            params.Logger,
		obsrecv:           obsrecv,
		config:            config,
		telemetrySettings: params.TelemetrySettings,
	}
	if config.Mode == "push" {
		receiver = &pubsubPushReceiver{
			pubsubReceiver: &psr,
		}
	} else {
		receiver = &pubsubPullReceiver{
			pubsubReceiver: &psr,
			userAgent:      strings.ReplaceAll(config.UserAgent, "{{version}}", params.BuildInfo.Version),
		}
	}
	factory.receivers[config] = receiver
	return receiver, nil
}

func (factory *pubsubReceiverFactory) CreateTracesReceiver(
	_ context.Context,
	params receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Traces) (receiver.Traces, error) {

	if consumer == nil {
		return nil, component.ErrNilNextConsumer
	}
	err := cfg.(*Config).validateForTrace()
	if err != nil {
		return nil, err
	}
	receiver, err := factory.ensureReceiver(params, cfg.(*Config))
	if err != nil {
		return nil, err
	}
	receiver.setTracesConsumer(consumer)
	return receiver, nil
}

func (factory *pubsubReceiverFactory) CreateMetricsReceiver(
	_ context.Context,
	params receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Metrics) (receiver.Metrics, error) {

	if consumer == nil {
		return nil, component.ErrNilNextConsumer
	}
	err := cfg.(*Config).validateForMetric()
	if err != nil {
		return nil, err
	}
	receiver, err := factory.ensureReceiver(params, cfg.(*Config))
	if err != nil {
		return nil, err
	}
	receiver.setMetricsConsumer(consumer)
	return receiver, nil
}

func (factory *pubsubReceiverFactory) CreateLogsReceiver(
	_ context.Context,
	params receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Logs) (receiver.Logs, error) {

	if consumer == nil {
		return nil, component.ErrNilNextConsumer
	}
	err := cfg.(*Config).validateForLog()
	if err != nil {
		return nil, err
	}
	receiver, err := factory.ensureReceiver(params, cfg.(*Config))
	if err != nil {
		return nil, err
	}
	receiver.setLogsConsumer(consumer)
	return receiver, nil
}

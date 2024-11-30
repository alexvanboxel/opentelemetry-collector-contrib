// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package googlecloudlogentryencodingextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/googlecloudlogentryextension"

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding"
)

var _ encoding.LogsUnmarshalerExtension = (*otlpExtension)(nil)

type otlpExtension struct{}

func newExtension(_ *Config) (*otlpExtension, error) {
	return &otlpExtension{}, nil
}

func (ex *otlpExtension) UnmarshalLogs(buf []byte) (plog.Logs, error) {
	out := plog.NewLogs()

	var options Options
	resource, lr, err := TranslateLogEntry(buf, options)
	lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))

	if err != nil {
		return out, err
	}

	logs := out.ResourceLogs()
	rls := logs.AppendEmpty()
	resource.CopyTo(rls.Resource())

	ills := rls.ScopeLogs().AppendEmpty()
	lr.CopyTo(ills.LogRecords().AppendEmpty())
	return out, nil
}

func (ex *otlpExtension) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (ex *otlpExtension) Shutdown(_ context.Context) error {
	return nil
}

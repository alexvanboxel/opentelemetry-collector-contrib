// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package googlecloudlogentryencodingextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/googlecloudlogentryextension"
import (
	"go.opentelemetry.io/collector/component"
)

var _ component.ConfigValidator = (*Config)(nil)

type Config struct {
	HandleJSONPayloadAs  string `mapstructure:"handle_json_payload_as"`
	HandleProtoPayloadAs string `mapstructure:"handle_proto_payload_as"`
}

func (c *Config) Validate() error {
	return nil
}

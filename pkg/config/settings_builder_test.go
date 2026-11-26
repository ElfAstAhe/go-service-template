package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingsBuilder_WithPrefixed(t *testing.T) {
	tests := []struct {
		name       string
		keyPrefix  string
		flagPrefix string
		keySuffix  string
		flagSuffix string
		defaultVal any
		usage      string
		wantKey    string
		wantFlag   string
	}{
		{
			name:       "both prefix and suffix provided",
			keyPrefix:  "http.server",
			flagPrefix: "http-server",
			keySuffix:  "port",
			flagSuffix: "port",
			defaultVal: 8080,
			usage:      "HTTP listen port",
			wantKey:    "http.server.port",
			wantFlag:   "http-server-port",
		},
		{
			name:       "empty prefixes omit separators",
			keyPrefix:  "",
			flagPrefix: "",
			keySuffix:  "port",
			flagSuffix: "port",
			defaultVal: 8080,
			usage:      "HTTP listen port",
			wantKey:    "port",
			wantFlag:   "port",
		},
		{
			name:       "empty suffixes omit separators",
			keyPrefix:  "http.server",
			flagPrefix: "http-server",
			keySuffix:  "",
			flagSuffix: "",
			defaultVal: true,
			usage:      "enable server",
			wantKey:    "http.server",
			wantFlag:   "http-server",
		},
		{
			name:       "both empty return empty strings",
			keyPrefix:  "",
			flagPrefix: "",
			keySuffix:  "",
			flagSuffix: "",
			defaultVal: nil,
			usage:      "empty",
			wantKey:    "",
			wantFlag:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewSettingsBuilder(tt.keyPrefix, tt.flagPrefix).
				WithPrefixed(tt.keySuffix, tt.flagSuffix, tt.defaultVal, tt.usage).
				Build()

			require.Len(t, res, 1)
			assert.Equal(t, tt.wantKey, res[0].Key)
			assert.Equal(t, tt.wantFlag, res[0].Flag)
			assert.Equal(t, tt.defaultVal, res[0].Default)
			assert.Equal(t, tt.usage, res[0].Usage)
		})
	}
}

func TestSettingsBuilder_CompositeAndReuse(t *testing.T) {
	t.Run("WithTemplated formats key, flag and usage with accumulated prefix", func(t *testing.T) {
		res := NewSettingsBuilder("db.postgres", "pg").
			WithTemplated("%s.max_conns", "%s-max-conns", 25, "%s max pool connections").
			Build()

		require.Len(t, res, 1)
		assert.Equal(t, "db.postgres.max_conns", res[0].Key)
		assert.Equal(t, "pg-max-conns", res[0].Flag)
		assert.Equal(t, 25, res[0].Default)
		assert.Equal(t, "pg max pool connections", res[0].Usage)
	})

	t.Run("WithBuilder merges flat settings without parent prefix", func(t *testing.T) {
		metricsBuilder := NewSettingsBuilder("metrics", "metrics").
			WithPrefixed("port", "port", 9090, "metrics port")

		root := NewSettingsBuilder("app", "app").
			WithPrefixed("name", "name", "forge", "app name").
			WithBuilder(metricsBuilder).
			Build()

		require.Len(t, root, 2)
		assert.Equal(t, "app.name", root[0].Key)
		assert.Equal(t, "app-name", root[0].Flag)
		assert.Equal(t, "metrics.port", root[1].Key)
		assert.Equal(t, "metrics-port", root[1].Flag)
	})

	t.Run("WithSubBuilder nests multi-level prefixes and supports template reuse", func(t *testing.T) {
		// Базовый переиспользуемый шаблон коннектора
		connectorTemplate := NewSettingsBuilder("", "").
			WithPrefixed("url", "url", "amqp://localhost:5672/", "broker URL").
			WithTemplated("%s.timeout", "%s-timeout", "5s", "%s connect timeout")

		ordersRabbit := NewSettingsBuilder("orders", "orders").
			WithSubBuilder(connectorTemplate)

		billingRabbit := NewSettingsBuilder("billing", "billing").
			WithSubBuilder(connectorTemplate)

		appSettings := NewSettingsBuilder("rabbit", "rabbit").
			WithSubBuilder(ordersRabbit).
			WithSubBuilder(billingRabbit).
			Build()

		require.Len(t, appSettings, 4)

		assert.Equal(t, "rabbit.orders.url", appSettings[0].Key)
		assert.Equal(t, "rabbit-orders-url", appSettings[0].Flag)

		assert.Equal(t, "rabbit.orders.timeout", appSettings[1].Key)
		assert.Equal(t, "rabbit-orders-timeout", appSettings[1].Flag)
		assert.Equal(t, "rabbit-orders connect timeout", appSettings[1].Usage)

		assert.Equal(t, "rabbit.billing.url", appSettings[2].Key)
		assert.Equal(t, "rabbit-billing-url", appSettings[2].Flag)

		assert.Equal(t, "rabbit.billing.timeout", appSettings[3].Key)
		assert.Equal(t, "rabbit-billing-timeout", appSettings[3].Flag)
		assert.Equal(t, "rabbit-billing connect timeout", appSettings[3].Usage)
	})

	t.Run("nil inputs are safely ignored and Build is non-destructive", func(t *testing.T) {
		fixed := NewSetting("global.debug", "global-debug", false, "debug flag")

		sb := NewSettingsBuilder("svc", "svc").
			With(nil).
			WithBuilder(nil).
			WithSubBuilder(nil).
			With(fixed).
			WithPrefixed("host", "host", "0.0.0.0", "bind host")

		firstBuild := sb.Build()
		secondBuild := sb.Build()

		require.Len(t, firstBuild, 2)
		assert.Equal(t, firstBuild, secondBuild)
		assert.Equal(t, "global.debug", firstBuild[0].Key)
		assert.Equal(t, "svc.host", firstBuild[1].Key)
	})
}

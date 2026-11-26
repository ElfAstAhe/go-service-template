package config

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultAuthBasicConfig(t *testing.T) {
	cfg := NewDefaultAuthBasicConfig()

	require.NotNil(t, cfg)
	assert.Equal(t, DefaultAuthBasicEnabled, cfg.Enabled)
	assert.Equal(t, DefaultAuthBasicRealm, cfg.Realm)
	assert.Empty(t, cfg.Users)
}

func TestAuthBasicConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *AuthBasicConfig
		expectErr bool
	}{
		{
			name: "success - disabled config skips validation",
			cfg: &AuthBasicConfig{
				Enabled: false,
			},
			expectErr: false,
		},
		{
			name: "success - valid enabled config",
			cfg: NewAuthBasicConfig(
				true,
				"Admin Area",
				map[string]string{"admin": "secret123"},
			),
			expectErr: false,
		},
		{
			name: "failed - empty realm",
			cfg: NewAuthBasicConfig(
				true,
				"   ",
				map[string]string{"admin": "secret123"},
			),
			expectErr: true,
		},
		{
			name: "failed - empty users map",
			cfg: NewAuthBasicConfig(
				true,
				"Admin Area",
				nil,
			),
			expectErr: true,
		},
		{
			name: "failed - empty login in users",
			cfg: NewAuthBasicConfig(
				true,
				"Admin Area",
				map[string]string{"": "secret123"},
			),
			expectErr: true,
		},
		{
			name: "failed - empty password for user",
			cfg: NewAuthBasicConfig(
				true,
				"Admin Area",
				map[string]string{"admin": "   "},
			),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.expectErr {
				var cfgErr *errs.ConfigValidateError
				require.Error(t, err)
				assert.ErrorAs(t, err, &cfgErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestBuildDefaultAuthBasicSettings(t *testing.T) {
	settings := BuildDefaultAuthBasicSettings("auth.basic", "svc")

	require.NotEmpty(t, settings)
	assert.Len(t, settings, 2)
}

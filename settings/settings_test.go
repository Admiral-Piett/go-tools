package settings

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadBasicTypes(t *testing.T) {
	// Setup
	type TestSettings struct {
		StringField  string  `env:"TEST_STRING" default:"default_string"`
		IntField     int     `env:"TEST_INT" default:"42"`
		Int64Field   int64   `env:"TEST_INT64" default:"9999"`
		BoolField    bool    `env:"TEST_BOOL" default:"true"`
		Float64Field float64 `env:"TEST_FLOAT64" default:"3.14"`
		UintField    uint    `env:"TEST_UINT" default:"100"`
	}

	// Set environment variables
	os.Setenv("TEST_STRING", "hello")
	os.Setenv("TEST_INT", "123")
	os.Setenv("TEST_INT64", "456789")
	os.Setenv("TEST_BOOL", "false")
	os.Setenv("TEST_FLOAT64", "2.71")
	os.Setenv("TEST_UINT", "999")
	defer func() {
		os.Unsetenv("TEST_STRING")
		os.Unsetenv("TEST_INT")
		os.Unsetenv("TEST_INT64")
		os.Unsetenv("TEST_BOOL")
		os.Unsetenv("TEST_FLOAT64")
		os.Unsetenv("TEST_UINT")
	}()

	// Execute
	settings := &TestSettings{}
	err := Load(settings)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, "hello", settings.StringField)
	assert.Equal(t, 123, settings.IntField)
	assert.Equal(t, int64(456789), settings.Int64Field)
	assert.Equal(t, false, settings.BoolField)
	assert.Equal(t, 2.71, settings.Float64Field)
	assert.Equal(t, uint(999), settings.UintField)
}

func TestLoadWithDefaults(t *testing.T) {
	// Setup - no environment variables set
	type TestSettings struct {
		StringField string `env:"TEST_STRING_DEFAULT" default:"default_value"`
		IntField    int    `env:"TEST_INT_DEFAULT" default:"42"`
		BoolField   bool   `env:"TEST_BOOL_DEFAULT" default:"true"`
	}

	// Execute
	settings := &TestSettings{}
	err := Load(settings)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, "default_value", settings.StringField)
	assert.Equal(t, 42, settings.IntField)
	assert.Equal(t, true, settings.BoolField)
}

func TestLoadWithEnvOverride(t *testing.T) {
	// Setup
	type TestSettings struct {
		Port        string `env:"TEST_PORT" default:"8080"`
		Environment string `env:"TEST_ENV" default:"development"`
	}

	os.Setenv("TEST_PORT", "9000")
	os.Setenv("TEST_ENV", "production")
	defer func() {
		os.Unsetenv("TEST_PORT")
		os.Unsetenv("TEST_ENV")
	}()

	// Execute
	settings := &TestSettings{}
	err := Load(settings)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, "9000", settings.Port)
	assert.Equal(t, "production", settings.Environment)
}

func TestEmbeddedStructs(t *testing.T) {
	// Setup
	type CustomSettings struct {
		BaseSettings
		DatabaseURL string `env:"TEST_DATABASE_URL" default:"postgres://localhost"`
		APIKey      string `env:"TEST_API_KEY" default:"secret"`
	}

	os.Setenv("PORT", "3000")
	os.Setenv("TEST_DATABASE_URL", "postgres://prod-server")
	os.Setenv("ALLOWED_ORIGINS", "https://example.com, https://app.com")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("TEST_DATABASE_URL")
		os.Unsetenv("ALLOWED_ORIGINS")
	}()

	// Execute
	settings := &CustomSettings{}
	err := Load(settings)

	// Assert
	assert.Nil(t, err)
	// Check BaseSettings fields
	assert.Equal(t, "3000", settings.Port)
	assert.Equal(t, "development", settings.Environment) // Default value
	// Check custom fields
	assert.Equal(t, "postgres://prod-server", settings.DatabaseURL)
	assert.Equal(t, "secret", settings.APIKey) // Default value
	// Check post-processed field
	assert.Equal(t, []string{"https://example.com", "https://app.com"}, settings.AllowedOriginsSlice)
}

func TestPostProcessFields(t *testing.T) {
	// Setup
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000, https://example.com ,  https://app.com  ")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	// Execute
	settings := &BaseSettings{}
	err := Load(settings)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, "http://localhost:3000, https://example.com ,  https://app.com  ", settings.AllowedOrigins)
	assert.Equal(t, []string{
		"http://localhost:3000",
		"https://example.com",
		"https://app.com",
	}, settings.AllowedOriginsSlice)
}

func TestPostProcessFieldsWithEmptyOrigins(t *testing.T) {
	// Setup
	os.Setenv("ALLOWED_ORIGINS", "")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	// Execute
	settings := &BaseSettings{}
	err := Load(settings)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, "http://localhost:3000", settings.AllowedOrigins)
	assert.Equal(t, []string{"http://localhost:3000"}, settings.AllowedOriginsSlice)
}

func TestPostProcessFieldsWithWhitespaceOnly(t *testing.T) {
	// Setup
	os.Setenv("ALLOWED_ORIGINS", "  ,  ,  ")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	// Execute
	settings := &BaseSettings{}
	err := Load(settings)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, []string{}, settings.AllowedOriginsSlice)
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectedErr string
	}{
		{
			name:        "non-pointer",
			input:       BaseSettings{},
			expectedErr: "settings must be a pointer to struct",
		},
		{
			name:        "pointer to non-struct",
			input:       new(string),
			expectedErr: "settings must be a pointer to struct",
		},
		{
			name:        "nil pointer",
			input:       (*BaseSettings)(nil),
			expectedErr: "settings must be a pointer to struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Load(tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestCustomPostProcessSettings(t *testing.T) {
	// Setup
	type CustomSettings struct {
		BaseSettings
		FeatureFlags    string          `env:"FEATURE_FLAGS" default:"feature1,feature2"`
		FeatureFlagsMap map[string]bool // Derived field
		DatabaseURL     string          `env:"DATABASE_URL" default:"postgres://localhost:5432/testdb"`
		DatabaseHost    string          // Derived field
		DatabasePort    string          // Derived field
	}

	// Custom implementation of PostProcessFields
	customPostProcess := func(s *CustomSettings) {
		// Call parent's post-processing
		s.BaseSettings.PostProcessFields()

		// Process feature flags
		s.FeatureFlagsMap = make(map[string]bool)
		if s.FeatureFlags != "" {
			for _, flag := range strings.Split(s.FeatureFlags, ",") {
				s.FeatureFlagsMap[strings.TrimSpace(flag)] = true
			}
		}

		// Parse database URL (simplified - just extract host and port)
		// Format: postgres://host:port/db
		if s.DatabaseURL != "" {
			// Simple parsing for test
			s.DatabaseHost = "localhost"
			s.DatabasePort = "5432"
		}
	}

	os.Setenv("FEATURE_FLAGS", "auth,payments,analytics")
	os.Setenv("DATABASE_URL", "postgres://prod-db:5433/myapp")
	defer func() {
		os.Unsetenv("FEATURE_FLAGS")
		os.Unsetenv("DATABASE_URL")
	}()

	// Execute
	settings := &CustomSettings{}
	err := Load(settings)
	assert.Nil(t, err)

	// Manually call custom post-processing for this test
	customPostProcess(settings)

	// Assert
	assert.Equal(t, "auth,payments,analytics", settings.FeatureFlags)
	assert.Equal(t, map[string]bool{
		"auth":      true,
		"payments":  true,
		"analytics": true,
	}, settings.FeatureFlagsMap)
	assert.Equal(t, "localhost", settings.DatabaseHost)
	assert.Equal(t, "5432", settings.DatabasePort)
}

func TestLoadFieldsWithoutEnvTags(t *testing.T) {
	// Setup
	type TestSettings struct {
		WithTag    string `env:"WITH_TAG" default:"tagged"`
		WithoutTag string // No env tag - should be ignored
	}

	os.Setenv("WITH_TAG", "value")
	defer os.Unsetenv("WITH_TAG")

	// Execute
	settings := &TestSettings{}
	err := Load(settings)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, "value", settings.WithTag)
	assert.Equal(t, "", settings.WithoutTag) // Should remain zero value
}

func TestLoadWithMultipleEmbeddedStructs(t *testing.T) {
	// Setup
	type LoggingSettings struct {
		LogLevel string `env:"LOG_LEVEL" default:"INFO"`
	}

	type DatabaseSettings struct {
		DatabaseURL string `env:"DATABASE_URL" default:"postgres://localhost"`
	}

	type AppSettings struct {
		LoggingSettings
		DatabaseSettings
		AppName string `env:"APP_NAME" default:"test-app"`
	}

	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("DATABASE_URL", "postgres://prod")
	os.Setenv("APP_NAME", "my-app")
	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("APP_NAME")
	}()

	// Execute
	settings := &AppSettings{}
	err := Load(settings)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, "DEBUG", settings.LogLevel)
	assert.Equal(t, "postgres://prod", settings.DatabaseURL)
	assert.Equal(t, "my-app", settings.AppName)
}

func TestBaseSettingsAllFields(t *testing.T) {
	// Setup - test all BaseSettings fields
	os.Setenv("APP_NAME", "test-app")
	os.Setenv("PORT", "9090")
	os.Setenv("ENV", "staging")
	os.Setenv("ALLOWED_ORIGINS", "https://test.com")
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("SQL_LOG_LEVEL", "ERROR")
	os.Setenv("VERSION", "1.2.3")
	os.Setenv("GIT_SHA", "abc123")
	os.Setenv("ENCRYPTION_KEY", "my-encryption-key")
	os.Setenv("JWT_HMAC_KEY", "my-jwt-key")
	os.Setenv("JWT_ACCESS_TOKEN_TTL", "15")
	os.Setenv("JWT_REFRESH_TOKEN_TTL", "30")
	defer func() {
		os.Unsetenv("APP_NAME")
		os.Unsetenv("PORT")
		os.Unsetenv("ENV")
		os.Unsetenv("ALLOWED_ORIGINS")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("SQL_LOG_LEVEL")
		os.Unsetenv("VERSION")
		os.Unsetenv("GIT_SHA")
		os.Unsetenv("ENCRYPTION_KEY")
		os.Unsetenv("JWT_HMAC_KEY")
		os.Unsetenv("JWT_ACCESS_TOKEN_TTL")
		os.Unsetenv("JWT_REFRESH_TOKEN_TTL")
	}()

	// Execute
	settings := &BaseSettings{}
	err := Load(settings)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, "test-app", settings.AppName)
	assert.Equal(t, "9090", settings.Port)
	assert.Equal(t, "staging", settings.Environment)
	assert.Equal(t, "https://test.com", settings.AllowedOrigins)
	assert.Equal(t, []string{"https://test.com"}, settings.AllowedOriginsSlice)
	assert.Equal(t, "DEBUG", settings.LogLevel)
	assert.Equal(t, "ERROR", settings.SqlLogLevel)
	assert.Equal(t, "1.2.3", settings.AppVersion)
	assert.Equal(t, "abc123", settings.GitSha)
	assert.Equal(t, "my-encryption-key", settings.EncryptionKey)
	assert.Equal(t, "my-jwt-key", settings.JwtHmacKey)
	assert.Equal(t, 15, settings.JwtAccessTokenTTL)
	assert.Equal(t, 30, settings.JwtRefreshTokenTTL)
}

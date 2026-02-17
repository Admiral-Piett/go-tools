package settings

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// PostProcessSettingsInterface is an optional interface for settings that need post-processing.
// Implement this interface in your custom settings struct to perform any derived
// field calculations or validation after environment variables are loaded.
type PostProcessSettingsInterface interface {
	PostProcessFields()
}

// BaseSettings contains common fields most apps need.
// Apps should embed this struct and add their own custom fields.
//
// Example usage in your app:
//
//	type MyAppSettings struct {
//	    settings.BaseSettings
//	    // Add your custom fields
//	    DatabaseURL string `env:"DATABASE_URL" default:""`
//	    APIKey      string `env:"API_KEY"`
//	}
//
//	var GLOBAL_SETTINGS *MyAppSettings
//
//	func init() {
//	    GLOBAL_SETTINGS = &MyAppSettings{}
//	    if err := settings.Load(GLOBAL_SETTINGS); err != nil {
//	        log.Fatalf("Failed to load settings: %v", err)
//	    }
//	}
type BaseSettings struct {
	AppName     string `env:"APP_NAME" default:"my-app"`
	Port        string `env:"PORT" default:"8080"`
	Environment string `env:"ENV" default:"development"`

	// CORS configuration
	AllowedOrigins string `env:"ALLOWED_ORIGINS" default:"http://localhost:3000"`

	// Logging
	LogLevel    string `env:"LOG_LEVEL" default:"INFO"`
	SqlLogLevel string `env:"SQL_LOG_LEVEL" default:"WARNING"`

	SqlType string `env:"SQL_TYPE" default:"sqlite"`
	SqlUri  string `env:"SQL_URI" default:"path/to/sqlite.db"`

	// Versioning
	AppVersion string `env:"VERSION" default:"local.0"`
	GitSha     string `env:"GIT_SHA" default:"local"`

	// Authentication (JWT)
	EncryptionKey      string `env:"ENCRYPTION_KEY"`
	JwtHmacKey         string `env:"JWT_HMAC_KEY"`
	JwtAccessTokenTTL  int    `env:"JWT_ACCESS_TOKEN_TTL" default:"5"`
	JwtRefreshTokenTTL int    `env:"JWT_REFRESH_TOKEN_TTL" default:"10"`

	// Derived/Post-Processed fields
	AllowedOriginsSlice []string `json:"-"` // Derived field - populated by PostProcessFields
}

// PostProcessFields implements the PostProcessSettingsInterface
// This is called automatically by Load if your settings embed BaseSettings
func (s *BaseSettings) PostProcessFields() {
	// Parse allowed origins into slice
	parts := strings.Split(s.AllowedOrigins, ",")
	if len(parts) > 1 {
		parts := strings.Split(s.AllowedOrigins, ",")
		s.AllowedOriginsSlice = make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				s.AllowedOriginsSlice = append(s.AllowedOriginsSlice, trimmed)
			}
		}
	} else {
		s.AllowedOriginsSlice = parts
	}
}

// Load populates any struct with env tags using reflection, then calls PostProcessFields
// if the struct implements PostProcessSettingsInterface.
//
// The struct can contain embedded structs (like BaseSettings) and Load will
// recursively process all fields with `env` tags.
//
// Example:
//
//	type MySettings struct {
//	    settings.BaseSettings
//	    DatabaseURL string `env:"DATABASE_URL" default:"postgres://localhost"`
//	}
//	s := &MySettings{}
//	err := settings.Load(s)
func Load(settingsPtr interface{}) error {
	v := reflect.ValueOf(settingsPtr)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("settings must be a pointer to struct, got %T", settingsPtr)
	}

	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("settings must be a pointer to struct, got pointer to %v", elem.Kind())
	}

	loadFromEnv(elem)

	// If the settings implement PostProcessSettingsInterface, call PostProcessFields
	if processable, ok := settingsPtr.(PostProcessSettingsInterface); ok {
		processable.PostProcessFields()
	}

	return nil
}

// loadFromEnv recursively loads environment variables into struct fields
func loadFromEnv(structValue reflect.Value) {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		// Handle embedded structs recursively (like BaseSettings)
		if fieldType.Anonymous && field.Kind() == reflect.Struct {
			loadFromEnv(field)
			continue
		}

		// Skip fields without env tags
		envKey := fieldType.Tag.Get("env")
		if envKey == "" {
			continue
		}

		defaultValue := fieldType.Tag.Get("default")
		value := getEnvString(envKey, defaultValue)

		setField(field, value)
	}
}

// setField sets a struct field value based on its type
func setField(field reflect.Value, value string) {
	if !field.CanSet() {
		return
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			field.SetInt(intVal)
		}
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		if uintVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			field.SetUint(uintVal)
		}
	case reflect.Bool:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			field.SetBool(boolVal)
		}
	case reflect.Float32, reflect.Float64:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			field.SetFloat(floatVal)
		}
	}
}

// getEnvString returns the value of an environment variable or a default value
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

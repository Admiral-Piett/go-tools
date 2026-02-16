package settings

import (
    "os"
    "reflect"
    "strconv"
    "strings"

    log "github.com/sirupsen/logrus"
)

// App-level Use Pattern
//  var GLOBAL_SETTINGS *Settings
//
//  func init() {
//      GLOBAL_SETTINGS = NewSettings()
//  }

// Settings holds all configuration for the PolyTracker backend
type Settings struct {
    AppName string `env:"APP_NAME" default:"go-tools"`

    // Server configuration
    Port        string `env:"PORT" default:"8080"`
    Environment string `env:"ENV"  default:"development"`

    // CORS configuration
    AllowedOrigins string `env:"ALLOWED_ORIGINS" default:"http://localhost:3000"`

    // Database configuration
    DatabaseURL string `env:"DATABASE_URL" default:""`

    // Logging
    LogLevel    string `env:"LOG_LEVEL"     default:"INFO"`
    SqlLogLevel string `env:"SQL_LOG_LEVEL" default:"WARNING"`

    // Runtime fields (populated from processed values)
    AllowedOriginsSlice []string `json:"-"`

    AppVersion string `env:"VERSION" default:"local.0"`
    GitSha     string `env:"GIT_SHA" default:"local"`

    // Authentication
    EncryptionKey      string `env:"ENCRYPTION_KEY"`
    JwtHmacKey         string `env:"JWT_HMAC_KEY"`
    JwtAccessTokenTTL  int    `env:"JWT_ACCESS_TOKEN_TTL"  default:"5"`
    JwtRefreshTokenTTL int    `env:"JWT_REFRESH_TOKEN_TTL" default:"10"`

    // Test
    TestBoolean bool `env:"TEST_BOOLEAN" default:"true"`
}

// NewSettings creates and loads a new Settings instance
func NewSettings() *Settings {
    s := &Settings{}
    s.load()
    return s
}

func (s *Settings) load() {
    // Load from environment variables using reflection
    s.loadFromEnv()

    // Process derived fields
    s.processFields()

    log.WithFields(log.Fields{
        "port":        s.Port,
        "environment": s.Environment,
        "log_level":   s.LogLevel,
        "origins":     len(s.AllowedOriginsSlice),
    }).Info("Settings loaded from environment")
}

func (s *Settings) loadFromEnv() {
    keys := s.getStructFieldNames()
    t := reflect.TypeOf(*s)

    for _, key := range keys {
        field, ok := t.FieldByName(key)
        if !ok {
            continue
        }

        envKey := field.Tag.Get("env")
        if envKey == "" {
            continue
        }

        defaultValue := field.Tag.Get("default")
        value := getEnvString(envKey, defaultValue)
        s.set(key, value)
    }
}

func (s *Settings) processFields() {
    // Parse allowed origins into slice
    if s.AllowedOrigins != "" {
        origins := strings.Split(s.AllowedOrigins, ",")
        for i, origin := range origins {
            origins[i] = strings.TrimSpace(origin)
        }
        s.AllowedOriginsSlice = origins
    }
}

func (s *Settings) set(key string, value interface{}) {
    rv := reflect.ValueOf(s).Elem()
    field := rv.FieldByName(key)
    if !field.IsValid() || !field.CanSet() {
        return
    }

    switch field.Kind() {
    case reflect.String:
        if str, ok := value.(string); ok {
            field.SetString(str)
        }
    case reflect.Int, reflect.Int64:
        if str, ok := value.(string); ok {
            if intVal, err := strconv.ParseInt(str, 10, 64); err == nil {
                field.SetInt(intVal)
            }
        }
    case reflect.Bool:
        if str, ok := value.(string); ok {
            if boolVal, err := strconv.ParseBool(str); err == nil {
                field.SetBool(boolVal)
            }
        }
    }
}

func (s *Settings) getStructFieldNames() []string {
    var fields []string
    t := reflect.TypeOf(*s)

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        if field.Tag.Get("env") != "" {
            fields = append(fields, field.Name)
        }
    }

    return fields
}

// getEnvString returns the value of an environment variable or a default value
func getEnvString(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

# PolyTracker Backend

Cross-platform application for tracking US Congressional voting records with state-based filtering and real-time push notifications.

## Logging Format

The application uses a custom JSON formatter with guaranteed field ordering via the `app/logging` module.

### Field Order (guaranteed)
1. **message** - The log message
2. **severity** - Log level (debug, info, warn, error, fatal)  
3. **time** - ISO 8601 timestamp with milliseconds
4. **function** - Full module path of calling function
5. **error** - Error details (if present)
6. **[other fields]** - Additional fields sorted alphabetically

### Example Output
```json
{"message":"Server starting","severity":"info","time":"2025-08-04T10:30:00.123-07:00","function":"github.com/Admiral-Piett/go-tools/gin.Run","port":8080,"environment":"development"}
```

### Logging Module Structure
```
app/logging/
├── formatter.go  # Custom OrderedJSONFormatter
├── logging.go    # InitLogging function
└── access.go     # HTTP access log middleware
```

### Access Logging

HTTP requests are automatically logged using custom middleware that replaces Gin's default logger.

**Access Log Format**: `"<status_code> <method> <path>"`

**Example Access Log**:
```json
{"message":"200 GET /health","severity":"info","time":"2025-08-04T10:30:00.123-07:00","function":"github.com/Admiral-Piett/go-tools/gin/logging.AccessLogMiddleware.func1","status_code":200,"method":"GET","path":"/health","duration_ms":1.23,"client_ip":"127.0.0.1","user_agent":"curl/7.68.0"}
```

**Log Levels by Status Code**:
- 200-399: `info`
- 400-499: `warn` 
- 500+: `error`

### Usage
```go
import log "github.com/sirupsen/logrus"

// Logging is initialized in app.Run()
// All subsequent log calls use the ordered format

// Basic logging
log.Info("Server started")

// With fields
log.WithField("port", 8080).Info("Server listening")

// With multiple fields
log.WithFields(log.Fields{
    "user_id": "12345",
    "endpoint": "/v0/ping", 
}).Info("Request processed")

// Error logging
log.WithError(err).Error("Database connection failed")
```

## Running the Application

```bash
# Development
LOG_LEVEL=DEBUG go run cmd/main.go

# Production
LOG_LEVEL=INFO go run cmd/main.go
```

## Environment Variables

See `.env.example` for all available configuration options.

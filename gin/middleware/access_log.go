package middleware

import (
    "fmt"
    "time"

    "github.com/Admiral-Piett/go-tools/gin/utils"

    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
)

// AccessLogMiddleware creates a Gin middleware that logs HTTP requests in JSON format
//
// Example output:
// {"message":"200 GET /v0/ping","severity":"info","time":"2025-08-04T10:30:00.123-07:00","function":"github.com/Admiral-Piett/go-tools/gin/logging.AccessLogMiddleware.func1","status_code":200,"method":"GET","path":"/health","duration_ms":1.23,"client_ip":"127.0.0.1"}
func AccessLogMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // No need to log the health endpoint
        if "/health" == c.FullPath() {
            return
        }
        // Record start time
        start := time.Now()

        // Process request
        c.Next()

        // Calculate duration
        duration := time.Since(start)

        // Get response status code
        statusCode := c.Writer.Status()

        // Get request method and path
        method := c.Request.Method
        path := c.Request.URL.Path

        // Create the access log fields
        fields := log.Fields{
            "status_code": statusCode,
            "method":      method,
            "path":        path,
            "duration_ms": float64(
                duration.Nanoseconds(),
            ) / 1e6, // This will show partial milliseconds
            "client_ip": c.ClientIP(),
        }
        userId, ok := utils.GetUserId(c)
        if ok {
            fields["userId"] = userId
        }

        // Add query parameters if present
        if c.Request.URL.RawQuery != "" {
            fields["query"] = c.Request.URL.RawQuery
        }

        // Add user agent if present
        if userAgent := c.GetHeader("User-Agent"); userAgent != "" {
            fields["user_agent"] = userAgent
        }

        // Add request Id if present (useful for tracing)
        if requestID := c.GetHeader("X-Request-Id"); requestID != "" {
            fields["request_id"] = requestID
        }

        // Add content length if present
        if c.Request.ContentLength > 0 {
            fields["content_length"] = c.Request.ContentLength
        }

        // Create the log message in the format: "<status> <method> <path>"
        message := fmt.Sprintf("%d %s %s", statusCode, method, path)

        // Log with appropriate level based on status code
        logLevel := getLogLevel(statusCode)
        logEntry := log.WithFields(fields)

        switch logLevel {
        case log.ErrorLevel:
            logEntry.Error(message)
        case log.WarnLevel:
            logEntry.Warn(message)
        default:
            logEntry.Info(message)
        }
    }
}

// getLogLevel determines the appropriate log level based on HTTP status code
func getLogLevel(statusCode int) log.Level {
    switch {
    case statusCode >= 500:
        return log.ErrorLevel
    case statusCode >= 400:
        return log.WarnLevel
    default:
        return log.InfoLevel
    }
}

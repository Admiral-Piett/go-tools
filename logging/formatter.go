package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// OrderedJSONFormatter is a custom logrus formatter that guarantees field order
type OrderedJSONFormatter struct {
	// TimestampFormat sets the format used for marshaling timestamps.
	TimestampFormat string
}

// Format implements the logrus.Formatter interface
func (f *OrderedJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var buf bytes.Buffer

	// Start JSON object
	buf.WriteByte('{')

	// 1. Message field (always first)
	buf.WriteString(`"message":`)
	messageBytes, err := json.Marshal(entry.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}
	buf.Write(messageBytes)

	// 2. Severity field (always second)
	buf.WriteString(`,"severity":`)
	severityBytes, err := json.Marshal(strings.ToLower(entry.Level.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal severity: %v", err)
	}
	buf.Write(severityBytes)

	// 3. Time field (always third)
	buf.WriteString(`,"time":`)
	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.RFC3339
	}
	timeBytes, err := json.Marshal(entry.Time.Format(timestampFormat))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal time: %v", err)
	}
	buf.Write(timeBytes)

	// 4. Function field (always fourth) - if caller is reported
	if entry.HasCaller() {
		buf.WriteString(`,"function":`)
		functionName := f.formatCaller(entry.Caller)
		functionBytes, err := json.Marshal(functionName)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal function: %v", err)
		}
		buf.Write(functionBytes)
	}

	// 5. Error field (if present) - always after core fields
	if entry.Data != nil {
		if err, hasError := entry.Data[logrus.ErrorKey]; hasError {
			buf.WriteString(`,"error":`)
			err, ok := err.(error)
			if ok {
				errorBytes, err := json.Marshal(err.Error())
				if err != nil {
					return nil, fmt.Errorf("failed to marshal error.Error(): %v", err)
				}
				buf.Write(errorBytes)
			}

			errorBytes, err := json.Marshal(err)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal error: %v", err)
			}
			buf.Write(errorBytes)
		}
	}

	// 6. All other fields (sorted alphabetically for consistency)
	if entry.Data != nil {
		// Get all field keys except the error key (already handled)
		var keys []string
		for k := range entry.Data {
			if k != logrus.ErrorKey {
				keys = append(keys, k)
			}
		}

		// Sort keys for consistent output
		for i := 0; i < len(keys); i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[i] > keys[j] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}

		// Add sorted fields
		for _, k := range keys {
			buf.WriteByte(',')
			keyBytes, err := json.Marshal(k)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal key %s: %v", k, err)
			}
			buf.Write(keyBytes)
			buf.WriteByte(':')

			valueBytes, err := json.Marshal(entry.Data[k])
			if err != nil {
				return nil, fmt.Errorf(
					"failed to marshal value for key %s: %v",
					k,
					err,
				)
			}
			buf.Write(valueBytes)
		}
	}

	// Close JSON object and add newline
	buf.WriteByte('}')
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}

// formatCaller formats the caller information with full module path
func (f *OrderedJSONFormatter) formatCaller(caller *runtime.Frame) string {
	// Extract package and function name
	if caller == nil {
		return "unknown"
	}

	// Get the full function name which includes the package path
	funcName := caller.Function
	if funcName == "" {
		return "unknown"
	}

	// For methods, the function name includes the receiver type
	// e.g., "github.com/Admiral-Piett/polytracker-backend/src.(*Settings).load"
	// We want to keep the full path for better traceability
	return funcName
}

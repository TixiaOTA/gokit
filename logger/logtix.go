package logger

import (
	"os"
	"strings"
	"time"

	"github.com/TixiaOTA/gokit/loki"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap.Logger
type Logger struct {
	*zap.Logger
	lokiClient *loki.Client
}

// Config represents logger configuration
type Config struct {
	Level       string
	JSONOutput  bool
	FilePath    string
	Environment string
	Loki        *LokiConfig
}

// LokiConfig represents Loki-specific configuration
type LokiConfig struct {
	Enabled   bool
	URL       string
	BatchSize int
	BatchWait time.Duration
	Labels    map[string]string
}

// New creates a new logger with the given configuration
func New(config Config) *Logger {
	// Set up encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Determine encoder type
	var encoder zapcore.Encoder
	if config.JSONOutput {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Set up output
	var core zapcore.Core
	var lokiClient *loki.Client

	// Setup cores
	cores := []zapcore.Core{}

	// In development environment, always log to stdout
	if config.Environment == "development" {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), parseLevel(config.Level)))
	} else if config.FilePath != "" {
		// Use lumberjack for log rotation in non-development environments
		writer := &lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    100, // MB
			MaxBackups: 5,
			MaxAge:     30, // days
			Compress:   true,
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(writer), parseLevel(config.Level)))
	} else {
		// Fallback to stdout for any environment if no file path specified
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), parseLevel(config.Level)))
	}

	// Set up Loki client if enabled
	if config.Loki != nil && config.Loki.Enabled && config.Loki.URL != "" {
		lokiClient = loki.NewClient(loki.Config{
			URL:       config.Loki.URL,
			BatchSize: config.Loki.BatchSize,
			BatchWait: config.Loki.BatchWait,
			Labels:    config.Loki.Labels,
		})

		// Create a custom core that writes to both the primary core and Loki
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.AddSync(&lokiWriter{client: lokiClient}),
			parseLevel(config.Level),
		))
	}

	// Combine cores
	core = zapcore.NewTee(cores...)

	// Create logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		Logger:     zapLogger,
		lokiClient: lokiClient,
	}
}

// Default creates a default logger
func Default() *Logger {
	return New(Config{
		Level:      "info",
		JSONOutput: false,
		FilePath:   "",
	})
}

// With returns a new Logger with additional fields
func (l *Logger) With(fields ...zapcore.Field) *Logger {
	return &Logger{
		Logger:     l.Logger.With(fields...),
		lokiClient: l.lokiClient,
	}
}

// Named returns a new Logger with the given name
func (l *Logger) Named(name string) *Logger {
	return &Logger{
		Logger:     l.Logger.Named(name),
		lokiClient: l.lokiClient,
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Close stops the logger and any background goroutines
func (l *Logger) Close() error {
	if l.lokiClient != nil {
		l.lokiClient.Stop()
	}
	return l.Sync()
}

// Helper function to parse log level
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// lokiWriter implements zapcore.WriteSyncer for Loki
type lokiWriter struct {
	client *loki.Client
}

func (w *lokiWriter) Write(p []byte) (n int, err error) {
	// Extract level from the log message (this is a simple approach)
	// In a real implementation, you might want to parse the JSON log
	level := "info"
	if len(p) > 0 {
		lowered := string(p)
		switch {
		case contains(lowered, "debug"):
			level = "debug"
		case contains(lowered, "info"):
			level = "info"
		case contains(lowered, "warn"):
			level = "warn"
		case contains(lowered, "error"):
			level = "error"
		case contains(lowered, "fatal"):
			level = "fatal"
		}
	}

	w.client.Log(time.Now(), level, string(p))
	return len(p), nil
}

func (w *lokiWriter) Sync() error {
	// Sync is a no-op for Loki client
	return nil
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

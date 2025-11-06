package logs

import (
	"context"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type FileRotate struct {
	Filename   string `json:"filename" yaml:"filename"`
	MaxSize    int    `json:"maxSize" yaml:"maxSize"`       // MB
	MaxBackups int    `json:"maxBackups" yaml:"maxBackups"`
	MaxAge     int    `json:"maxAge" yaml:"maxAge"`         // days
	Compress   bool   `json:"compress" yaml:"compress"`
	Enable     bool   `json:"enable" yaml:"enable"`
}

type Config struct {
	Level            string     `json:"level" yaml:"level"`                   // debug|info|warn|error
	Encoding         string     `json:"encoding" yaml:"encoding"`             // json|console
	Development      bool       `json:"development" yaml:"development"`
	EnableCaller     bool       `json:"enableCaller" yaml:"enableCaller"`
	EnableStacktrace bool       `json:"enableStacktrace" yaml:"enableStacktrace"`
	Sampling         bool       `json:"sampling" yaml:"sampling"`
	OutputPaths      []string   `json:"outputPaths" yaml:"outputPaths"`       // e.g. ["stdout"]
	ErrorOutputPaths []string   `json:"errorOutputPaths" yaml:"errorOutputPaths"`
	ServiceName      string     `json:"serviceName" yaml:"serviceName"`
	Environment      string     `json:"environment" yaml:"environment"`       // dev|prod
	File             FileRotate `json:"file" yaml:"file"`
}

// ---- Global ----
var (
	globalLogger *zap.Logger
	globalSugar  *zap.SugaredLogger
)

func L() *zap.Logger {
	if globalLogger == nil {
		_ = InitGlobal(DevConfig()) // fallback
	}
	return globalLogger
}

func S() *zap.SugaredLogger {
	if globalSugar == nil {
		_ = InitGlobal(DevConfig())
	}
	return globalSugar
}

// InitGlobal builds a logger and sets it as global.
func InitGlobal(cfg Config, opts ...zap.Option) error {
	logger, err := New(cfg, opts...)
	if err != nil {
		return err
	}
	globalLogger = logger
	globalSugar = logger.Sugar()
	zap.ReplaceGlobals(logger)
	return nil
}

// New builds a zap logger from Config (no side effects).
func New(cfg Config, opts ...zap.Option) (*zap.Logger, error) {
	level := parseLevel(cfg.Level)

	encCfg := encoderConfig(cfg.Development)
	enc := func() zapcore.Encoder {
		if strings.ToLower(cfg.Encoding) == "console" {
			return zapcore.NewConsoleEncoder(encCfg)
		}
		return zapcore.NewJSONEncoder(encCfg)
	}()

	var cores []zapcore.Core

	// stdout/stderr
	if len(cfg.OutputPaths) == 0 {
		cfg.OutputPaths = []string{"stdout"}
	}
	for _, p := range cfg.OutputPaths {
		ws := zapcore.AddSync(writerFor(p))
		cores = append(cores, zapcore.NewCore(enc, ws, level))
	}

	// optional file rotate
	if cfg.File.Enable && cfg.File.Filename != "" {
		fw := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.File.Filename,
			MaxSize:    defaultInt(cfg.File.MaxSize, 128),
			MaxBackups: defaultInt(cfg.File.MaxBackups, 7),
			MaxAge:     defaultInt(cfg.File.MaxAge, 7),
			Compress:   cfg.File.Compress,
		})
		cores = append(cores, zapcore.NewCore(enc, fw, level))
	}

	core := zapcore.NewTee(cores...)

	// options
	if cfg.EnableCaller {
		opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(1))
	}
	if cfg.EnableStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}
	if cfg.Sampling {
		opts = append(opts, zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(c, time.Second, 100, 100)
		}))
	}

	fields := []zap.Field{
		zap.String("svc", cfg.ServiceName),
		zap.String("env", cfg.Environment),
	}
	opts = append(opts, zap.Fields(fields...))

	logger := zap.New(core, opts...)
	return logger, nil
}

// WithContext adds common IDs from context to logger (e.g., trace_id, user_id).
// You can adapt key names to your project’s context keys.
func WithContext(ctx context.Context) *zap.Logger {
	l := L()
	if ctx == nil {
		return l
	}
	type ctxKey string
	traceID, _ := ctx.Value(ctxKey("trace_id")).(string)
	userID, _ := ctx.Value(ctxKey("user_id")).(string)
	reqID, _ := ctx.Value(ctxKey("request_id")).(string)

	fields := make([]zap.Field, 0, 3)
	if traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	if reqID != "" {
		fields = append(fields, zap.String("request_id", reqID))
	}
	if userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}
	if len(fields) == 0 {
		return l
	}
	return l.With(fields...)
}

// ---- helpers ----

func DevConfig() Config {
	return Config{
		Level:            "debug",
		Encoding:         "console",
		Development:      true,
		EnableCaller:     true,
		EnableStacktrace: false,
		Sampling:         false,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		ServiceName:      "example-svc",
		Environment:      "dev",
	}
}

func ProdConfig() Config {
	return Config{
		Level:            "info",
		Encoding:         "json",
		Development:      false,
		EnableCaller:     true,
		EnableStacktrace: true,
		Sampling:         true,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		ServiceName:      "example-svc",
		Environment:      "prod",
		File: FileRotate{
			Enable:     true,
			Filename:   "./logs/app.log",
			MaxSize:    256,
			MaxBackups: 10,
			MaxAge:     14,
			Compress:   true,
		},
	}
}

func parseLevel(lv string) zapcore.LevelEnabler {
	switch strings.ToLower(lv) {
	case "debug":
		return zap.DebugLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

func encoderConfig(dev bool) zapcore.EncoderConfig {
	c := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     iso8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
	}
	if dev {
		c.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	return c
}

func iso8601TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339Nano))
}

func writerFor(path string) zapcore.WriteSyncer {
	switch strings.ToLower(path) {
	case "stdout":
		return zapcore.AddSync(os.Stdout)
	case "stderr":
		return zapcore.AddSync(os.Stderr)
	default:
		return zapcore.AddSync(&lumberjack.Logger{
			Filename:   path,
			MaxSize:    128,
			MaxBackups: 7,
			MaxAge:     7,
			Compress:   true,
		})
	}
}

func defaultInt(v, def int) int {
	if v > 0 {
		return v
	}
	return def
}
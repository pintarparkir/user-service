package logger

import "go.uber.org/zap/zapcore"

// logAt dispatches a log entry at the given level with structured fields.
func logAt(level zapcore.Level, ctx interface{}, msg string, fields map[string]interface{}) {
	if instance == nil {
		return
	}
	zf := fieldsFromMap(ctx, fields)
	switch level {
	case zapcore.InfoLevel:
		instance.Info(msg, zf...)
	case zapcore.WarnLevel:
		instance.Warn(msg, zf...)
	case zapcore.ErrorLevel:
		instance.Error(msg, zf...)
	case zapcore.FatalLevel:
		instance.Fatal(msg, zf...)
	}
}

// Info writes an info-level log entry with optional structured fields.
func Info(ctx interface{}, msg string, fields map[string]interface{}) {
	logAt(zapcore.InfoLevel, ctx, msg, fields)
}

// Warn writes a warn-level log entry.
func Warn(ctx interface{}, msg string, fields map[string]interface{}) {
	logAt(zapcore.WarnLevel, ctx, msg, fields)
}

// Error writes an error-level log entry.
func Error(ctx interface{}, msg string, fields map[string]interface{}) {
	logAt(zapcore.ErrorLevel, ctx, msg, fields)
}

// Fatal writes a fatal-level log entry and calls os.Exit(1).
func Fatal(ctx interface{}, msg string, fields map[string]interface{}) {
	logAt(zapcore.FatalLevel, ctx, msg, fields)
}

// Sync flushes pending log entries; call before exit.
func Sync() error {
	if instance == nil {
		return nil
	}
	return instance.Sync()
}

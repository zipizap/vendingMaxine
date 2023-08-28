package collection

import "go.uber.org/zap"

var slog *zap.SugaredLogger

func initSlog(slogger *zap.SugaredLogger) {
	slog = slogger
}

func slogGetLevel() (currentLevel string) {
	switch slog.Level() {
	case zap.DebugLevel:
		return "DEBUG"
	case zap.InfoLevel:
		return "INFO"
	case zap.ErrorLevel:
		return "ERROR"
	default:
		return "(unknown)"
	}
}

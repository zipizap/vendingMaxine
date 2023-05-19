package collection

import "go.uber.org/zap"

var slog *zap.SugaredLogger

func initSlog(slogger *zap.SugaredLogger) {
	slog = slogger
}

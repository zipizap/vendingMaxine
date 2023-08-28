package main

import (
	"os"
	"os/signal"
	"syscall"
	"vendingMaxine/packages/config"
	"vendingMaxine/packages/web"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var webF *web.Facilitator
var slog *zap.SugaredLogger
var slogConfig zap.Config

// newLevel one of "DEBUG" or "INFO" or "ERROR"
func slogSetLogLevel(newLevel string) {
	var newZapcoreLevel zapcore.Level
	switch newLevel {
	case "DEBUG":
		newZapcoreLevel = zapcore.DebugLevel
	case "INFO":
		newZapcoreLevel = zapcore.InfoLevel
	case "ERROR":
		newZapcoreLevel = zapcore.ErrorLevel
	default:
		slog.Panicf("Unrecognized level %s", newLevel)
	}
	slogConfig.Level.SetLevel(newZapcoreLevel)
}

func Init_slog() error {
	// Create a new logger configuration
	slogConfig = zap.NewProductionConfig()

	slogConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Set the encoder to use ISO8601 time format
	slogConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)      // Set log-level: DEBUG > INFO (default) > ERROR

	// logger, err := zap.NewProduction()
	logger, err := slogConfig.Build()
	if err != nil {
		return err
	}
	defer logger.Sync() // flushes buffer, if any
	slog = logger.Sugar()
	slog.Info("Initialized logger")
	return nil
}

func init() {
	err := Init_slog()
	if err != nil {
		panic(err)
	}

	slog.Info("Config - Initializing config from ./config.yaml > env-vars > defaultValues")
	var catalogDefaultName, catalogDefaultDirpath, dbFilepath, logLevel string
	{
		config.InitViperConfig("config", ".", "VD",
			map[string]string{
				"catalog.default.name":    "default-0-0-0",
				"catalog.default.dirpath": "./catalog-default-0-0-0",
				"db.filepath":             "./sqlite.db",
				"logs.loglevel":           "INFO",
			})
		catalogDefaultName = viper.GetString("catalog.default.name")
		catalogDefaultDirpath = viper.GetString("catalog.default.dirpath")
		dbFilepath = viper.GetString("db.filepath")
		logLevel = viper.GetString("logs.loglevel")
		slog.Infof("Config - catalog.name: %s", catalogDefaultName)
		slog.Infof("Config - catalog.dirpath: %s", catalogDefaultDirpath)
		slog.Infof("Config - db.filepath: %s", dbFilepath)
		slog.Infof("Config - logs.loglevel: %s", logLevel)
	}

	slog.Info("Config - setting loglevel to " + logLevel)
	slogSetLogLevel(logLevel)

	slog.Info("web.Facilitator - creating and initializing")
	{
		webF, err = web.NewFacilitator()
		if err != nil {
			slog.Fatal(err)
		}
		webF.InitSetup(dbFilepath, catalogDefaultName, catalogDefaultDirpath, slog)
	}
}

func wait4SignalsForever() {
	// create a channel to receive signals
	sigChan := make(chan os.Signal, 1)
	// notify the channel for SIGINT and SIGTERM signals
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// wait for a signal to be received
	signalReceived := <-sigChan

	// exit with status code 1
	slog.Fatalf("Aborting - received signal %s", signalReceived)
}

func main() {
	webServerListeningAddr := ":8080"
	slog.Infof("Starting Htpp-web server at %s", webServerListeningAddr)
	go webF.StartServer(webServerListeningAddr)

	slog.Infoln("Monitoring signals SIGINT, SIGTERM")
	wait4SignalsForever()
}

package main

import (
	"os"

	"github.com/hodlmee/shrimpy-go/shrimpy"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// APIConfig reads config from environment variables
type APIConfig struct {
	URL    string `envconfig:"shrimpy_url"`
	Key    string `envconfig:"shrimpy_key"`
	Secret string `envconfig:"shrimpy_secret"`
}

func main() {

	// create a logger instance
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "ts"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), consoleDebugging, zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl < zapcore.ErrorLevel && lvl > zapcore.DebugLevel
		})),
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), consoleErrors, zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		})),
	)
	logger := zap.New(core, zap.AddCaller())
	defer func() { _ = logger.Sync() }()

	// read the config from environment variables
	var config APIConfig
	if err := envconfig.Process("", &config); err != nil {
		logger.Fatal("error reading environment variables")
	}

	// validate the environment variables have been set
	if config.URL == "" {
		logger.Fatal("SHRIMPY_URL is not set")
	}
	if config.Key == "" {
		logger.Fatal("SHRIMPY_KEY is not set")
	}
	if config.Secret == "" {
		logger.Fatal("SHRIMPY_SECRET is not set")
	}

	// initialize the shrimpy client and list accounts
	shrimpyClient := shrimpy.MustNewShrimpy(config.URL, config.Key, config.Secret, logger)
	accounts, err := shrimpyClient.GetAccounts()
	if err != nil {
		logger.Fatal("error retrieving accounts", zap.Error(err))
	}
	logger.Info("successfully found accounts", zap.Any("accounts", accounts))
}

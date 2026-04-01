package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log  *zap.SugaredLogger
	once sync.Once
)

func Init(isProduction bool) {
	once.Do(func() {
		var config zap.Config

		if isProduction {
			config = zap.NewProductionConfig()
		} else {
			config = zap.NewDevelopmentConfig()
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}

		baseLogger, err := config.Build()
		if err != nil {
			panic(err)
		}

		Log = baseLogger.Sugar()
	})
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

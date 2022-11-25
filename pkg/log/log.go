package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

func init() {
	configuration := zap.NewDevelopmentConfig()
	configuration.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	logger, _ := configuration.Build(zap.IncreaseLevel(zap.InfoLevel),
		zap.AddCallerSkip(1))
	log = logger.Sugar()
}

func Close() {
	_ = log.Sync()
}

func Debug(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Info(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Warn(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Error(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

package logging

import (
	"io"
	"strings"
	"time"
	"unilab-backend/setting"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	logger.Fatalf(template, args...)
}

func Panic(args ...interface{}) {
	logger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	logger.Panicf(template, args...)
}

// dpanic only shutdown app in debug mode
func DPanic(args ...interface{}) {
	logger.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	logger.DPanicf(template, args...)
}

func GetWriter(filename string) io.Writer {
	// 生成rotatelogs
	hook, err := rotatelogs.New(
		strings.ReplaceAll(filename, ".log", "")+"-%Y%m%d.log", // 实际生成的文件名 filename-YYmmdd.log
		rotatelogs.WithLinkName(filename),                      // filename.log是指向最新日志的链接
		rotatelogs.WithMaxAge(time.Hour*24*7),                  // 仅保存7天内的日志
		rotatelogs.WithRotationTime(time.Hour*24),              // 每一天(24:00)分割一次日志
	)
	if err != nil {
		panic(err)
	}
	return hook
}

func init() {
	// log基本格式
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 输出level序列化为全大写字符串，如 INFO DEBUG ERROR
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	encoderConfig.CallerKey = "file"
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	// 实现判断日志等级的interface
	infoLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level < zapcore.WarnLevel
	})
	warnLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapcore.WarnLevel
	})
	infoWriter := GetWriter(setting.RuntimeRootDir + "logs/backend.log")
	warnWriter := GetWriter(setting.RuntimeRootDir + "logs/backend_error.log")
	// 创建logger
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(infoWriter), infoLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(warnWriter), warnLevel),
	)
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
}

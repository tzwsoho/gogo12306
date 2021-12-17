package logger

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger
var logLevel zap.AtomicLevel

func Init(isDevEnv bool, logFilepath, logLevelName, serviceName string, splitMBSize, keepDays int) {
	var (
		err error
		ws  zapcore.WriteSyncer
	)
	if isDevEnv {
		// if logger, err = zap.NewDevelopment(zap.WithCaller(true)); err != nil {
		// 	panic("Init dev logger err " + err.Error())
		// }

		ws = zapcore.AddSync(os.Stdout)
	} else {
		// if logger, err = zap.NewProduction(zap.WithCaller(true)); err != nil {
		// 	panic("Init pro logger err " + err.Error())
		// }

		ws = zapcore.AddSync(&lumberjack.Logger{
			Filename:   logFilepath,
			MaxBackups: 0,
			Compress:   true,
			MaxAge:     keepDays,
			MaxSize:    splitMBSize,
		})
	}

	//////////////////////////////////////////////////////////////////////////////////////////////

	enc := zap.NewProductionEncoderConfig()
	enc.TimeKey = "@timestamp"
	enc.MessageKey = "_msg"
	enc.LevelKey = "_level"
	enc.NameKey = "_logger"
	enc.CallerKey = "_caller"
	enc.StacktraceKey = "_stacktrace"
	enc.EncodeTime = zapcore.ISO8601TimeEncoder

	//////////////////////////////////////////////////////////////////////////////////////////////

	logLevel = zap.NewAtomicLevel()
	if err = SetLogLevel(logLevelName); err != nil {
		panic(err)
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(enc), ws, logLevel)
	logger = zap.New(core, zap.Fields(zap.String("_serviceName", serviceName)))

	logger.Info("Logger initialized", zap.String("logLevel", logLevelName))
}

func Close() {
	if logger != nil {
		logger.Sync()
	}
}

func SetLogLevel(logLevelName string) (err error) {
	var logLvl zapcore.Level
	if err = logLvl.Set(logLevelName); err != nil {
		return errors.New("invalid logLevelName: " + logLevelName)
	}

	logLevel.SetLevel(logLvl)
	return
}

func getCallStack() (fs []zap.Field) {
	if pc, fn, ln, ok := runtime.Caller(2); ok {
		if fnc := runtime.FuncForPC(pc); fnc != nil {
			_ = fn
			fs = append(fs,
				// zap.Uintptr("pc", pc),
				// zap.String("file", fn),
				zap.Int("line", ln),
				zap.String("func", fnc.Name()),
			)

			return fs
		}
	}

	return
}

func getFullCallStack() (fs []zap.Field) {
	var (
		n   int = 10
		pcs []uintptr
	)
	for {
		old := n
		pcs = make([]uintptr, n)
		n = runtime.Callers(3, pcs[:])
		if old > n {
			break
		} else {
			n *= 2
		}
	}

	for i := 0; i < n; i++ {
		if fnc := runtime.FuncForPC(pcs[i]); fnc != nil {
			_, ln := fnc.FileLine(pcs[i])
			fs = append(fs,
				// zap.Uintptr("pc", pcs[n]),
				// zap.String("file", fn),
				zap.Int(fmt.Sprintf("line_%d", i), ln),
				zap.String(fmt.Sprintf("func_%d", i), fnc.Name()),
			)
		}
	}

	return
}

func Debug(msg string, fields ...zap.Field) {
	fields = append(fields, getCallStack()...)
	logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	// fields = append(fields, getCallStack()...)
	fields = append(fields, getFullCallStack()...)
	logger.Error(msg, fields...)
}

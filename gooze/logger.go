package gooze

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/soryetong/gooze-starter/pkg/gzutil"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var rotationSchedulerProcess *rotationScheduler

type iLog struct {
	*zap.Logger
}

func initILog() {
	viper.SetDefault("Log.Path", "./logs")
	viper.SetDefault("Log.Mode", "both")
	viper.SetDefault("Log.Recover", false)
	viper.SetDefault("Log.MaxSize", 100)
	viper.SetDefault("Log.MaxBackups", 3)
	viper.SetDefault("Log.MaxAge", 7)
	viper.SetDefault("Log.Compress", true)
	viper.SetDefault("Log.Logrotate", true)
	newILog()

	// 日志轮转
	if viper.GetBool("Log.Logrotate") || viper.GetBool("Log.Recover") {
		rotationSchedulerProcess = newRotationScheduler(func() {
			newILog()
		})
	}
}

func newILog() {
	Log = &iLog{
		Logger: zap.New(
			getCore(),
			zap.AddCaller(),
			zap.AddCallerSkip(0),
			zap.AddStacktrace(zap.ErrorLevel),
		),
	}
}

func (l *iLog) WithCtx(ctx context.Context) *iLog {
	var fields []zap.Field

	if traceId := ctx.Value("trace_id"); traceId != nil {
		if s, ok := traceId.(string); ok && s != "" {
			fields = append(fields, zap.String("trace_id", s))
		}
	}

	if source := ctx.Value("source"); source != nil {
		if s, ok := source.(string); ok && s != "" {
			fields = append(fields, zap.String("source", s))
		}
	}

	if len(fields) > 0 {
		return l.With(fields...)
	}

	return l
}

func (l *iLog) With(fields ...zap.Field) *iLog {
	return &iLog{
		Logger: l.Logger.With(fields...),
	}
}

// Core is a minimal, fast logger interface. It's designed for library authors
// to wrap in a more user-friendly API.
// only use infoLevel、errorLevel. want update can change == to > or >= or <= or <
func getCore() zapcore.Core {
	path := viper.GetString("Log.Path")
	mode := viper.GetString("Log.Mode")
	doRecover := viper.GetBool("Log.Recover")

	encoder := zapcore.NewJSONEncoder(getEncoderConfig())
	debugWrite := getLogWriter(path, mode, doRecover, zapcore.DebugLevel)
	infoWrite := getLogWriter(path, mode, doRecover, zapcore.InfoLevel)
	warnWrite := getLogWriter(path, mode, doRecover, zapcore.WarnLevel)
	errorWrite := getLogWriter(path, mode, doRecover, zapcore.ErrorLevel)
	fatalWrite := getLogWriter(path, mode, doRecover, zapcore.FatalLevel)
	debugLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.DebugLevel
	})
	infoLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.InfoLevel
	})
	warnLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.WarnLevel
	})
	errorLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.ErrorLevel
	})
	fatalLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.FatalLevel
	})
	return zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(debugWrite), debugLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(infoWrite), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(warnWrite), warnLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(errorWrite), errorLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(fatalWrite), fatalLevel),
	)
}

// A WriteSyncer is an io.Writer that can also flush any buffered data. Note
// that *os.File (and thus, os.Stderr and os.Stdout) implement WriteSyncer.
func getLogWriter(path, mode string, recover bool, level zapcore.Level) zapcore.WriteSyncer {
	maxSize := viper.GetInt("Log.MaxSize")
	maxBackups := viper.GetInt("Log.MaxBackups")
	maxAge := viper.GetInt("Log.MaxAge")
	compress := viper.GetBool("Log.Compress")
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	fileName := fmt.Sprintf("%s%s/%s.log", path, time.Now().Format("2006-01-02"), level)
	var fileWriter io.Writer
	if recover {
		fileWriter = newCustomWrite(fileName, maxSize, maxBackups, maxAge, compress)
	} else {
		fileWriter = &lumberjack.Logger{
			Filename:   fileName,
			MaxSize:    maxSize,    // 单文件最大容量, 单位是MB
			MaxBackups: maxBackups, // 最大保留过期文件个数
			MaxAge:     maxAge,     // 保留过期文件的最大时间间隔, 单位是天
			Compress:   compress,   // 是否需要压缩滚动日志, 使用的gzip压缩
			LocalTime:  true,       // 是否使用计算机的本地时间, 默认UTC
		}
	}
	var writer zapcore.WriteSyncer
	switch mode {
	case "file":
		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileWriter))
	case "console":
		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
	case "close":
		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(io.Discard))
	default:
		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter))
	}

	return writer
}

// An EncoderConfig allows users to configure the concrete encoders supplied by zap core
func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "file_line",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
}

// PrimitiveArrayEncoder is the subset of the ArrayEncoder interface that deals
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// rotationScheduler 日志轮转调度器
type rotationScheduler struct {
	done       chan struct{}
	rotateFunc func()
	wg         sync.WaitGroup
}

// newRotationScheduler 创建新的日志轮转调度器
func newRotationScheduler(rotateFunc func()) *rotationScheduler {
	rs := &rotationScheduler{
		done:       make(chan struct{}),
		rotateFunc: rotateFunc,
	}

	rs.wg.Add(1)
	gzutil.SafeGo(func() {
		rs.scheduleRotation()
	})

	return rs
}

// scheduleRotation 调度日志轮转
func (rs *rotationScheduler) scheduleRotation() {
	defer rs.wg.Done()

	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		duration := next.Sub(now)

		select {
		case <-time.After(duration):
			if rs.rotateFunc != nil {
				rs.rotateFunc()
			}
		case <-rs.done:
			return
		}
	}
}

// Stop 停止调度器
func (rs *rotationScheduler) Stop() {
	close(rs.done)
	rs.wg.Wait()
}

type customWrite struct {
	filepath string
	logger   *lumberjack.Logger
	inner    zapcore.WriteSyncer
	done     chan struct{}
	once     sync.Once
}

// newCustomWrite 创建自定义写入器
func newCustomWrite(filepath string, maxSize, maxBackups, maxAge int, compress bool) *customWrite {
	cw := &customWrite{
		filepath: filepath,
		done:     make(chan struct{}),
	}
	cw.initLogger(filepath, maxSize, maxBackups, maxAge, compress)

	// 启动文件状态监控
	gzutil.SafeGo(func() {
		cw.monitorFile()
	})

	return cw
}

// initLogger 初始化日志文件和写入器
func (cw *customWrite) initLogger(filepath string, maxSize, maxBackups, maxAge int, compress bool) {
	cw.logger = &lumberjack.Logger{
		Filename:   filepath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	}

	cw.inner = zapcore.AddSync(cw.logger)
}

// Write 写入日志
func (cw *customWrite) Write(p []byte) (n int, err error) {
	n, err = cw.inner.Write(p)
	if err != nil {
		cw.recreateLogger()
	}
	return n, err
}

// recreateLogger 重新创建日志文件和写入器
func (cw *customWrite) recreateLogger() {
	cw.once.Do(func() {
		cw.logger.Close()
		cw.initLogger(cw.filepath, cw.logger.MaxSize, cw.logger.MaxBackups, cw.logger.MaxAge, cw.logger.Compress)
	})
}

// monitorFile 异步监控日志文件状态
func (cw *customWrite) monitorFile() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if _, err := os.Stat(cw.filepath); os.IsNotExist(err) {
				cw.recreateLogger()
			}
		case <-cw.done:
			return
		}
	}
}

// Sync 刷新日志到文件
func (cw *customWrite) Sync() error {
	return cw.inner.Sync()
}

// Close 关闭写入器
func (cw *customWrite) Close() error {
	close(cw.done)

	return cw.Sync()
}

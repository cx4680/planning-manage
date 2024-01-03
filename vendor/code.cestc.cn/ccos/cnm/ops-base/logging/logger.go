package logging

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"code.cestc.cn/ccos/cnm/ops-base/rolling"
	"code.cestc.cn/ccos/cnm/ops-base/trace"
)

const (
	LevelEnv     = "LOGGING_LEVEL"
	defaultLevel = "info"
	gb           = 0x40100000
)

// Logger...
type Logger struct {
	*zap.SugaredLogger
	path         string
	dir          string
	rolling      rolling.RollingFormat
	rollingFiles []io.Writer
	loglevel     zap.AtomicLevel
	prefix       string
	encoderCfg   zapcore.EncoderConfig
	callSkip     int
}

//		FunctionKey:      "func",
//		StacktraceKey:    "stack",
//		NameKey:          "name",
//		MessageKey:       "msg",
//		LevelKey:         "level",
//		ConsoleSeparator: " | ",
//		EncodeLevel:      EncodeLevel,
//		TimeKey:          "s",
//		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
//			enc.AppendString(t.Format("2006/01/02 15:04:05"))
//		},
//		CallerKey:    "file",
//		EncodeCaller: zapcore.ShortCallerEncoder,
//		EncodeName: func(n string, enc zapcore.PrimitiveArrayEncoder) {
//			enc.AppendString(n)
//		},

var defaultEncoderConfig = zapcore.EncoderConfig{
	CallerKey:        "caller",
	StacktraceKey:    "stack",
	LineEnding:       zapcore.DefaultLineEnding,
	TimeKey:          "time",
	MessageKey:       "msg",
	LevelKey:         "level",
	NameKey:          "logger",
	EncodeCaller:     zapcore.ShortCallerEncoder,
	EncodeLevel:      zapcore.CapitalColorLevelEncoder,
	EncodeTime:       MilliSecondTimeEncoder,
	EncodeDuration:   zapcore.StringDurationEncoder,
	EncodeName:       zapcore.FullNameEncoder,
	ConsoleSeparator: " | ",
}

var (
	_defaultLogger  *Logger
	_jsonDataLogger *Logger
	_operatorLog    *Logger
	_config         LogConfig
)

const (
	_jsonDataTaskKey = "service_name"
	localIP          = "local_ip"
	uniqID           = "uniq_id"
)

// Logger name for default loggers
const (
	DefaultLoggerName = "_default"
	SlowLoggerName    = "_slow"
	GenLoggerName     = "_gen"
	CrashLoggerName   = "_crash"
	BalanceLoggerName = "_balance"
)

var (
	defaultStorageDay int64 = 7
	defaultMaxSize    int64 = 1
	defaultDir              = "./logs"
)

func init() {
	_defaultLogger = New()
	logs[DefaultLoggerName] = _defaultLogger
	logs[GenLoggerName] = genLog
	logs[CrashLoggerName] = crashLog

	// 初始化
	_defaultLogger.SetRotateByDay()

	// 默认为info等级
	_defaultLogger.SetLevelByString(defaultLevel)

	// 操作日志
	_operatorLog = NewOperator()

	// 其它日志
	setKitLog("")
}

func SetConfig(cfg LogConfig) {
	_config = cfg
}

var logs = map[string]*Logger{}
var logsMtx sync.RWMutex

func Log(name string) *Logger {
	logsMtx.RLock()
	defer logsMtx.RUnlock()
	return logs[name]
}

func New() *Logger {
	cfg := defaultEncoderConfig
	lvl := zap.NewAtomicLevelAt(zap.DebugLevel)
	debugFile := rolling.NewRollingStd()
	return &Logger{
		SugaredLogger: zap.New(zapcore.NewCore(NewConsoleEncoder(&cfg), debugFile, lvl)).WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Sugar(),
		path:          "",
		dir:           "",
		rolling:       rolling.DailyRolling,
		rollingFiles:  nil,
		loglevel:      lvl,
		prefix:        "",
		encoderCfg:    cfg,
	}
}

// NewJSON build json data format logger
func NewJSON(path string, r rolling.RollingFormat) (*Logger, error) {
	cfg := defaultEncoderConfig
	cfg.LevelKey = ""
	cfg.MessageKey = "topic"
	lvl := zap.NewAtomicLevelAt(zap.DebugLevel)
	rollFile, err := rolling.NewRollingFile(path, r)
	if err != nil {
		return nil, err
	}
	return &Logger{
		SugaredLogger: zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(cfg), rollFile, lvl)).WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Sugar(),
		path:          path,
		dir:           "",
		rolling:       rolling.DailyRolling,
		rollingFiles:  []io.Writer{rollFile},
		loglevel:      lvl,
		prefix:        "",
		encoderCfg:    cfg,
	}, nil
}

func NewOperator() *Logger {
	cfg := getEncoderConfig()
	lvl := zap.NewAtomicLevelAt(zap.DebugLevel)

	debugFile := rolling.NewRollingStd()

	return &Logger{
		SugaredLogger: zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), debugFile, lvl)).Sugar(),
		path:          "",
		dir:           "",
		rolling:       rolling.DailyRolling,
		rollingFiles:  nil,
		loglevel:      lvl,
		prefix:        "",
		encoderCfg:    cfg,
	}
}

// InitData logger
func InitData(path string, rolling rolling.RollingFormat) error {
	if _jsonDataLogger != nil {
		return nil
	}
	l, err := NewJSON(path, rolling)
	if err != nil {
		return err
	}
	_jsonDataLogger = l
	return nil
}

// InitData logger
func InitDataWithKey(path string, rolling rolling.RollingFormat, task string) error {
	err := InitData(path, rolling)
	if err != nil {
		return err
	}

	_jsonDataLogger.SugaredLogger = _jsonDataLogger.SugaredLogger.With(_jsonDataTaskKey, task)
	return nil
}

func (l *Logger) SetOutput(out io.Writer) {
	l.SugaredLogger = zap.New(zapcore.NewCore(NewConsoleEncoder(&l.encoderCfg), zapcore.Lock(zapcore.AddSync(out)), zap.DebugLevel)).WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
	l.SugaredLogger.Named(l.prefix)
}

func (l *Logger) GetOutput() io.Writer {
	return nil
}

func (l *Logger) SetColors(color bool) {
	if !color {
		l.encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	} else {
		l.encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
}

func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{SugaredLogger: l.SugaredLogger.With(args...).Desugar().WithOptions(zap.AddCallerSkip(0)).Sugar()}
}

func (l *Logger) WithService(serviceName string) *Logger {
	_config.ServiceName = serviceName
	return l
}

func (l *Logger) WithFile(dir string) *Logger {
	_config.LogPath = dir

	// 默认保留7天的文件
	go RemoveExpireLogs(defaultDir, defaultStorageDay)

	// 默认保留5g的文件
	go RemoveStrongLogs(defaultDir, defaultMaxSize)

	_ = l.SetOutputPath(dir)

	setKitLog(dir)
	crashLog.SetOutputByName(filepath.Join(dir, "crash.log"), OtherLevelCrash)
	return l
}

func (l *Logger) StorageDay(day int64) *Logger {
	defaultStorageDay = day
	return l
}

func (l *Logger) MaxSize(size int64) *Logger {
	defaultMaxSize = size
	return l
}

func (l *Logger) SetLogPrefix(prefix string) {
	l.prefix = prefix
	l.SugaredLogger.Named(prefix)
}

func (l *Logger) SetRotateByDay() {
	l.rolling = rolling.DailyRolling
	l.refreshRotate()
}

func (l *Logger) refreshRotate() {
	for _, w := range l.rollingFiles {
		r, ok := w.(*rolling.RollingFile)
		if ok {
			r.SetRolling(l.rolling)
		}
	}
}

func (l *Logger) SetRotateByHour() {
	l.rolling = rolling.HourlyRolling
	l.refreshRotate()
}

func (l *Logger) SetRotateBySecond() {
	l.rolling = rolling.SecondlyRolling
	l.refreshRotate()
}

func (l *Logger) SetFlags(flags int) {
	if flags == 0 {
		l.encoderCfg = zapcore.EncoderConfig{
			CallerKey:     "",
			StacktraceKey: "",
			LineEnding:    zapcore.DefaultLineEnding,
			TimeKey:       "",
			MessageKey:    "msg",
			LevelKey:      "",
			NameKey:       "",
		}
	}
}

func (l *Logger) SetHighlighting(highlighting bool) {
	l.SetColors(highlighting)
}

func (l *Logger) SetPrintLevel(printLevel bool) {
	if !printLevel {
		l.encoderCfg.LevelKey = ""
	} else {
		l.encoderCfg.LevelKey = "level"
	}
}

// SetPrintTime 设置是否打印时间
func (l *Logger) SetPrintTime(printTime bool) {
	if !printTime {
		l.encoderCfg.TimeKey = ""
	} else {
		l.encoderCfg.TimeKey = "time"
	}
}

// SetTimeFmt 设置打印时间格式
func (l *Logger) SetTimeFmt() error {
	l.encoderCfg.EncodeTime = NewTimeEncoder()
	return nil
}

// SetPrintCaller 设置是否打印路径
func (l *Logger) SetPrintCaller(printCaller bool) {
	if !printCaller {
		l.encoderCfg.CallerKey = ""
	} else {
		l.encoderCfg.CallerKey = "caller"
	}
}

func (l *Logger) SetOutputByName(path, otherLevel string) error {
	if l.path == path {
		return nil
	}
	l.closeFiles()
	l.path = path
	l.dir = ""
	l.encoderCfg.CallerKey = ""
	debugFile, err := rolling.NewRollingFile(path, l.rolling)
	if err != nil {
		return err
	}
	core := zapcore.NewTee(
		zapcore.NewCore(NewConsoleEncoderWithOtherLevel(&l.encoderCfg, otherLevel), debugFile, l.loglevel),
	)
	l.rollingFiles = []io.Writer{debugFile}
	l.SugaredLogger = zap.New(core).WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
	l.SugaredLogger.Named(l.prefix)
	return nil
}

func (l *Logger) SetOutputNotWithFile(otherLevel string) error {
	core := zapcore.NewTee(
		zapcore.NewCore(NewConsoleEncoderWithOtherLevel(&l.encoderCfg, otherLevel), rolling.NewRollingStd(), l.loglevel),
	)
	l.SugaredLogger = zap.New(core).WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
	l.SugaredLogger.Named(l.prefix)
	return nil
}

func (l *Logger) closeFiles() {
	for _, w := range l.rollingFiles {
		r, ok := w.(*rolling.RollingFile)
		if ok {
			r.Close()
		}
	}
	l.rollingFiles = nil
}

func (l *Logger) SetOutputPath(path string) error {
	if l.dir == path {
		return nil
	}
	l.closeFiles()
	l.path = ""
	l.dir = path
	debugFile, err := rolling.NewRollingFile(path+"/debug.log", l.rolling)
	if err != nil {
		return err
	}
	infoFile, err := rolling.NewRollingFile(path+"/info.log", l.rolling)
	if err != nil {
		return err
	}
	errorFile, err := rolling.NewRollingFile(path+"/error.log", l.rolling)
	if err != nil {
		return err
	}
	debugLogEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		if l.loglevel.Level() > zapcore.DebugLevel {
			return false
		}
		return l.loglevel.Level() == lvl
	})
	errorLogEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})
	infoLogEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return l.loglevel.Level() <= zapcore.InfoLevel && zapcore.InfoLevel == lvl
	})
	core := zapcore.NewTee(
		zapcore.NewCore(NewConsoleEncoder(&l.encoderCfg), debugFile, debugLogEnabler),
		zapcore.NewCore(NewConsoleEncoder(&l.encoderCfg), infoFile, infoLogEnabler),
		zapcore.NewCore(NewConsoleEncoder(&l.encoderCfg), errorFile, errorLogEnabler),
	)
	l.rollingFiles = []io.Writer{debugFile, infoFile, errorFile}
	l.SugaredLogger = zap.New(core).WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
	l.SugaredLogger.Named(l.prefix)
	return nil
}

func (l *Logger) SetLevel(level int) {
	l.loglevel.SetLevel(zapcore.Level(level))
}

func (l *Logger) SetLevelByString(level string) *Logger {
	l.loglevel.SetLevel(stringToLogLevel(level))
	return l
}

func (l *Logger) Logger() *log.Logger {
	stdLogger := log.New(logWriter{
		logFunc: func() func(msg string, fileds ...interface{}) {
			// fmt.Printf("logFunc %v.\n", l)
			logger := &Logger{SugaredLogger: l.SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(3)).Sugar()}
			return logger.Debugf
		},
	}, "", 0)
	return stdLogger
}

func GetLogger() *log.Logger {
	return _defaultLogger.Logger()
}

func SetRotateByHour() {
	_defaultLogger.SetRotateByHour()
}

func SetRotateByDay() {
	_defaultLogger.SetRotateByDay()
}

func SetLevelByString(level string) *Logger {
	_defaultLogger.SetLevelByString(level)
	return _defaultLogger
}

func SetOutputPath(dir string) {
	_defaultLogger.SetOutputPath(dir)
}

func Debug(v ...interface{}) {
	_defaultLogger.Debug(v...)
}

func Info(v ...interface{}) {
	_defaultLogger.Info(v...)
}

func Warn(v ...interface{}) {
	_defaultLogger.Warn(v...)
}

func Warning(v ...interface{}) {
	_defaultLogger.Warn(v...)
}

func Error(v ...interface{}) {
	_defaultLogger.Error(v...)
}

func Fatal(v ...interface{}) {
	_defaultLogger.Fatal(v...)
}

func Debugf(format string, v ...interface{}) {
	_defaultLogger.Debugf(format, v...)
}

func Infof(format string, v ...interface{}) {
	_defaultLogger.Infof(format, v...)
}

func Warnf(format string, v ...interface{}) {
	_defaultLogger.Warnf(format, v...)
}

func Warningf(format string, v ...interface{}) {
	_defaultLogger.Warnf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	_defaultLogger.Errorf(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	_defaultLogger.Fatalf(format, v...)
}

func With(args ...interface{}) *Logger {
	return &Logger{SugaredLogger: _defaultLogger.SugaredLogger.With(args...).Desugar().WithOptions(zap.AddCallerSkip(-1)).Sugar()}
}

func For(ctx context.Context, args ...interface{}) *Logger {
	tid := trace.ExtraTraceID(ctx)
	var fields []interface{}
	if len(tid) != 0 {
		fields = make([]interface{}, 0, len(args)+2)
		fields = append(fields, traceIDKey, tid)
	} else {
		fields = make([]interface{}, 0, len(args))
	}
	fields = append(fields, args...)
	return &Logger{SugaredLogger: _defaultLogger.With(fields...).Desugar().WithOptions(zap.AddCallerSkip(-1)).Sugar()}
}

func WithService(serviceName string) *Logger {
	_config.ServiceName = serviceName
	return _defaultLogger
}

func WithFile(dir string) *Logger {
	_defaultLogger.WithFile(dir)
	setKitLog(dir)
	crashLog.SetOutputByName(filepath.Join(dir, "crash.log"), OtherLevelCrash)
	return _defaultLogger
}

func StorageDay(day int64) *Logger {
	defaultStorageDay = day
	return _defaultLogger
}

func MaxSize(size int64) *Logger {
	defaultMaxSize = size
	return _defaultLogger
}

func Debugw(msg string, keysAndValues ...interface{}) {
	_defaultLogger.Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	_defaultLogger.Infow(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	_defaultLogger.Errorw(msg, keysAndValues...)
}

func Warningw(msg string, keysAndValues ...interface{}) {
	_defaultLogger.Warnw(msg, keysAndValues...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	_defaultLogger.Warnw(msg, keysAndValues...)
}

func stringToLogLevel(level string) zapcore.Level {
	switch level {
	case "fatal":
		return zap.FatalLevel
	case "error":
		return zap.ErrorLevel
	case "warn":
		return zap.WarnLevel
	case "warning":
		return zap.WarnLevel
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	}
	return zap.DebugLevel
}

func normalizeLoggerWithOption(res *Logger, opt *Options) {
	if opt.DisableColors {
		res.SetColors(false)
	}
	if opt.DisableLevel {
		res.encoderCfg.LevelKey = ""
	}
	if opt.DisableFullTimestamp {
		res.encoderCfg.TimeKey = ""
	}
	if opt.Level != "" {
		res.SetLevelByString(opt.Level)
	}
	if opt.Rolling != "" {
		res.rolling = rolling.RollingFormat(opt.Rolling)
	}
}

type logWriter struct {
	logFunc func() func(msg string, fileds ...interface{})
}

func (l logWriter) Write(p []byte) (int, error) {
	p = bytes.TrimSpace(p)
	if l.logFunc != nil {
		l.logFunc()(string(p))
	}
	return len(p), nil
}

func getLogWriter(file string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   file,
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// RemoveExpireLogs 删除超过过期时间的文件
func RemoveExpireLogs(path string, storageDay int64) {
	if storageDay <= 0 {
		storageDay = defaultStorageDay
	}
	t := time.NewTicker(1 * time.Hour)
	expireDayTime := storageDay * 86400
	defer t.Stop()
	for {
		// 获取文件夹下所有文件
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			fileName := fmt.Sprintf("%s/%s", path, f.Name())
			fi, err := getFileStat(fileName)
			if err != nil {
				_defaultLogger.Errorw("f.Stat() error", zap.Error(err))
				continue
			}
			if fi.ModTime().Unix() > time.Now().Unix()-expireDayTime {
				continue
			}
			// 删除文件
			err = os.Remove(fileName)
			if err != nil {
				_defaultLogger.Errorw("file remove error", zap.Error(err))
				continue
			}
		}

		<-t.C
	}
}

func getFileStat(fileName string) (os.FileInfo, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = f.Close()
	}()

	fi, err := f.Stat()
	return fi, err
}

type sortFile struct {
	filename string
	createAt int64
}

type sortFileList []sortFile

func (r sortFileList) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

func (r sortFileList) Len() int { return len(r) }

func (r sortFileList) Less(i, j int) bool { return r[i].createAt < r[j].createAt }

// RemoveStrongLogs 删除超过最大size的文件
func RemoveStrongLogs(path string, maxSize int64) {
	if maxSize <= 0 {
		maxSize = defaultMaxSize
	}
	t := time.NewTicker(1 * time.Hour)
	defer t.Stop()
	for {
		// 获取文件夹下所有文件
		files, _ := ioutil.ReadDir(path)
		var totalSize int64 = 0 // 总大小
		var sortList sortFileList
		for _, f := range files {
			fileName := fmt.Sprintf("%s/%s", path, f.Name())
			fi, err := getFileStat(fileName)
			if err != nil {
				_defaultLogger.Errorw("f.Stat() error", zap.Error(err))
				continue
			}
			sortList = append(sortList, sortFile{
				filename: fileName,
				createAt: fi.ModTime().Unix(),
			})
			totalSize += fi.Size()
		}
		if totalSize >= maxSize*gb {
			sort.Sort(sortList)
			// 删除文件
			err := os.Remove(sortList[0].filename)
			if err != nil {
				_defaultLogger.Errorw("file remove error", zap.Error(err))
			}
		}
		<-t.C
	}
}

func getEncoderConfig() zapcore.EncoderConfig {
	//EncodeLevel := zapcore.CapitalColorLevelEncoder
	return zapcore.EncoderConfig{
		FunctionKey:      "func",
		StacktraceKey:    "stack",
		NameKey:          "name",
		MessageKey:       "msg",
		LevelKey:         "level",
		ConsoleSeparator: " | ",
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		TimeKey:          "s",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006/01/02 15:04:05"))
		},
		CallerKey:    "file",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeName: func(n string, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(n)
		},
	}
}

func WithOpFields(args ...interface{}) *Logger {
	return _operatorLog.With(args...)
}

func (l *Logger) OpLogPrint(key string) {
	l.Infow(key)
}

func OpLogPrint(msg string, keysAndValues ...interface{}) {
	_operatorLog.Infow(msg, keysAndValues...)
}

//func rollEditLevel() {
//	t := time.NewTicker(1 * time.Second)
//	defer t.Stop()
//
//	for {
//		<-t.C
//		level := os.Getenv(LevelEnv)
//		if len(level) <= 0 {
//			level = defaultLevel
//		}
//		_defaultLogger.SetLevelByString(level)
//	}
//}

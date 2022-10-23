package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/sirupsen/logrus"
)

var (
	e             *logrus.Entry
	jsonFormatter *logrus.JSONFormatter
	textFormatter *logrus.TextFormatter
)

type (
	writerHook struct {
		Writer    []io.Writer
		LogLevels []logrus.Level
	}
	logger struct {
		*logrus.Entry
	}
	Logger interface {
		WithFields(args map[string]interface{}) Logger
		WithField(key string, value interface{}) Logger

		Log(level logrus.Level, args ...interface{})

		// Trace log methods
		// Trace f
		Trace(args ...interface{})
		// Debug ...
		Debug(args ...interface{})
		// Info ...
		Info(args ...interface{})
		// Warn ...
		Warn(args ...interface{})
		// Error ...
		Error(args ...interface{})
		// Fatal ...
		Fatal(args ...interface{})
		// Panic ...
		Panic(args ...interface{})

		// Tracef f log methods
		Tracef(format string, args ...interface{})
		// Debugf ...
		Debugf(format string, args ...interface{})
		// Infof ...
		Infof(format string, args ...interface{})
		// Warnf ...
		Warnf(format string, args ...interface{})
		// Errorf ...
		Errorf(format string, args ...interface{})
		// Fatalf ...
		Fatalf(format string, args ...interface{})
		// Panicf ...
		Panicf(format string, args ...interface{})
		GetLevel() uint32
		GetEntry() *logrus.Entry
	}
)

// init
func init() {
	var config struct {
		LogLevel string `env:"LOG_LEVEL" envDefault:"trace"`
	}
	l := logrus.New()

	l.SetReportCaller(true)

	textFormatter = &logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
		CallerPrettyfier: func(f *runtime.Frame) (fun string, file string) {
			filename := path.Base(f.File)
			return f.Function, fmt.Sprintf("%s:%d", filename, f.Line)
		},
		TimestampFormat: time.RFC3339,
	}

	jsonFormatter = &logrus.JSONFormatter{
		DisableTimestamp: false,
		TimestampFormat:  time.RFC3339,
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			filename := path.Base(f.File)
			return f.Function, fmt.Sprintf("%s:%d", filename, f.Line)
		},
		PrettyPrint: false,
	}

	if format := os.Getenv("LOG_FORMATTER"); format == "text" {
		l.SetFormatter(textFormatter)
	} else {
		l.SetFormatter(jsonFormatter)
	}

	if err := os.Mkdir("logs", 0777); err != nil && !errors.Is(err, os.ErrExist) {
		l.Panicf("mkdir: %v", err)
	}

	allFile, err := os.OpenFile(fmt.Sprintf("logs/%s.log", time.Now().Format("2006-1")), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil && !errors.Is(err, os.ErrExist) {
		l.Panicf("open file: %v", err)
	}

	l.SetOutput(io.Discard)
	l.AddHook(&writerHook{
		Writer:    []io.Writer{allFile, os.Stdout},
		LogLevels: logrus.AllLevels,
	})
	if err := env.Parse(&config); err != nil {
		l.Panicf("env parse: %v", err)
	}
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.TraceLevel
		l.Warnf("parse level: %v", err)
	}
	l.SetLevel(level)

	e = logrus.NewEntry(l)
}

// Fire ...
func (h *writerHook) Fire(e *logrus.Entry) error {
	line, err := e.String()
	if err != nil {
		return fmt.Errorf("entry: String(); %v", err)
	}
	for _, w := range h.Writer {
		_, _ = w.Write([]byte(line))
	}
	return nil
}

// Levels ...
func (h *writerHook) Levels() []logrus.Level {
	return h.LogLevels
}

// GetLogger ...
func GetLogger() Logger {
	return &logger{e}
}

// DeleteLogFolderAndFile ...
func DeleteLogFolderAndFile(t *testing.T) {
	t.Helper()
	_ = os.RemoveAll("logs")
}

// WithFields ...
func (l *logger) WithFields(args map[string]interface{}) Logger {
	return &logger{e.WithFields(args)}
}

// WithField ...
func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{e.WithField(key, value)}
}

// GetEntry
func (l *logger) GetEntry() *logrus.Entry {
	return l.Entry
}

func (l *logger) GetLevel() uint32 {
	return uint32(l.Entry.Level)
}

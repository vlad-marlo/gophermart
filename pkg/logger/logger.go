package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

var e *logrus.Entry

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

		// Trace log methods
		Trace(args ...interface{})
		Debug(args ...interface{})
		Info(args ...interface{})
		Warn(args ...interface{})
		Error(args ...interface{})
		Fatal(args ...interface{})
		Panic(args ...interface{})

		// Tracef f log methods
		Tracef(format string, args ...interface{})
		Debugf(format string, args ...interface{})
		Infof(format string, args ...interface{})
		Warnf(format string, args ...interface{})
		Errorf(format string, args ...interface{})
		Panicf(format string, args ...interface{})
	}
)

// init
func init() {
	l := logrus.New()

	l.SetReportCaller(true)

	textFormatter := &logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: true,
		CallerPrettyfier: func(f *runtime.Frame) (fun string, file string) {
			filename := path.Base(f.File)
			return f.Function, fmt.Sprintf("%s:%d", filename, f.Line)
		},
		TimestampFormat: time.RFC3339,
	}

	jsonFormatter := &logrus.JSONFormatter{
		DisableTimestamp: false,
		TimestampFormat:  time.RFC3339,
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			filename := path.Base(f.File)
			return f.Function, fmt.Sprintf("%s:%d", filename, f.Line)
		},
		PrettyPrint: false,
	}

	format := os.Getenv("LOG_FORMATTER")
	if format == "text" {
		l.SetFormatter(textFormatter)
	} else {
		l.SetFormatter(jsonFormatter)
	}

	if err := os.Mkdir("logs", 0777); err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}
	allFile, err := os.OpenFile(fmt.Sprintf("logs/%s.log", time.Now().Format("2006-1")), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}

	l.SetOutput(io.Discard)
	l.AddHook(&writerHook{
		Writer:    []io.Writer{allFile, os.Stdout},
		LogLevels: logrus.AllLevels,
	})
	l.SetLevel(logrus.TraceLevel)

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
func DeleteLogFolderAndFile() {
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

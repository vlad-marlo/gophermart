package logger

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"runtime"
)

var e *logrus.Entry

type (
	writerHook struct {
		Writer    []io.Writer
		LogLevels []logrus.Level
	}
	Logger struct {
		*logrus.Entry
	}
)

func init() {
	l := logrus.New()

	l.SetReportCaller(true)
	l.Formatter = &logrus.JSONFormatter{
		PrettyPrint: false,
		CallerPrettyfier: func(f *runtime.Frame) (fun string, file string) {
			filename := path.Base(f.File)
			return f.Function, fmt.Sprintf("%s:%d", filename, f.Line)
		},
		DisableTimestamp: false,
	}

	if err := os.Mkdir("logs", 0777); err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}
	allFile, err := os.OpenFile("logs/all.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil && err != os.ErrExist {
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

func (h *writerHook) Levels() []logrus.Level {
	return h.LogLevels
}

func GetLogger() Logger {
	return Logger{e}
}
func (l *Logger) GetLoggerWithWithField(k string, v interface{}) Logger {
	return Logger{l.WithField(k, v)}
}

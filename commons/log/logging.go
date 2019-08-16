package log

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
	"time"
)

type OrtooLog struct {
	*logrus.Logger
}

var Logger = NewOrtooLog()

func NewOrtooLog() *OrtooLog {
	log := logrus.New()
	log.SetFormatter(&ortooFormatter{})
	log.SetReportCaller(true)
	return &OrtooLog{log}

}

const (
	colorRed    = 31
	colorYellow = 33
	colorBlue   = 36
	colorGray   = 37
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = strings.Replace(b, "/commons/log/logging.go", "/", 1)
)

const (
	fieldErrorAt = "errorAt"
	fieldError   = "error"
)

func getColorByLevel(level logrus.Level) int {
	switch level {
	case logrus.DebugLevel:
		return colorGray
	case logrus.WarnLevel:
		return colorYellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return colorRed
	default:
		return colorBlue
	}
}

type ortooFormatter struct {
}

func (o *ortooFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestampFormat := time.StampMilli
	b := &bytes.Buffer{}
	b.WriteString(entry.Time.Format(timestampFormat))
	level := strings.ToUpper(entry.Level.String())

	_, _ = fmt.Fprintf(b, "\x1b[%dm", getColorByLevel(entry.Level))

	b.WriteString(" [" + level[:4] + "]")
	b.WriteString("\x1b[0m")

	if entry.Data[fieldErrorAt] != nil {
		b.WriteString("[" + entry.Data[fieldError].(string) + "]")
		b.WriteString("[ " + entry.Data[fieldErrorAt].(string) + " ] ")
	} else {
		relativeCallFile := strings.Replace(entry.Caller.File, basepath, "", 1)
		fileLine := fmt.Sprintf("[ %s:%d ] ", relativeCallFile, entry.Caller.Line)
		b.WriteString(fileLine)
	}
	b.WriteString(entry.Message)
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (o *OrtooLog) OrtooError(err error, format string, args ...interface{}) error {
	_, file, line, _ := runtime.Caller(2)
	relativeCallFile := strings.Replace(file, basepath, "", 1)
	errorPlace := fmt.Sprintf("%s:%d", relativeCallFile, line)
	var errString = "nil"
	if err != nil {
		errString = err.Error()
	}
	o.WithFields(logrus.Fields{
		fieldErrorAt: errorPlace,
		fieldError:   errString,
	}).Errorf(format, args...)
	return err
}

func OrtooError(err error, format string, args ...interface{}) error {
	return Logger.OrtooError(err, format, args...)
}

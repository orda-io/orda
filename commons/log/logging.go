package log

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
	"time"
)

// OrtooLog defines the log of OrtooLog
type OrtooLog struct {
	*logrus.Entry
}

// Logger is a global instance of OrtooLog
var Logger = NewOrtooLog()

// NewOrtooLog creates a new OrtooLog.
func NewOrtooLog() *OrtooLog {
	logger := logrus.New()
	logger.SetFormatter(&ortooFormatter{})
	logger.SetReportCaller(true)
	return &OrtooLog{logrus.NewEntry(logger)}
}

// NewOrtooLogWithTag creates a new OrtooLog with a tag.
func NewOrtooLogWithTag(tag string) *OrtooLog {
	return &OrtooLog{NewOrtooLog().WithFields(logrus.Fields{tagField: tag})}
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
	tagField     = "tagFiled"
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

// Format implements format of the OrtooLog.
func (o *ortooFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestampFormat := time.StampMilli
	b := &bytes.Buffer{}
	b.WriteString(entry.Time.Format(timestampFormat))
	level := strings.ToUpper(entry.Level.String())

	_, _ = fmt.Fprintf(b, "\x1b[%dm", getColorByLevel(entry.Level))

	b.WriteString(" [" + level[:4] + "]")
	b.WriteString("\x1b[0m ")

	if entry.Data[tagField] != nil {
		b.WriteString("[" + entry.Data[tagField].(string) + "] ")
	} else if strings.Contains(entry.Caller.File, "server/") {
		b.WriteString("[SERVER] ")
	} else if strings.Contains(entry.Caller.File, "commons/") {
		b.WriteString("[SDK] ")
	}

	b.WriteString(entry.Message)

	if entry.Data[fieldErrorAt] != nil {
		b.WriteString("\t\t")
		if entry.Data[fieldError] != nil {
			b.WriteString("[" + entry.Data[fieldError].(string) + "]")
		}
		b.WriteString("[ " + entry.Data[fieldErrorAt].(string) + " ] ")
	} else {
		relativeCallFile := strings.Replace(entry.Caller.File, basepath, "", 1)
		_, _ = fmt.Fprintf(b, "\t\t[ %s:%d ] ", relativeCallFile, entry.Caller.Line)
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

// OrtooErrorf is a method to present a error log.
func (o *OrtooLog) OrtooErrorf(err error, format string, args ...interface{}) error {
	return o.ortooErrorf(err, 3, format, args...)
}

func (o *OrtooLog) ortooErrorf(err error, skip int, format string, args ...interface{}) error {
	_, file, line, _ := runtime.Caller(skip)
	relativeCallFile := strings.Replace(file, basepath, "", 1)
	errorPlace := fmt.Sprintf("%s:%d", relativeCallFile, line)
	var errString = "nil"
	if err != nil {
		errString = err.Error()
	} else {
		err = fmt.Errorf("nil")
	}
	o.WithFields(logrus.Fields{
		fieldErrorAt: errorPlace,
		fieldError:   errString,
	}).Errorf(format, args...)
	return err
}

// OrtooErrorf is a method wrapping Logger.OrtooErrorf()
func OrtooErrorf(err error, format string, args ...interface{}) error {
	return Logger.ortooErrorf(err, 2, format, args...)
}

func OrtooError(err error) error {
	_, file, line, _ := runtime.Caller(1)
	relativeCallFile := strings.Replace(file, basepath, "", 1)
	errorPlace := fmt.Sprintf("%s:%d", relativeCallFile, line)
	var errString = "nil"
	if err != nil {
		errString = err.Error()
	} else {
		err = fmt.Errorf("nil")
	}

	Logger.WithFields(logrus.Fields{
		fieldErrorAt: errorPlace,
	}).Errorf("%s", errString)

	return err
}

// OrtooErrorWithSkip is a method wrapping Logger.OrtooErrorf()
func OrtooErrorWithSkip(err error, skip int, format string, args ...interface{}) error {
	return Logger.ortooErrorf(err, skip, format, args...)
}

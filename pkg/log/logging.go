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
var Logger = NewWithTags("ORTOO", "DEFAULT")

const (
	colorRed    = 31
	colorYellow = 33
	colorBlue   = 36
	colorGray   = 37
)

var (
	_, b, _, _ = runtime.Caller(0)
	basePath   = strings.Replace(b, "/pkg/log/logging.go", "/", 1)
)

const (
	tag1Field = "1stTag"
	tag2Field = "2ndTag"
)

// New creates a new OrtooLog.
func New() *OrtooLog {
	logger := logrus.New()
	logger.SetFormatter(&ortooFormatter{})
	logger.SetReportCaller(true)
	return &OrtooLog{logrus.NewEntry(logger)}
}

func (its *OrtooLog) GetTag1() string {
	return its.Data[tag1Field].(string)
}

func (its *OrtooLog) GetTag2() string {
	return its.Data[tag2Field].(string)
}

func (its *OrtooLog) SetTags(tag1, tag2 string) {
	its.Data[tag1Field] = tag1
	its.Data[tag2Field] = tag2
}

// NewWithTags creates a new OrtooLog with a tag.
func NewWithTags(tag1, tag2 string) *OrtooLog {
	return &OrtooLog{
		New().WithFields(logrus.Fields{
			tag1Field: tag1,
			tag2Field: tag2,
		}),
	}
}

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

type ortooFormatter struct{}

// Format implements format of the OrtooLog.
func (o *ortooFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	b := &bytes.Buffer{}
	level := strings.ToUpper(entry.Level.String())
	_, _ = fmt.Fprintf(b, "\x1b[%dm", getColorByLevel(entry.Level))

	b.WriteString("[" + level[:4] + "]")
	b.WriteString("\x1b[0m ")
	b.WriteString("[")
	// main level
	if v, ok := entry.Data[tag1Field]; ok {
		b.WriteString(v.(string))
	} else if strings.Contains(entry.Caller.File, "server/") {
		b.WriteString("SERV")
	} else if strings.Contains(entry.Caller.File, "pkg/") {
		b.WriteString("SDKS")
	} else {
		b.WriteString("NONE")
	}
	if v, ok := entry.Data[tag2Field]; ok && v != "" {
		b.WriteString("|")
		b.WriteString(v.(string))
	}
	b.WriteString("] ")

	b.WriteString(entry.Message)
	b.WriteString("\t\t")

	timestampFormat := time.StampMilli
	b.WriteString(" [")
	b.WriteString(entry.Time.Format(timestampFormat))
	b.WriteString("] ")

	relativeCallFile := strings.Replace(entry.Caller.File, basePath, "", 1)
	_, _ = fmt.Fprintf(b, "[ %s:%d ] ", relativeCallFile, entry.Caller.Line)
	b.WriteByte('\n')
	return b.Bytes(), nil
}

package log

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// OrdaLog defines the log of OrdaLog
type OrdaLog struct {
	*logrus.Entry
}

// Logger is a global instance of OrdaLog
var Logger = NewWithTags("🎪", "", "", "", "", "", "")

const (
	colorRed    = 31
	colorYellow = 33
	colorBlue   = 36
	colorGray   = 37
)

var (
	_, b, _, _ = runtime.Caller(0)
	basePath   = strings.Replace(b, "/client/pkg/log/logging.go", "/", 1)
)

const (
	tagEmoji       = "context"
	tagCollection  = "collection"
	tagColNum      = "collectionNum"
	tagClient      = "client"
	tagCuid        = "cuid"
	tagDatatype    = "datatype"
	tagDuid        = "duid"
	tagOrdaContext = "ordaContext"
)

// New creates a new OrdaLog.
func New() *OrdaLog {
	logger := logrus.New()
	jsonFormatter := logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	}
	logFormat := os.Getenv("ORDA_LOG_FORMAT")
	if logFormat == "json" {
		logger.SetFormatter(&jsonFormatter)
	} else {
		logger.SetFormatter(&ordaFormatter{})
	}
	logger.SetReportCaller(true)
	return &OrdaLog{logrus.NewEntry(logger)}
}

func (its *OrdaLog) Clone() *OrdaLog {
	newOne := New()
	for s := range its.Data {
		newOne.Data[s] = its.Data[s]
	}
	return newOne
}

// GetTagEmoji returns context tag
func (its *OrdaLog) GetTagEmoji() string {
	return its.Data[tagEmoji].(string)
}

func (its *OrdaLog) setTag(tag, val string) {
	if val == "" {
		return
	}
	its.Data[tag] = val
}

func (its *OrdaLog) updateTagOrdaContext() {
	maxLen := 12
	var collection, colNum, client, cuid, datatype, duid = "", "", "", "", "", ""
	if s, ok := its.Data[tagCollection]; ok {
		collection = s.(string)
	}
	if s, ok := its.Data[tagColNum]; ok {
		colNum = s.(string)
	}
	if s, ok := its.Data[tagClient]; ok {
		client = s.(string)
	}
	if s, ok := its.Data[tagCuid]; ok {
		cuid = s.(string)
	}
	if s, ok := its.Data[tagDatatype]; ok {
		datatype = s.(string)
	}
	if s, ok := its.Data[tagDuid]; ok {
		duid = s.(string)
	}
	col, cli, dty := "🙈", "🙉", "🙊"
	if collection != "" {
		col = MakeShort(collection, maxLen)
		if colNum != "" {
			col = col + "(" + colNum + ")"
		}
	}
	if client != "" && cuid != "" {
		cli = fmt.Sprintf("%s(%s)", MakeShort(client, maxLen), cuid)
	}
	if datatype != "" && duid != "" {
		dty = fmt.Sprintf("%s(%s)", MakeShort(datatype, maxLen), duid)
	}
	its.Data[tagOrdaContext] = fmt.Sprintf("%s|%s|%s", col, cli, dty)
}

// SetTags sets tags
func (its *OrdaLog) SetTags(emoji, collection, colNum, client, cuid, datatype, duid string) {
	its.setTag(tagEmoji, emoji)
	its.setTag(tagCollection, collection)
	its.setTag(tagColNum, colNum)
	its.setTag(tagClient, client)
	its.setTag(tagCuid, cuid)
	its.setTag(tagDatatype, datatype)
	its.setTag(tagDuid, duid)
	its.updateTagOrdaContext()
}

// NewWithTags creates a new OrdaLog with a tag.
func NewWithTags(context, collection, colNum, client, cuid, datatype, duid string) *OrdaLog {
	l := &OrdaLog{
		New().WithFields(logrus.Fields{}),
	}
	l.SetTags(context, collection, colNum, client, cuid, datatype, duid)
	return l
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

type ordaFormatter struct{}

// Format implements format of the OrdaLog.
func (o *ordaFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	b := &bytes.Buffer{}
	level := strings.ToUpper(entry.Level.String())
	_, _ = fmt.Fprintf(b, "\x1b[%dm", getColorByLevel(entry.Level))

	b.WriteString("[" + level[:4] + "]")
	b.WriteString("\x1b[0m ")
	b.WriteString("[")
	// main level

	if v, ok := entry.Data[tagEmoji]; ok && v != "" {
		b.WriteString(v.(string))
		b.WriteString("|")
	}

	if v, ok := entry.Data[tagOrdaContext]; ok {
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

package utils

import (
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var Log = logrus.New()

func init() {
	formatter := new(prefixed.TextFormatter)
	formatter.DisableColors = false
	formatter.ForceColors = true
	formatter.SetColorScheme(&prefixed.ColorScheme{
		PrefixStyle:    "blue+b",
		TimestampStyle: "white+h",
	})
	Log.Formatter = formatter
	Log.Level = logrus.DebugLevel
	Log.Info("Log Init")
}

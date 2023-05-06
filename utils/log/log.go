package log

import (
	"github.com/astaxie/beego/logs"
)

var Log *logs.BeeLogger

func init() {
	Log = logs.NewLogger(10000)
}

func InitLog(logWay string, logFile string, logLevel string) {
	if logWay == "console" {
		Log.SetLogger("console", "")
	} else {
		Log.SetLogger("file", `{"filename":`+logFile+`"}`)
	}

	level := 4
	switch logLevel {
	case "error":
		level = 3
	case "warn":
		level = 4
	case "info":
		level = 6
	case "debug":
		level = 8
	default:
		level = 4
	}
	Log.SetLevel(level)
}

func Info(format string, v ...interface{}) {
	Log.Info(format, v...)
}
func Warn(format string, v ...interface{}) {
	Log.Warn(format, v...)
}
func Debug(format string, v ...interface{}) {
	Log.Debug(format, v...)
}
func Error(format string, v ...interface{}) {
	Log.Error(format, v...)
}

package main

import (
	"log"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

const (
	INFO = iota + 1
	WARN
)

type Logger struct {
	lgrs *logrus.Logger
}

func NewLogger(logDir string, logFile string) *Logger {
	logPath := filepath.Join(logDir, logFile)
	rotatelog, err := rotatelogs.New(
		logPath+".%Y%m%d",
		rotatelogs.WithLinkName(logPath),
		rotatelogs.WithRotationTime(time.Hour*24),
	)
	if err != nil {
		log.Fatalln(err)
	}
	lg := &Logger{lgrs: logrus.New()}
	lg.lgrs.Formatter = &logrus.JSONFormatter{}
	lg.lgrs.Level = logrus.InfoLevel
	lg.lgrs.Out = rotatelog

	return lg
}

func (*Logger) Log(lv int, msg string, fld logrus.Fields) {
	switch lv {
	case INFO:
		logger.lgrs.WithFields(fld).Info(msg)
	case WARN:
		logger.lgrs.WithFields(fld).Warn(msg)
	}
}

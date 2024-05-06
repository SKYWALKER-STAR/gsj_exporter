package utils

import (
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

var (
	//日志模块
	//日志记录器
	logger *logrus.Logger
)

func GetLogger() *logrus.Logger {
	if logger == nil {
		path := "logs/gaussdb_exporter_%Y%m%d.log"
		writer, _ := rotatelogs.New(
			//路径名称至关重要，rotatelogs通过该路径名匹配来生成新文件名称的，如果粒度太粗则无法生成新文件
			path,
			//日志保留时长7天
			//rotatelogs.WithRotationCount(7),
			rotatelogs.WithMaxAge(24*time.Duration(7)*time.Hour),
			//24小时轮滚一次
			rotatelogs.WithRotationTime(time.Hour*24),
		)
		logger = logrus.New()
		logger.SetOutput(writer)
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05"})
	}
	return logger
}

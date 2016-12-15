package fluent 

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/evalphobia/logrus_fluent"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"time"

	"sync"
)

type LogConfig struct {
	IP      string // fluentip
	Port    int64  // fluentport
	Tag     string // App name
	Level   string // log level
	APPIP   string
	APPPort int64
}

const (
	Panic = "Panic"
	Fatal = "Fatal"
	Error = "Error"
	Warn  = "Warn"
	Info  = "Info"
	Debug = "Debug"
)

var logConfig LogConfig

var logger *logrus.Logger

var once sync.Once
func init() {
	// 读取配置文件
	filename, _ := filepath.Abs("./logConfig.yml.def")
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlFile, &logConfig)
	if err != nil {
		panic(err)
	}

	once.Do( func() {
		if logger == nil {
			logger = NewLog()
		}
	})
}

func NewLog() *logrus.Logger {
	logger := logrus.New()
	hook, err := logrus_fluent.New(logConfig.IP, int(logConfig.Port))
	if err != nil {
		return nil//panic(err)
	}

	switch logConfig.Level {
	case Panic:
		{
			// set custom fire level
			hook.SetLevels([]logrus.Level{
				logrus.PanicLevel,
			})
		}
	case Fatal:
		{
			// set custom fire level
			hook.SetLevels([]logrus.Level{
				logrus.PanicLevel,
				logrus.FatalLevel,
			})
		}
	case Error:
		{
			// set custom fire level
			hook.SetLevels([]logrus.Level{
				logrus.PanicLevel,
				logrus.FatalLevel,
				logrus.ErrorLevel,
			})
		}
	case Warn:
		{
			// set custom fire level
			hook.SetLevels([]logrus.Level{
				logrus.PanicLevel,
				logrus.FatalLevel,
				logrus.ErrorLevel,
				logrus.WarnLevel,
			})
		}
	case Info:
		{
			// set custom fire level
			hook.SetLevels([]logrus.Level{
				logrus.PanicLevel,
				logrus.FatalLevel,
				logrus.ErrorLevel,
				logrus.WarnLevel,
				logrus.InfoLevel,
			})
		}
	case Debug:
		{
			// set custom fire level
			hook.SetLevels([]logrus.Level{
				logrus.PanicLevel,
				logrus.FatalLevel,
				logrus.ErrorLevel,
				logrus.WarnLevel,
				logrus.InfoLevel,
				logrus.DebugLevel,
			})
		}
	}
	// set static tag
	hook.SetTag(logConfig.Tag)
	hook.AddFilter("error", logrus_fluent.FilterError)
	logger.Hooks.Add(hook)

	return logger
}

func Debugs(place string, message string) {
	if logger == nil {
		logger = NewLog()
	}
	if len(message) > 0 {
		logger.Debug(fmt.Sprintf(`level:Debug,ip:%s,port:%v,time:%s,place:%s, value:%s`,
			logConfig.APPIP,
			logConfig.APPPort,
			time.Now().Format("2006-01-02 15:04:05"),
			place,
			message))
		return
	}
	logger.Debug(fmt.Sprintf(`level:Debug,ip:%s,port:%v,time:%s,place:%s`,
		logConfig.APPIP,
		logConfig.APPPort,
		time.Now().Format("2006-01-02 15:04:05"),
		place))
}

func Infos(place string, message string) {
	if logger == nil {
		logger = NewLog()
	}
	if len(message) > 0 {
		logger.Info(fmt.Sprintf(`level:Info,ip:%s,port:%v,time:%s,place:%s, value:%s`,
			logConfig.APPIP,
			logConfig.APPPort,
			time.Now().Format("2006-01-02 15:04:05"),
			place,
			message))
		return
	}
	logger.Info(fmt.Sprintf(`level:Info,ip:%s,port:%v,time:%s,place:%s`,
		logConfig.APPIP,
		logConfig.APPPort,
		time.Now().Format("2006-01-02 15:04:05"),
		place))
}

func Warns(place string, message string) {
	if logger == nil {
		logger = NewLog()
	}
	if len(message) > 0 {
		logger.Warn(fmt.Sprintf(`level:Warn,ip:%s,port:%v,time:%s,place:%s, value:%s`,
			logConfig.APPIP,
			logConfig.APPPort,
			time.Now().Format("2006-01-02 15:04:05"),
			place,
			message))
		return
	}
	logger.Warn(fmt.Sprintf(`level:Warn,ip:%s,port:%v,time:%s,place:%s`,
		logConfig.APPIP,
		logConfig.APPPort,
		time.Now().Format("2006-01-02 15:04:05"),
		place))

}

func Errors(place string, message string) {
	if logger == nil {
		logger = NewLog()
	}
	if len(message) > 0 {
		logger.Error(fmt.Sprintf(`level:Error,ip:%s,port:%v,time:%s,place:%s, value:%s`,
			logConfig.APPIP,
			logConfig.APPPort,
			time.Now().Format("2006-01-02 15:04:05"),
			place,
			message))
		return
	}
	logger.Error(fmt.Sprintf(`level:Error,ip:%s,port:%v,time:%s,place:%s`,
		logConfig.APPIP,
		logConfig.APPPort,
		time.Now().Format("2006-01-02 15:04:05"),
		place))
}

func Fatals(place string, message string) {
	if logger == nil {
		logger = NewLog()
	}
	if len(message) > 0 {
		logger.Fatal(fmt.Sprintf(`level:Fatal,ip:%s,port:%v,time:%s,place:%s, value:%s`,
			logConfig.APPIP,
			logConfig.APPPort,
			time.Now().Format("2006-01-02 15:04:05"),
			place,
			message))
		return
	}
	logger.Fatal(fmt.Sprintf(`level:Fatal,ip:%s,port:%v,time:%s,place:%s`,
		logConfig.APPIP,
		logConfig.APPPort,
		time.Now().Format("2006-01-02 15:04:05"),
		place))
}

func Panics(place string, message string) {
	if logger == nil {
		logger = NewLog()
	}
	if len(message) > 0 {
		logger.Panic(fmt.Sprintf(`level:Panic,ip:%s,port:%v,time:%s,place:%s, value:%s`,
			logConfig.APPIP,
			logConfig.APPPort,
			time.Now().Format("2006-01-02 15:04:05"),
			place,
			message))
		return
	}
	logger.Panic(fmt.Sprintf(`level:Panic,ip:%s,port:%v,time:%s,place:%s`,
		logConfig.APPIP,
		logConfig.APPPort,
		time.Now().Format("2006-01-02 15:04:05"),
		place))
}

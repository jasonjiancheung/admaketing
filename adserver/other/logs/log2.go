package logs

import "fluent"
import "sync"

type LoggerInfo struct {
	Level 	string
	Place 	[]byte
	Content []byte
}

var log  *Logger
var once sync.Once

type Logger struct {
	sync.RWMutex
	InfoChan chan *LoggerInfo
}

func New() *Logger {
	f := func() {
		log = &Logger{}
	}
	once.Do(f) 

	return log
}

func (log *Logger) Debug(place string, content []byte) {
	fluent.Debugs(place,string(content)) 
}

func (log *Logger) Info(place string, content []byte) {
fluent.Infos(place,string(content))
}

func (log *Logger) Warn(place string, content []byte) {
fluent.Warns(place,string(content))
}

func (log *Logger) Error(place string, content []byte) {
fluent.Errors(place,string(content))
}

func (log *Logger) Fatal(place string, content []byte) {
	fluent.Fatals(place,string(content))
}

func (log *Logger) Panic(place string, content []byte) {
	fluent.Panics(place,string(content))
}
